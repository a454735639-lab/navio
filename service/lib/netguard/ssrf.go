// Package netguard 提供统一的出站 HTTP 请求安全校验，防止 SSRF。
// 所有由用户可控 URL 发起的服务端请求都应通过 ValidateURL + NewSafeClient。
package netguard

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrUnsafeURL   = errors.New("unsafe url")
	ErrBlockedHost = errors.New("blocked host")
)

// 云厂商元数据地址，SSRF 头号目标
var blockedCIDRs []*net.IPNet

func init() {
	for _, cidr := range []string{
		"169.254.169.254/32", // 云元数据（AWS/GCP/Azure/阿里云等）
		"100.64.0.0/10",      // CGNAT，常见于内网/容器
	} {
		_, n, err := net.ParseCIDR(cidr)
		if err == nil {
			blockedCIDRs = append(blockedCIDRs, n)
		}
	}
}

// isBlockedIP 判断 IP 是否属于不应被服务端请求的地址段。
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	for _, n := range blockedCIDRs {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateURL 校验 URL 协议与主机，拒绝非 http(s) 及私网/环回/链路本地/元数据地址。
// 注意：这里解析的是主机名字符串形式的 IP；对域名 rebinding 的防护在 SafeDialContext 中完成。
func ValidateURL(rawURL string) (*url.URL, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("%w: empty url", ErrUnsafeURL)
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: parse: %v", ErrUnsafeURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("%w: scheme %q not allowed", ErrUnsafeURL, u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("%w: empty host", ErrUnsafeURL)
	}
	// 用户名密码内嵌 URL 一律拒绝（https://user:pass@host 这类混淆）
	if u.User != nil {
		return nil, fmt.Errorf("%w: userinfo not allowed", ErrUnsafeURL)
	}
	// 字面量 IP 直接判定
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("%w: %s", ErrBlockedHost, host)
		}
		return u, nil
	}
	// 域名：解析出的所有 IP 必须都安全，否则拒绝（防 DNS 指向内网）
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve %s: %v", ErrUnsafeURL, host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("%w: no address for %s", ErrUnsafeURL, host)
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("%w: %s resolves to %s", ErrBlockedHost, host, ip)
		}
	}
	return u, nil
}

// safeDialContext 在 TCP 连接建立阶段再次校验解析出的 IP，防止 DNS rebinding
// （校验时刻与实际拨号时刻域名被解析到不同 IP）。
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, ipa := range ips {
		if isBlockedIP(ipa.IP) {
			continue
		}
		conn, e := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ipa.IP.String(), port))
		if e == nil {
			return conn, nil
		}
		lastErr = e
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("%w: %s", ErrBlockedHost, host)
}

// NewSafeClient 返回带 SSRF 防护的 http.Client：
//  - 拨号阶段校验 IP（防 rebinding）
//  - 限制重定向次数，且每跳都校验目标
//  - 设置总超时，避免慢速攻击长期占用连接
func NewSafeClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:         safeDialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConns:        20,
			IdleConnTimeout:     30 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			_, err := ValidateURL(req.URL.String())
			return err
		},
	}
}
