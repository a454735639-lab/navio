package middleware

import (
	"sun-panel/global"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 为所有响应补上基础安全头。
// 面板是自托管场景，这里只做不破坏功能的低风险项：
//   - nosniff：阻止浏览器把上传的文件猜成 HTML/JS 执行
//   - SAMEORIGIN：阻止被第三方页面 iframe 嵌套（点击劫持）
//   - Referrer-Policy：跨站跳转不泄露面板地址
//   - Permissions-Policy：关闭不需要的硬件权限
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// 上传目录：强制下载，避免 svg/html 被当作同源页面渲染（存储型 XSS）
		if len(c.Request.URL.Path) > 9 && c.Request.URL.Path[:9] == "/uploads/" {
			h.Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; sandbox")
		}

		c.Next()
		_ = global.Logger // 保留依赖，便于后续按需加日志
	}
}
