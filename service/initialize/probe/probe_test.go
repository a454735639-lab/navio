package probe

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeSelfSigned(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.local"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     []string{"test.local"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0o600); err != nil {
		t.Fatal(err)
	}
	return certFile, keyFile
}

func TestNormalizeAndValidate(t *testing.T) {
	cfg := normalize(Config{})
	if cfg.Port != 8443 || cfg.Provider != ProviderPath {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}

	cfg = Config{Enabled: true, Port: 70000, Provider: ProviderPath, CertPath: "a", KeyPath: "b"}
	if err := validate(cfg); err == nil {
		t.Fatal("expected invalid port error")
	}

	cfg = Config{Enabled: true, Port: 8443, Provider: "nope"}
	if err := validate(cfg); err == nil {
		t.Fatal("expected unknown provider error")
	}
}

func TestFileProvider(t *testing.T) {
	cert, key := writeSelfSigned(t)
	p := &fileProvider{name: ProviderPath, cert: cert, key: key}

	cf, kf, err := p.Ensure()
	if err != nil || cf != cert || kf != key {
		t.Fatalf("ensure failed: %v", err)
	}

	st, err := p.Status()
	if err != nil {
		t.Fatal(err)
	}
	if st.NotAfter.IsZero() || st.Issuer == "" {
		t.Fatalf("unexpected status: %+v", st)
	}

	info, err := ParseCertInfo(cert)
	if err != nil || info["subject"] == "" {
		t.Fatalf("parse cert info failed: %v %v", err, info)
	}
}

func TestFileProviderMissingFile(t *testing.T) {
	p := &fileProvider{name: ProviderUpload, cert: "/nope/cert.pem", key: "/nope/key.pem"}
	if _, _, err := p.Ensure(); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSecretRoundTrip(t *testing.T) {
	old := secretPath
	secretPath = filepath.Join(t.TempDir(), ".secret")
	defer func() { secretPath = old }()

	// 纯函数部分：加解密往返（依赖 SystemSetting 的用例放到集成环境）
	enc, err := EncryptSecret("cf-token-123")
	if err != nil {
		t.Fatal(err)
	}
	if enc == "" {
		t.Fatal("empty ciphertext")
	}
	dec, err := DecryptSecret(enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != "cf-token-123" {
		t.Fatalf("unexpected plaintext: %s", dec)
	}
	if _, err := DecryptSecret("bad"); err == nil {
		t.Fatal("expected error for bad payload")
	}
}

func TestValidateACME(t *testing.T) {
	cfg := Config{Enabled: true, Port: 8443, Provider: ProviderACME}
	if err := validate(cfg); err == nil {
		t.Fatal("expected error for missing acme fields")
	}
	cfg.ACMEEmail = "a@b.c"
	cfg.ACMEDomains = []string{"probe.example.com"}
	cfg.DNSProvider = "cloudflare"
	cfg.DNSTokenEnc = "enc"
	if err := validate(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
