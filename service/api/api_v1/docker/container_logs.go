package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/lib/docker"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gin-gonic/gin"
)

// StreamLogs 以 SSE 流式输出容器日志。
//
// 入参（query，便于前端 fetch 读取流）：id(必填)、follow(默认 true)、tail(默认 200)、since(可选)、timestamps(默认 false)。
// 选 SSE 而非 websocket：日志只需单向推送；前端用 fetch+ReadableStream 读取以便携带 token header 复用登录鉴权。
func (a *Container) StreamLogs(c *gin.Context) {
	containerID := c.Query("id")
	follow := c.DefaultQuery("follow", "true") == "true"
	tail := c.DefaultQuery("tail", "200")
	since := c.Query("since")
	timestamps := c.DefaultQuery("timestamps", "false") == "true"

	if containerID == "" {
		apiReturn.ErrorParamFomat(c, "id is empty")
		return
	}

	cli, err := docker.NewDockerClient()
	if err != nil {
		apiReturn.ErrorDockerClientFail(c, err.Error())
		return
	}
	defer cli.Close()

	// 绑定请求生命周期：客户端断开时 Gin 取消 ctx，日志流随之关闭，不残留 goroutine/docker 连接
	ctx := c.Request.Context()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
		Timestamps: timestamps,
	}
	if since != "" {
		options.Since = since
	}

	logReader, err := cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		apiReturn.ErrorDocker(c, err.Error())
		return
	}
	defer logReader.Close()

	// SSE 响应头：必须在写 body 之前设置
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// 关闭反代（nginx 等）缓冲，保证实时
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	// docker 在容器非 TTY 时返回多路复用帧（每帧 8 字节 header），用 stdcopy 解成纯文本；
	// TTY 模式下是原始字节流，直接逐行读。
	src, stop := buildLogSource(ctx, logReader, cli, containerID)
	defer stop()

	bufReader := bufio.NewReaderSize(src, 64*1024)
	c.Stream(func(w io.Writer) bool {
		line, err := bufReader.ReadString('\n')
		trimmed := strings.TrimRight(line, "\r\n")
		if err != nil {
			// EOF 或连接关闭：输出可能存在的最后一行后结束
			if trimmed != "" {
				_, _ = fmt.Fprintf(w, "data: %s\n\n", trimmed)
			}
			return false
		}
		if trimmed == "" {
			return true
		}
		_, _ = fmt.Fprintf(w, "data: %s\n\n", trimmed)
		return true
	})
}

// buildLogSource 返回按行可读的纯文本日志源与停止函数。
// 非 TTY 容器日志为多路复用帧，用 stdcopy 解到 pipe；TTY 直接返回原始流。
// 停止函数在 ctx 取消时关闭底层连接，避免 goroutine 与 docker 连接残留。
func buildLogSource(ctx context.Context, logReader io.ReadCloser, cli *client.Client, containerID string) (io.Reader, func()) {
	isTTY := false
	if inspect, ierr := cli.ContainerInspect(ctx, containerID); ierr == nil && inspect.Config != nil {
		isTTY = inspect.Config.Tty
	}

	stop := func() { _ = logReader.Close() }

	if isTTY {
		return logReader, stop
	}

	pr, pw := io.Pipe()
	copyDone := make(chan struct{})
	go func() {
		defer close(copyDone)
		defer pw.Close()
		// stdout 与 stderr 都写入同一 pipe，按原始顺序合并
		_, _ = stdcopy.StdCopy(pw, pw, logReader)
	}()
	// 客户端断开（ctx 取消）时主动关底层连接，解除 stdcopy 阻塞
	go func() {
		select {
		case <-ctx.Done():
			_ = logReader.Close()
		case <-copyDone:
		}
	}()

	return pr, stop
}
