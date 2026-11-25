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

// Start 启动探针（读取数据库配置）
func Start() {
	Apply(LoadConfig())
}

var state = struct {
	mu       sync.Mutex
	srv      *http.Server
	reloader *certReloader
	port     int
}{}

var renewOnce sync.Once

func startRenewLoop(p CertProvider) {
	renewOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(12 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				if _, _, err := p.Ensure(); err != nil {
					global.Logger.Warnf("[probe] acme renew check failed: %v", err)
				}
			}
		}()
	})
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
		// ACME 签发是长任务：异步执行，先起监听（证书就绪后按 mtime 自动加载）
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
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler(),
		ReadHeaderTimeout: 5 * time.Second,
		TLSConfig:         &tls.Config{GetCertificate: reloader.getCertificate, MinVersion: tls.VersionTLS12},
	}
	state.srv = srv
	state.reloader = reloader
	state.port = cfg.Port

	// ACME：后台巡检续期（每 12 小时判断一次，剩余 <30 天时自动续）
	if cfg.Provider == ProviderACME {
		startRenewLoop(provider)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Errorf("[probe] panic: %v", r)
			}
		}()
		global.Logger.Infof("[probe] listening on :%d (provider=%s)", cfg.Port, provider.Name())
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			global.Logger.Errorf("[probe] server stopped: %v", err)
		}
	}()
}
