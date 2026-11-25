package probe

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"sun-panel/global"
)

// certReloader 按需加载/热更新证书：证书文件变化后无需重启面板。
type certReloader struct {
	certFile string
	keyFile  string
	mu       sync.RWMutex
	cert     *tls.Certificate
	modTime  time.Time
}

func (c *certReloader) getCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if fi, err := os.Stat(c.certFile); err == nil {
		c.mu.RLock()
		loaded := c.cert != nil
		c.mu.RUnlock()
		if !loaded || fi.ModTime().After(c.modTime) {
			if cert, err := tls.LoadX509KeyPair(c.certFile, c.keyFile); err == nil {
				c.mu.Lock()
				c.cert = &cert
				c.modTime = fi.ModTime()
				c.mu.Unlock()
			}
		}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cert == nil {
		return nil, fmt.Errorf("certificate not loaded: %s", c.certFile)
	}
	return c.cert, nil
}

// handler 只提供 /probe（及根路径）用于网络位置探测，返回 204。
func handler() http.Handler {
	mux := http.NewServeMux()
	ok := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusNoContent)
	}
	mux.HandleFunc("/probe", ok)
	mux.HandleFunc("/", ok)
	return mux
}

// Start 启动可选的 TLS 探针服务（默认关闭）。
// 配置：[probe] enabled / port / cert_path / key_path
func Start() {
	enabled := strings.EqualFold(strings.TrimSpace(global.Config.GetValueStringOrDefault("probe", "enabled")), "true")
	if !enabled {
		return
	}

	port := global.Config.GetValueIntOrDefault("probe", "port", 8443)
	certPath := strings.TrimSpace(global.Config.GetValueString("probe", "cert_path"))
	keyPath := strings.TrimSpace(global.Config.GetValueString("probe", "key_path"))
	if certPath == "" || keyPath == "" {
		global.Logger.Warnln("[probe] cert_path/key_path 未配置，探针服务未启动")
		return
	}

	reloader := &certReloader{certFile: certPath, keyFile: keyPath}
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler(),
		ReadHeaderTimeout: 5 * time.Second,
		TLSConfig:         &tls.Config{GetCertificate: reloader.getCertificate, MinVersion: tls.VersionTLS12},
	}

	go func() {
		// 探针异常不能影响主面板
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Errorf("[probe] panic: %v", r)
			}
		}()
		global.Logger.Infof("[probe] listening on :%d", port)
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			global.Logger.Errorf("[probe] server stopped: %v", err)
		}
	}()
}
