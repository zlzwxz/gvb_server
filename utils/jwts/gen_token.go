package jwts

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"gvb-server/global"
	"strings"
	"time"
)

// GenToken 创建 Token
func GenToken(user JwtPayLoad) (string, error) {
	secret := strings.TrimSpace(global.Config.Jwt.Secret)
	if secret == "" {
		return "", errors.New("jwt secret 未配置")
	}
	MySecret = []byte(secret)
	claim := CustomClaims{
		user,
		jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Hour * time.Duration(global.Config.Jwt.Expires))}, // 默认2小时过期
			Issuer:    global.Config.Jwt.Issuer,                                                                     // 签发人
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(MySecret)
}
