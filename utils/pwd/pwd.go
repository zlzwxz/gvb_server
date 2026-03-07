package pwd

import (
	"golang.org/x/crypto/bcrypt"
)

// 加密密码
func HashPwd(pwd string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

// 验证密码
func ComparePasswords(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)

	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	if err != nil {
		return false
	}
	return true
}
