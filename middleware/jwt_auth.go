package middleware

import (
	"fmt"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/plugins/log_stash"
	"gvb-server/service/redis_ser"
	"gvb-server/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
)

// JwtAuth 是“普通登录用户即可访问”的鉴权中间件。
// 它完成的事情很固定：
// 1. 从请求头中提取 token；
// 2. 解析 JWT；
// 3. 检查 token 是否已被加入 Redis 黑名单（退出登录后会失效）；
// 4. 解析成功后把 claims 放进 Gin 上下文，供后续 handler 使用。
func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := resolveToken(c)
		if token == "" {
			res.FailWithMessage("未携带token", c)
			c.Abort()
			return
		}
		claims, err := jwts.ParseToken(token)
		if err != nil {
			res.FailWithMessage("token错误", c)
			c.Abort()
			return
		}
		// 判断是否在 redis 中。
		// 如果命中黑名单，通常代表用户已经退出登录，旧 token 不能再继续使用。
		if redis_ser.CheckLogout(token) {
			res.FailWithMessage("token已失效", c)
			c.Abort()
			return
		}
		// 登录成功的用户信息塞到上下文里，后续 handler 可通过 `c.Get("claims")` 读取。
		c.Set("claims", claims)
	}
}

// JwtAdmin 是“必须管理员才能访问”的鉴权中间件。
// 它在 JwtAuth 的基础上，多做了一次角色判断。
func JwtAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := log_stash.NewLogByGin(c)
		token := resolveToken(c)
		if token == "" {
			log.Info(fmt.Sprintf("未携带token"))
			res.FailWithMessage("未携带token", c)
			c.Abort()
			return
		}
		claims, err := jwts.ParseToken(token)
		if err != nil {
			log.Info(fmt.Sprintf("token错误"))
			res.FailWithMessage("未携带token,token错误", c)
			c.Abort()
			return
		}
		if redis_ser.CheckLogout(token) {
			log.Info(fmt.Sprintf("token已失效"))
			res.FailWithMessage("token已失效", c)
			c.Abort()
			return
		}
		// 只有管理员角色才能通过。
		if claims.Role != int(ctype.PermissionAdmin) {
			log.Info(fmt.Sprintf("权限错误"))
			res.FailWithMessage("权限错误", c)
			c.Abort()
			return
		}
		c.Set("claims", claims)
	}
}

// resolveToken 统一兼容两种 token 传法：
// 1. 历史写法：`token: xxx`
// 2. 标准写法：`Authorization: Bearer xxx`
func resolveToken(c *gin.Context) string {
	token := strings.TrimSpace(c.Request.Header.Get("token"))
	if token != "" {
		return token
	}
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}
