package menu_api

import (
	"errors"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MenuUpdateView 修改菜单
// @Summary 更新菜单
// @Description 按请求体中的菜单ID更新菜单信息，包括轮播图关联
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param request body MenuRequest true "菜单更新参数"
// @Success 200 {object} res.Response{msg=string} "更新成功"
// @Router /api/menus [put]
func (MenuApi) MenuUpdateView(c *gin.Context) {
	var cr MenuRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}
	if cr.ID == 0 {
		res.FailWithMessage("菜单ID不能为空", c)
		return
	}

	normalizeMenuRequest(&cr)

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		var menuModel models.MenuModel
		if err := tx.Take(&menuModel, cr.ID).Error; err != nil {
			return err
		}

		var duplicateCount int64
		if err := tx.Model(&models.MenuModel{}).
			Where("(title = ? OR path = ?) AND id <> ?", cr.Title, cr.Path, cr.ID).
			Count(&duplicateCount).Error; err != nil {
			return err
		}
		if duplicateCount > 0 {
			return errors.New("菜单标题或者路径重复，修改失败")
		}

		if err := tx.Model(&menuModel).Updates(buildMenuUpdateMap(cr)).Error; err != nil {
			return err
		}

		// 先删后建是最容易理解的写法：旧关系全部清掉，再按照前端传来的新顺序重建。
		if err := tx.Where("menu_id = ?", menuModel.ID).Delete(&models.MenuBannerModel{}).Error; err != nil {
			return err
		}

		menuBannerList := buildMenuBannerModels(menuModel.ID, cr.ImageSortList)
		if len(menuBannerList) == 0 {
			return nil
		}

		return tx.Create(&menuBannerList).Error
	})
	if err != nil {
		if err.Error() == "菜单标题或者路径重复，修改失败" {
			res.FailWithMessage(err.Error(), c)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.FailWithMessage("菜单不存在", c)
			return
		}
		global.Log.Error(err)
		res.FailWithMessage("修改菜单失败", c)
		return
	}

	res.OkWithMessage("修改菜单成功", c)
}
