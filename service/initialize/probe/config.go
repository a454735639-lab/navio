package probe

import (
	"fmt"
	"strings"

	"sun-panel/global"
)

// 存储键：探针配置以数据库为准，conf.ini 只在首次启动时做引导
const storeKey = "probe_config"

// 证书来源类型
const (
	ProviderPath   = "path"   // 使用已有证书文件
	ProviderUpload = "upload" // 用户上传到 conf/probe/
	ProviderACME   = "acme"   // 内置 ACME（M2 实现）
)

// Config 探针配置（全局、部署级）
type Config struct {
	Enabled  bool   `json:"enabled"`
	Port     int    `json:"port"`
	Provider string `json:"provider"`
	CertPath string `json:"certPath"`
	KeyPath  string `json:"keyPath"`

	// ACME（provider=acme）
	ACMEEmail   string   `json:"acmeEmail,omitempty"`
	ACMEDomains []string `json:"acmeDomains,omitempty"`
	DNSProvider string   `json:"dnsProvider,omitempty"`
	DNSTokenEnc string   `json:"dnsTokenEnc,omitempty"` // 加密后的 Token，接口返回前会清空
	Staging     bool     `json:"staging,omitempty"`
}

func DefaultConfig() Config {
	return Config{Enabled: false, Port: 8443, Provider: ProviderPath}
}

// LoadConfig 优先读数据库；没有则用 conf.ini 引导并写库。
func LoadConfig() Config {
	var cfg Config
	if err := global.SystemSetting.GetValueByInterface(storeKey, &cfg); err == nil && cfg.Port > 0 {
		return normalize(cfg)
	}

	cfg = DefaultConfig()
	cfg.Enabled = strings.EqualFold(strings.TrimSpace(global.Config.GetValueStringOrDefault("probe", "enabled")), "true")
	if p := global.Config.GetValueIntOrDefault("probe", "port", 8443); p > 0 {
		cfg.Port = p
	}
	cfg.Provider = ProviderPath
	cfg.CertPath = strings.TrimSpace(global.Config.GetValueString("probe", "cert_path"))
	cfg.KeyPath = strings.TrimSpace(global.Config.GetValueString("probe", "key_path"))

	_ = global.SystemSetting.Set(storeKey, cfg)
	return normalize(cfg)
}

// SaveConfig 校验后写库（下一次读取即生效；TLS 证书由 mtime 热加载）
func SaveConfig(cfg Config) error {
	cfg = normalize(cfg)
	if err := validate(cfg); err != nil {
		return err
	}
	if err := global.SystemSetting.Set(storeKey, cfg); err != nil {
		return err
	}
	Apply(cfg)
	return nil
}

func normalize(cfg Config) Config {
	if cfg.Port == 0 {
		cfg.Port = 8443
	}
	if cfg.Provider == "" {
		cfg.Provider = ProviderPath
	}
	if cfg.Provider == ProviderACME {
		if cfg.CertPath == "" {
			cfg.CertPath = "./conf/probe/acme/fullchain.pem"
		}
		if cfg.KeyPath == "" {
			cfg.KeyPath = "./conf/probe/acme/key.pem"
		}
		for i := range cfg.ACMEDomains {
			cfg.ACMEDomains[i] = strings.TrimSpace(cfg.ACMEDomains[i])
		}
	}
	cfg.Provider = strings.ToLower(strings.TrimSpace(cfg.Provider))
	cfg.CertPath = strings.TrimSpace(cfg.CertPath)
	cfg.KeyPath = strings.TrimSpace(cfg.KeyPath)
	return cfg
}

func validate(cfg Config) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid probe port: %d", cfg.Port)
	}
	switch cfg.Provider {
	case ProviderPath, ProviderUpload:
		if cfg.CertPath == "" || cfg.KeyPath == "" {
			return fmt.Errorf("certPath and keyPath are required")
		}
		if _, err := loadPair(cfg.CertPath, cfg.KeyPath); err != nil {
			return fmt.Errorf("certificate invalid: %w", err)
		}
	case ProviderACME:
		if strings.TrimSpace(cfg.ACMEEmail) == "" {
			return fmt.Errorf("acmeEmail is required")
		}
		if len(cfg.ACMEDomains) == 0 {
			return fmt.Errorf("acmeDomains is required")
		}
		if strings.TrimSpace(cfg.DNSProvider) == "" {
			return fmt.Errorf("dnsProvider is required")
		}
		if cfg.DNSTokenEnc == "" {
			return fmt.Errorf("dns token is required")
		}
	default:
		return fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
	return nil
}
