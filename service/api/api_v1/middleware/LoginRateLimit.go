package middleware

import (
	"sync"
	"time"

	"sun-panel/api/api_v1/common/apiReturn"

	"github.com/gin-gonic/gin"
)

// 登录失败限速：同一 IP 在窗口期内失败次数超限后暂时拒绝，
// 缓解在线暴力破解（面板常暴露在公网域名下）。
type failBucket struct {
	count int
	until time.Time
	last  time.Time
}

var (
	loginFailMu   sync.Mutex
	loginFails    = map[string]*failBucket{}
	maxFailures   = 8               // 窗口内最大失败次数
	failWindow    = 10 * time.Minute // 计数窗口
	blockDuration = 5 * time.Minute  // 超限后的封禁时长
)

// LoginRateLimit 在登录接口前置检查；记录失败在 LoginGuard 中完成
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		loginFailMu.Lock()
		b, ok := loginFails[ip]
		if ok {
			now := time.Now()
			if now.Before(b.until) {
				loginFailMu.Unlock()
				apiReturn.ErrorCode(c, 1006, "too many failed attempts, try again later", nil)
				c.Abort()
				return
			}
			if now.Sub(b.last) > failWindow {
				delete(loginFails, ip)
			}
		}
		loginFailMu.Unlock()
		c.Next()
	}
}

// RecordLoginFailure 登录失败时调用
func RecordLoginFailure(ip string) {
	loginFailMu.Lock()
	defer loginFailMu.Unlock()
	b, ok := loginFails[ip]
	now := time.Now()
	if !ok {
		loginFails[ip] = &failBucket{count: 1, last: now}
		return
	}
	if now.Sub(b.last) > failWindow {
		b.count = 0
	}
	b.count++
	b.last = now
	if b.count >= maxFailures {
		b.until = now.Add(blockDuration)
		b.count = 0
	}
}

// ResetLoginFailure 登录成功后清除计数
func ResetLoginFailure(ip string) {
	loginFailMu.Lock()
	defer loginFailMu.Unlock()
	delete(loginFails, ip)
}
