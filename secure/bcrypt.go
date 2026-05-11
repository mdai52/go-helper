package secure

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHash 使用 bcrypt 对密码进行哈希
func BcryptHash(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(hash), nil
}

// BcryptVerify 验证密码是否匹配存储的哈希值
func BcryptVerify(password, hashedPassword string) bool {
	if hashedPassword == "" || password == "" {
		return hashedPassword == password
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// IsBcrypt 检查密码是否是 bcrypt 格式
func IsBcrypt(password string) bool {
	return len(password) >= 4 && password[0] == '$' && password[1] == '2' && password[3] == '$'
}
