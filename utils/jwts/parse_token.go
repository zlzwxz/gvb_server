package jwts

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"gvb-server/global"
	"strings"
)

// ParseToken 解析 token
func ParseToken(tokenStr string) (*CustomClaims, error) {
	secret := strings.TrimSpace(global.Config.Jwt.Secret)
	if secret == "" {
		return nil, errors.New("jwt secret 未配置")
	}
	MySecret := []byte(secret)
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return MySecret, nil
	})
	if err != nil {
		logrus.Error(fmt.Sprintf("token 错误 parse err: %s", err.Error()))
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
