package cmn

import "fmt"

// MaskToken 日志脱敏：只留前 4 后 4，其余打码
func MaskToken(t string) string {
	if len(t) <= 8 {
		return fmt.Sprintf("%s***(len=%d)", t[:len(t)/2], len(t))
	}
	return fmt.Sprintf("%s****%s(len=%d)", t[:4], t[len(t)-4:], len(t))
}

// MaskPassword 密码永远不进日志
func MaskPassword(s string) string {
	return fmt.Sprintf("***(%d chars)", len(s))
}
