package user_api

import (
	"fmt"
	"strings"
	"time"

	"gvb-server/service/redis_ser"
)

const (
	loginFailLimit  int64         = 8
	loginFailWindow time.Duration = 15 * time.Minute
)

func buildLoginFailIPKey(ip string) string {
	return fmt.Sprintf("login_fail:ip:%s", strings.TrimSpace(ip))
}

func buildLoginFailUserKey(userName string) string {
	return fmt.Sprintf("login_fail:user:%s", strings.ToLower(strings.TrimSpace(userName)))
}

func isLoginRateLimited(ip string, userName string) bool {
	return redis_ser.GetInt64(buildLoginFailIPKey(ip)) >= loginFailLimit ||
		redis_ser.GetInt64(buildLoginFailUserKey(userName)) >= loginFailLimit
}

func markLoginFailed(ip string, userName string) {
	_, _ = redis_ser.IncrWithTTL(buildLoginFailIPKey(ip), loginFailWindow)
	_, _ = redis_ser.IncrWithTTL(buildLoginFailUserKey(userName), loginFailWindow)
}

func clearLoginFailRecord(ip string, userName string) {
	redis_ser.DelKeys(buildLoginFailIPKey(ip), buildLoginFailUserKey(userName))
}
