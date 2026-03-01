package middleware

import (
	"fmt"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/plugins/log_stash"
	"gvb-server/service/redis_ser"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
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
		// 判断是否在redis中
		if redis_ser.CheckLogout(token) {
			res.FailWithMessage("token已失效", c)
			c.Abort()
			return
		}
		// 登录的用户
		c.Set("claims", claims)
	}
}

// 管理员才能使用的中间件
func JwtAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := log_stash.NewLogByGin(c)
		token := c.Request.Header.Get("token")
		if token == "" {
			log.Info(fmt.Sprintf("未携带token"))
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
		// 登录的用户
		if claims.Role != int(ctype.PermissionAdmin) {
			log.Info(fmt.Sprintf("权限错误"))
			res.FailWithMessage("权限错误", c)
			c.Abort()
			return
		}
		c.Set("claims", claims)
	}
}
