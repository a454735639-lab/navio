package probe

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
)

// propagationDelayProvider 在写入 TXT 后等待一段时间，避免 LE 立即查询拿不到记录
type propagationDelayProvider struct {
	provider *cloudflare.DNSProvider
	delay    time.Duration
}

func (p *propagationDelayProvider) Present(domain, token, keyAuth string) error {
	if err := p.provider.Present(domain, token, keyAuth); err != nil {
		return err
	}
	time.Sleep(p.delay)
	return nil
}

func (p *propagationDelayProvider) CleanUp(domain, token, keyAuth string) error {
	return p.provider.CleanUp(domain, token, keyAuth)
}

// acmeUser 满足 lego 的 User 接口
type acmeUser struct {
	Email string
	key   *ecdsa.PrivateKey
}

func (u *acmeUser) GetEmail() string                        { return u.Email }
func (u *acmeUser) GetRegistration() *registration.Resource { return nil }
func (u *acmeUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

// acmeEnsureMu 串行化所有 ACME 签发，避免并发订单争用同一 .tmp 文件导致
// 证书与私钥来自不同订单、以及 Let's Encrypt 重复签发限流。
var acmeEnsureMu sync.Mutex

// acmeProvider 内置 ACME（DNS-01），当前支持 Cloudflare，结构可扩展到其他 provider
type acmeProvider struct {
	cfg Config
}

func (p *acmeProvider) Name() string { return ProviderACME }

func (p *acmeProvider) Ensure() (string, string, error) {
	// 串行化整个「检查 + 签发」过程：Apply 初次签发与续期循环可能并发进入，
	// 必须持有同一把锁，否则会并发走完整订单并争用证书文件。
	acmeEnsureMu.Lock()
	defer acmeEnsureMu.Unlock()

	certFile, keyFile := p.paths()

	// 证书仍有效（剩余 > 30 天）则直接复用
	if st, err := p.Status(); err == nil && time.Until(st.NotAfter) > 30*24*time.Hour {
		return certFile, keyFile, nil
	}
	if err := p.issue(certFile, keyFile); err != nil {
		return "", "", err
	}
	return certFile, keyFile, nil
}

func (p *acmeProvider) Status() (CertStatus, error) {
	certFile, keyFile := p.paths()
	st := CertStatus{Provider: ProviderACME, CertFile: certFile, KeyFile: keyFile}
	cert, err := loadPair(certFile, keyFile)
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

func (p *acmeProvider) paths() (string, string) {
	certFile, keyFile := p.cfg.CertPath, p.cfg.KeyPath
	if certFile == "" {
		certFile = "./conf/probe/acme/fullchain.pem"
	}
	if keyFile == "" {
		keyFile = "./conf/probe/acme/key.pem"
	}
	return certFile, keyFile
}

func (p *acmeProvider) issue(certFile, keyFile string) error {
	token, err := DecryptSecret(p.cfg.DNSTokenEnc)
	if err != nil {
		return fmt.Errorf("dns token unavailable: %w", err)
	}
	// ProviderPath 是证书来源而非 DNS 服务商，只有 cloudflare 是合法的 DNS-01 provider
	if strings.ToLower(p.cfg.DNSProvider) != "cloudflare" {
		return fmt.Errorf("unsupported dns provider: %s", p.cfg.DNSProvider)
	}

	dnsProvider, err := cloudflare.NewDNSProviderConfig(&cloudflare.Config{
		AuthToken: token,
		TTL:       120,
	})
	if err != nil {
		return fmt.Errorf("cloudflare provider: %w", err)
	}

	userKey, err := certcrypto.GeneratePrivateKey(certcrypto.EC256)
	if err != nil {
		return err
	}
	ecKey, ok := userKey.(*ecdsa.PrivateKey)
	if !ok {
		return fmt.Errorf("unexpected account key type")
	}
	user := &acmeUser{Email: p.cfg.ACMEEmail, key: ecKey}

	cfg := lego.NewConfig(user)
	if p.cfg.Staging {
		cfg.CADirURL = lego.LEDirectoryStaging
	}
	client, err := lego.NewClient(cfg)
	if err != nil {
		return err
	}
	// Cloudflare 记录写入后，等待其同步到公共解析，再让 LE 校验
	delayed := &propagationDelayProvider{provider: dnsProvider, delay: 25 * time.Second}

	opts := []dns01.ChallengeOption{
		dns01.AddRecursiveNameservers([]string{"1.1.1.1:53", "8.8.8.8:53"}),
		dns01.AddDNSTimeout(2 * time.Minute),
		dns01.DisableCompletePropagationRequirement(),
	}
	if err := client.Challenge.SetDNS01Provider(delayed, opts...); err != nil {
		return err
	}
	if _, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true}); err != nil {
		return err
	}

	req := certificate.ObtainRequest{Domains: p.cfg.ACMEDomains, Bundle: true}
	res, err := client.Certificate.Obtain(req)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(certFile), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(keyFile), 0o755); err != nil {
		return err
	}
	if err := writeFileAtomic(certFile, res.Certificate, 0o644); err != nil {
		return err
	}
	keyPEM, err := encodePrivateKey(res.PrivateKey)
	if err != nil {
		return err
	}
	return writeFileAtomic(keyFile, keyPEM, 0o600)
}

func encodePrivateKey(key any) ([]byte, error) {
	// lego 在部分版本里直接返回 PEM 字节
	if b, ok := key.([]byte); ok {
		if block, _ := pem.Decode(b); block != nil {
			return b, nil
		}
		return nil, fmt.Errorf("invalid private key bytes")
	}
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}), nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b}), nil
	default:
		return nil, fmt.Errorf("unsupported key type %T", key)
	}
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	// 用唯一临时文件名，避免并发签发（不同配置/订单）争用固定 .tmp 相互覆盖
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	defer os.Remove(tmpName) // Rename 成功后 tmpName 已不存在，Remove 是无害的
	if err := os.WriteFile(tmpName, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	// 挂载卷可能带宽松权限，这里强制收敛
	return os.Chmod(path, perm)
}
