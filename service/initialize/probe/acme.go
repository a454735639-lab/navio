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
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
)

// acmeUser 满足 lego 的 User 接口
type acmeUser struct {
	Email string
	key   *ecdsa.PrivateKey
}

func (u *acmeUser) GetEmail() string                        { return u.Email }
func (u *acmeUser) GetRegistration() *registration.Resource { return nil }
func (u *acmeUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

// acmeProvider 内置 ACME（DNS-01），当前支持 Cloudflare，结构可扩展到其他 provider
type acmeProvider struct {
	cfg Config
}

func (p *acmeProvider) Name() string { return ProviderACME }

func (p *acmeProvider) Ensure() (string, string, error) {
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
	if strings.ToLower(p.cfg.DNSProvider) != ProviderPath && strings.ToLower(p.cfg.DNSProvider) != "cloudflare" {
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
	if err := client.Challenge.SetDNS01Provider(dnsProvider, dns01.AddRecursiveNameservers([]string{"1.1.1.1:53", "8.8.8.8:53"})); err != nil {
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
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
