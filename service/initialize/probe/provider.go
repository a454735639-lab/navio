package probe

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// CertStatus 证书状态，供诊断页/接口使用
type CertStatus struct {
	Provider  string    `json:"provider"`
	CertFile  string    `json:"certFile"`
	KeyFile   string    `json:"keyFile"`
	NotAfter  time.Time `json:"notAfter"`
	Issuer    string    `json:"issuer"`
	LastError string    `json:"lastError,omitempty"`
}

// CertProvider 证书来源抽象；M2 会增加 ACME 实现
type CertProvider interface {
	Name() string
	Ensure() (certFile, keyFile string, err error)
	Status() (CertStatus, error)
}

// NewProvider 根据配置构造 provider
func NewProvider(cfg Config) (CertProvider, error) {
	switch cfg.Provider {
	case ProviderPath:
		return &fileProvider{name: ProviderPath, cert: cfg.CertPath, key: cfg.KeyPath}, nil
	case ProviderUpload:
		return &fileProvider{name: ProviderUpload, cert: cfg.CertPath, key: cfg.KeyPath}, nil
	case ProviderACME:
		return &acmeProvider{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}

// fileProvider 覆盖 path / upload 两种来源：都是本地文件，只是来源语义不同
type fileProvider struct {
	name string
	cert string
	key  string
}

func (p *fileProvider) Name() string { return p.name }

func (p *fileProvider) Ensure() (string, string, error) {
	if _, err := loadPair(p.cert, p.key); err != nil {
		return "", "", err
	}
	return p.cert, p.key, nil
}

func (p *fileProvider) Status() (CertStatus, error) {
	st := CertStatus{Provider: p.name, CertFile: p.cert, KeyFile: p.key}
	cert, err := loadPair(p.cert, p.key)
	if err != nil {
		st.LastError = err.Error()
		return st, err
	}
	if len(cert.Certificate) > 0 {
		if leaf, err := x509.ParseCertificate(cert.Certificate[0]); err == nil {
			st.NotAfter = leaf.NotAfter
			st.Issuer = leaf.Issuer.String()
		}
	}
	return st, nil
}

func loadPair(certFile, keyFile string) (*tls.Certificate, error) {
	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("cert/key path is empty")
	}
	if _, err := os.Stat(certFile); err != nil {
		return nil, err
	}
	if _, err := os.Stat(keyFile); err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// ParseCertInfo 读取证书可读信息（供接口返回）
func ParseCertInfo(certFile string) (map[string]any, error) {
	b, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("invalid pem")
	}
	leaf, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"subject":  leaf.Subject.String(),
		"issuer":   leaf.Issuer.String(),
		"notAfter": leaf.NotAfter,
		"dnsNames": leaf.DNSNames,
	}
	if j, err := json.Marshal(out); err == nil && len(j) == 0 {
		return out, nil
	}
	return out, nil
}
