package tag_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

// TagRequest 标签创建请求参数
type TagRequest struct {
	Title string `json:"title" binding:"required" msg:"请输入标签" swag:"description:标签名称"` // 标签
}

// TagCreateView 创建标签
// @Summary 创建标签
// @Description 创建新标签，会检查标签是否已存在
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body TagRequest true "标签信息"
// @Success 200 {object} res.Response{msg=string} "创建成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 409 {object} res.Response "标签已存在"
// @Router /api/tags [post]
func (TagApi) TagCreateView(c *gin.Context) {
	var cr TagRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	// 重复的判断
	var tag models.TagModel
	err = global.DB.Take(&tag, "title = ?", cr.Title).Error
	if err == nil {
		res.FailWithMessage("该标签已存在", c)
		return
	}

	err = global.DB.Create(&models.TagModel{
		Title: cr.Title,
	}).Error

	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("添加标签失败", c)
		return
	}

	res.OkWithMessage("添加标签成功", c)
}
