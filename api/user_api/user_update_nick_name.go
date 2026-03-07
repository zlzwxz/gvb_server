package user_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
	"gvb-server/utils/sanitize"
	"strings"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
)

// UserUpdateNicknameRequest 修改用户昵称请求参数
type UserUpdateNicknameRequest struct {
	NickName string `json:"nick_name" structs:"nick_name" swag:"description:用户昵称"`
	Sign     string `json:"sign" structs:"sign" swag:"description:用户签名"`
	Link     string `json:"link" structs:"link" swag:"description:用户链接"`
	Avatar   string `json:"avatar" structs:"avatar" swag:"description:用户头像"`
}

// UserUpdateNickName 修改当前登录人的昵称，签名，链接
// @Summary 修改用户信息
// @Description 修改当前登录用户的昵称、签名、链接
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body UserUpdateNicknameRequest true "用户信息"
// @Success 200 {object} res.Response{msg=string} "修改成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "用户不存在"
// @Router /api/user_update_nick_name [put]
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
			newMaps[key] = strings.TrimSpace(val)
		}
	}
	if link, ok := newMaps["link"].(string); ok {
		cleanLink := sanitize.CleanURL(link, true)
		if cleanLink == "" {
			res.FailWithMessage("主页链接仅支持 http/https 或站内相对路径", c)
			return
		}
		newMaps["link"] = cleanLink
	}
	if nickName, ok := newMaps["nick_name"].(string); ok {
		if len([]rune(nickName)) > 30 {
			res.FailWithMessage("昵称长度不能超过 30 个字符", c)
			return
		}
		newMaps["nick_name"] = strings.ReplaceAll(strings.ReplaceAll(nickName, "<", ""), ">", "")
	}
	if sign, ok := newMaps["sign"].(string); ok {
		if len([]rune(sign)) > 100 {
			res.FailWithMessage("签名长度不能超过 100 个字符", c)
			return
		}
		newMaps["sign"] = strings.ReplaceAll(strings.ReplaceAll(sign, "<", ""), ">", "")
	}
	if avatar, ok := newMaps["avatar"].(string); ok {
		cleanAvatar := sanitize.CleanURL(avatar, true)
		if cleanAvatar == "" {
			res.FailWithMessage("头像地址不合法", c)
			return
		}
		newMaps["avatar"] = cleanAvatar
	}
	var userModel models.UserModel
	err = global.DB.Take(&userModel, claims.UserID).Error
	if err != nil {
		res.FailWithMessage("用户不存在", c)
		return
	}
	err = global.DB.Model(&userModel).Updates(newMaps).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("修改用户信息失败", c)
		return
	}
	res.OkWithMessage("修改个人信息成功", c)

}
