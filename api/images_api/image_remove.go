package images_api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils/jwts"
)

// ImageRemoveView 删除图片
// @Tags 图片管理
// @Summary 删除图片
// @Description 根据ID列表删除对应的图片
// @Accept json
// @Produce json
// @Param data body models.RemoveRequest true "删除请求数据"
// @Success 200 {object} res.Response{}
// @Router /api/images [delete]
func (ImagesApi) ImageRemoveView(c *gin.Context) {
	_claims, ok := c.Get("claims")
	if !ok {
		res.FailWithMessage("未登录", c)
		return
	}
	claims := _claims.(*jwts.CustomClaims)

	var cr models.RemoveRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var imageList []models.BannerModel
	count := global.DB.Debug().Find(&imageList, cr.IDList).RowsAffected
	if count == 0 {
		res.FailWithMessage("文件不存在", c)
		return
	}

	if !isImageAdmin(claims) {
		ownImageList := make([]models.BannerModel, 0, len(imageList))
		for _, image := range imageList {
			if canOperateImage(claims, image) {
				ownImageList = append(ownImageList, image)
			}
		}
		if len(ownImageList) == 0 {
			res.FailWithMessage("只能删除自己上传的图片", c)
			return
		}
		global.DB.Delete(&ownImageList)
		res.OkWithMessage(fmt.Sprintf("共删除 %d 张图片", len(ownImageList)), c)
		return
	}

	global.DB.Delete(&imageList)
	res.OkWithMessage(fmt.Sprintf("共删除 %d 张图片", len(imageList)), c)

}
