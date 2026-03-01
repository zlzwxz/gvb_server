package user_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
	"strings"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
)

type UserUpdateNicknameRequest struct {
	NickName string `json:"nick_name" structs:"nick_name"`
	Sign     string `json:"sign" structs:"sign"`
	Link     string `json:"link" structs:"link"`
}

// UserUpdateNickName 修改当前登录人的昵称，签名，链接
func (UserApi) UserUpdateNickName(c *gin.Context) {
	var cr UserUpdateNicknameRequest
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)

	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	var newMaps = map[string]interface{}{}
	maps := structs.Map(cr)
	for key, v := range maps {
		if val, ok := v.(string); ok && strings.TrimSpace(val) != "" {
			newMaps[key] = val
		}
	}
	var userModel models.UserModel
	err = global.DB.Debug().Take(&userModel, claims.UserID).Error
	if err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}
	err = global.DB.Model(&userModel).Updates(newMaps).Debug().Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("修改用户信息失败", c)
		return
	}
	res.OkWithMessage("修改个人信息成功", c)

}
