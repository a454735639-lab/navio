package probe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
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

func (c *certReloader) setPaths(certFile, keyFile string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.certFile != certFile || c.keyFile != keyFile {
		c.certFile = certFile
		c.keyFile = keyFile
		c.cert = nil // 强制下次握手重新加载
	}
}

func (c *certReloader) getCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	// 所有共享字段的读取都在锁内完成，避免与 setPaths 的写入产生 data race
	c.mu.Lock()
	certFile, keyFile := c.certFile, c.keyFile
	cert, modTime := c.cert, c.modTime
	c.mu.Unlock()

	// 同时检查证书与私钥的 mtime：单独替换私钥也应触发热加载
	certInfo, certErr := os.Stat(certFile)
	keyInfo, keyErr := os.Stat(keyFile)
	needReload := cert == nil ||
		(certErr == nil && certInfo.ModTime().After(modTime)) ||
		(keyErr == nil && keyInfo.ModTime().After(modTime))
	if needReload && certErr == nil && keyErr == nil {
		if loaded, err := tls.LoadX509KeyPair(certFile, keyFile); err == nil {
			latest := certInfo.ModTime()
			if keyInfo.ModTime().After(latest) {
				latest = keyInfo.ModTime()
			}
			c.mu.Lock()
			c.cert = &loaded
			c.modTime = latest
			c.mu.Unlock()
			cert = &loaded
		}
	}

	if cert == nil {
		return nil, fmt.Errorf("certificate not loaded: %s", certFile)
	}
	return cert, nil
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

// Start 启动探针（读取数据库配置）
func Start() {
	Apply(LoadConfig())
}

var state = struct {
	mu          sync.Mutex
	srv         *http.Server
	reloader    *certReloader
	port        int
	renewCancel context.CancelFunc // 取消当前续期循环，配置变更/停用时调用
}{}

// startRenewLoop 启动 ACME 续期循环。
// 关键：每次 tick 都从最新配置重建 provider（而非永久捕获首次 provider），
// 这样用户修改域名/Cloudflare Token/证书路径后续期会用新配置签发。
// 返回的 cancel 用于配置变更或停用时停止旧循环，避免重复 goroutine 与停用后仍写盘。
func startRenewLoop() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(12 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				global.Logger.Infoln("[probe] acme renew loop stopped")
				return
			case <-ticker.C:
				cfg := LoadConfig()
				if !cfg.Enabled || cfg.Provider != ProviderACME {
					return
				}
				p, err := NewProvider(cfg)
				if err != nil {
					global.Logger.Warnf("[probe] renew provider init failed: %v", err)
					continue
				}
				if _, _, err := p.Ensure(); err != nil {
					global.Logger.Warnf("[probe] acme renew check failed: %v", err)
				}
			}
		}
	}()
	return cancel
}

// stopRenewLoop 停止当前续期循环（调用方须持有 state.mu）
func stopRenewLoop() {
	if state.renewCancel != nil {
		state.renewCancel()
		state.renewCancel = nil
	}
}

// Apply 热应用配置：端口/开关变化时重建监听；仅证书路径变化时更新文件路径（mtime 热加载）。
func Apply(cfg Config) {
	state.mu.Lock()
	defer state.mu.Unlock()

	shutdown := func() {
		if state.srv != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_ = state.srv.Shutdown(ctx)
			cancel()
			state.srv = nil
			state.port = 0
			global.Logger.Infoln("[probe] stopped")
		}
	}

	if !cfg.Enabled {
		stopRenewLoop() // 停用探针时一并停止续期，避免后台继续签发写盘
		shutdown()
		return
	}

	provider, err := NewProvider(cfg)
	if err != nil {
		global.Logger.Warnf("[probe] provider init failed: %v", err)
		return
	}

	certFile, keyFile := cfg.CertPath, cfg.KeyPath
	if cfg.Provider == ProviderACME {
		// ACME 签发是长任务：异步执行，先起监听（证书就绪后按 mtime 自动加载）。
		// 串行化由 acmeProvider.Ensure 内部的 acmeEnsureMu 完成（覆盖初次签发与续期两条路径）。
		go func(p CertProvider) {
			if _, _, err := p.Ensure(); err != nil {
				global.Logger.Warnf("[probe] acme ensure failed: %v", err)
			}
		}(provider)
	} else {
		if certFile, keyFile, err = provider.Ensure(); err != nil {
			global.Logger.Warnf("[probe] certificate unavailable: %v", err)
			return
		}
	}

	// 端口未变且已在运行：只更新证书路径，不重启监听
	if state.srv != nil && state.port == cfg.Port && state.reloader != nil {
		state.reloader.setPaths(certFile, keyFile)
		return
	}

	shutdown()

	reloader := &certReloader{certFile: certFile, keyFile: keyFile}
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.BindAddr, cfg.Port),
		Handler:           handler(),
		ReadHeaderTimeout: 5 * time.Second,
		TLSConfig:         &tls.Config{GetCertificate: reloader.getCertificate, MinVersion: tls.VersionTLS12},
	}
	state.srv = srv
	state.reloader = reloader
	state.port = cfg.Port

	// ACME：后台巡检续期（每 12 小时判断一次，剩余 <30 天时自动续）。
	// 配置变更时先取消旧循环再启动新的，避免多个循环并存与绑定旧 provider。
	stopRenewLoop()
	if cfg.Provider == ProviderACME {
		state.renewCancel = startRenewLoop()
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Errorf("[probe] panic: %v", r)
			}
		}()
		global.Logger.Infof("[probe] listening on %s:%d (provider=%s)", cfg.BindAddr, cfg.Port, provider.Name())
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			global.Logger.Errorf("[probe] server stopped: %v", err)
		}
	}()
}
