package middleware

import (
	"os"
	"strings"

	"sun-panel/global"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 为所有响应补上基础安全头。
// 面板是自托管场景，这里只做不破坏功能的低风险项：
//   - nosniff：阻止浏览器把上传的文件猜成 HTML/JS 执行
//   - 防 iframe 嵌套（点击劫持）：默认 SAMEORIGIN，仅允许同源嵌入
//   - Referrer-Policy：跨站跳转不泄露面板地址
//   - Permissions-Policy：关闭不需要的硬件权限
//
// 如需允许被指定源（如 NavMate 后台）iframe 内嵌，设置环境变量
// FRAME_ALLOW_ORIGINS（逗号分隔的源列表），例如：
//
//	FRAME_ALLOW_ORIGINS="http://192.168.3.20:8080"
//
// 此时改用 CSP frame-ancestors 白名单：仅允许 'self' + 列出的源嵌入，
// 其余源一概拒绝（比 X-Frame-Options: SAMEORIGIN 更精确、且保留点击劫持防护）。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// 允许被 iframe 嵌入的额外源（环境变量配置，逗号分隔）
		allowOrigins := os.Getenv("FRAME_ALLOW_ORIGINS")
		if strings.TrimSpace(allowOrigins) == "" {
			// 默认：仅允许同源嵌入（原 SAMEORIGIN 行为）
			h.Set("X-Frame-Options", "SAMEORIGIN")
		} else {
			// 白名单模式：CSP frame-ancestors 精确控制，'self' + 配置的源
			// 其余源一概拒绝，比 X-Frame-Options 更精确且保留点击劫持防护
			csp := "frame-ancestors 'self'"
			for _, o := range strings.Split(allowOrigins, ",") {
				o = strings.TrimSpace(o)
				if o != "" {
					csp += " " + o
				}
			}
			h.Set("Content-Security-Policy", csp)
		}

		// 上传目录：强制下载，避免 svg/html 被当作同源页面渲染（存储型 XSS）
		if len(c.Request.URL.Path) > 9 && c.Request.URL.Path[:9] == "/uploads/" {
			h.Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; sandbox")
		}

		c.Next()
		_ = global.Logger // 保留依赖，便于后续按需加日志
	}
}
