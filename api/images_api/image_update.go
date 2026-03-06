package images_api

import (
	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
)

type ImageUpdateRequest struct {
	ID   uint   `json:"id" binding:"required" msg:"请选择文件id"`
	Name string `json:"name" binding:"required" msg:"请输入文件名称"`
}

// ImageUpdateView 更新图片信息
// @Tags 图片管理
// @Summary 更新图片名称
// @Description 根据图片ID更新图片名称
// @Accept json
// @Produce json
// @Param data body ImageUpdateRequest true "图片更新请求数据"
// @Success 200 {object} res.Response{message=string}
// @Router /api/images [put]
func (ImagesApi) ImageUpdateView(c *gin.Context) {
	_claims, ok := c.Get("claims")
	if !ok {
		res.FailWithMessage("未登录", c)
		return
	}
	claims := _claims.(*jwts.CustomClaims)

	var cr ImageUpdateRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	var imageModel models.BannerModel
	err = global.DB.Take(&imageModel, cr.ID).Error
	if err != nil {
		res.FailWithMessage("文件不存在", c)
		return
	}
	if !canOperateImage(claims, imageModel) {
		res.FailWithMessage("只能修改自己上传的图片", c)
		return
	}
	err = global.DB.Model(&imageModel).Update("name", cr.Name).Error
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	res.OkWithMessage("图片名称修改成功", c)
	return

}
