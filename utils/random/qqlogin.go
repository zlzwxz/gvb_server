package random

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomString 生成指定长度的随机字符串

func RandString(length int) string {

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)

	for i := range result {

		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))

		if err != nil {

			return ""

		}

		result[i] = charset[num.Int64()]

	}

	return string(result)

}
