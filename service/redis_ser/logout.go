package redis_ser

import (
	"gvb-server/global"
	"time"
)

const prefix = "logout_"

// Logout 针对注销的操作
func Logout(token string, diff time.Duration) error {
	err := global.Redis.Set(prefix+token, "", diff).Err()
	return err
}

func CheckLogout(token string) bool {
	if token == "" {
		return false
	}
	return global.Redis.Exists(prefix+token).Val() > 0
}
