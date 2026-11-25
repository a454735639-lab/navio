package probe

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// 本地密钥文件：首次使用时生成，权限 600，只在本机有效
var secretPath = "./conf/probe/.secret"

func localSecret() ([]byte, error) {
	if b, err := os.ReadFile(secretPath); err == nil && len(b) >= 32 {
		return b, nil
	}
	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(secretPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(secretPath, buf, 0o600); err != nil {
		return nil, err
	}
	return buf, nil
}

func gcm() (cipher.AEAD, error) {
	secret, err := localSecret()
	if err != nil {
		return nil, err
	}
	key := sha256.Sum256(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// EncryptSecret 加密敏感字符串（DNS Token），返回 base64(nonce|ciphertext)
func EncryptSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	aead, err := gcm()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := aead.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptSecret 解密 EncryptSecret 的结果
func DecryptSecret(enc string) (string, error) {
	if enc == "" {
		return "", errors.New("empty secret")
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	aead, err := gcm()
	if err != nil {
		return "", err
	}
	if len(raw) < aead.NonceSize() {
		return "", errors.New("invalid secret payload")
	}
	nonce, ct := raw[:aead.NonceSize()], raw[aead.NonceSize():]
	plain, err := aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
