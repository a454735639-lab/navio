package password

import (
	"strings"

	"golang.org/x/crypto/bcrypt"

	"sun-panel/lib/cmn"
)

const bcryptCost = 10

// Hash 生成 bcrypt 哈希
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify 校验密码：兼容 md5^3 旧哈希，返回 (是否匹配, 是否需要升级)
func Verify(stored, plain string) (bool, bool) {
	if IsHashed(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil, false
	}
	return cmn.PasswordEncryption(plain) == stored, true
}

// IsHashed 判断是否已经是 bcrypt 哈希
func IsHashed(s string) bool {
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}
