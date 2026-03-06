package menu_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ImageSort 定义图片排序结构
// @Description 图片ID和排序值的组合
// 这里单独拆出来，是为了让前端可以明确告诉后端“第几张图排第几位”。
type ImageSort struct {
	ImageID uint `json:"image_id"`
	Sort    int  `json:"sort"`
}

// MenuRequest 定义菜单请求结构体
// @Description 菜单操作请求参数
// ID 只在更新时使用，创建时保持 0 即可。
type MenuRequest struct {
	ID            uint        `json:"id" structs:"-"`
	Title         string      `json:"title" binding:"required" msg:"请完善菜单名称" structs:"title"`
	Path          string      `json:"path" binding:"required" msg:"请完善菜单路径" structs:"path"`
	Slogan        string      `json:"slogan" structs:"slogan"`
	Abstract      ctype.Array `json:"abstract" structs:"abstract"`
	AbstractTime  int         `json:"abstract_time" structs:"abstract_time"`                // 切换的时间，单位秒
	BannerTime    int         `json:"banner_time" structs:"banner_time"`                    // 切换的时间，单位秒
	Sort          int         `json:"sort" binding:"required" msg:"请输入菜单序号" structs:"sort"` // 菜单的序号
	ImageSortList []ImageSort `json:"image_sort_list" structs:"-"`                          // 具体图片的顺序
}

// MenuCreateView 创建菜单
// @Summary 创建菜单
// @Description 创建新的菜单项，支持标题、路径、标语、摘要和轮播图配置
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param request body MenuRequest true "菜单创建参数"
// @Success 200 {object} res.Response{msg=string} "创建成功"
// @Router /api/menus [post]
func (MenuApi) MenuCreateView(c *gin.Context) {
	var cr MenuRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, &cr, c)
		return
	}

	normalizeMenuRequest(&cr)

	var duplicateCount int64
	if err := global.DB.Model(&models.MenuModel{}).
		Where("title = ? OR path = ?", cr.Title, cr.Path).
		Count(&duplicateCount).Error; err != nil {
		global.Log.Error(err)
		res.FailWithMessage("菜单添加失败", c)
		return
	}
	if duplicateCount > 0 {
		res.FailWithMessage("菜单标题或者路径重复，添加失败", c)
		return
	}

	menuModel := models.MenuModel{
		Title:        cr.Title,
		Path:         cr.Path,
		Slogan:       cr.Slogan,
		Abstract:     cr.Abstract,
		AbstractTime: cr.AbstractTime,
		BannerTime:   cr.BannerTime,
		Sort:         cr.Sort,
	}

	// 菜单主表和轮播关联表必须一起成功，才不会出现“菜单建好了但轮播没挂上”的半成功状态。
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&menuModel).Error; err != nil {
			return err
		}

		menuBannerList := buildMenuBannerModels(menuModel.ID, cr.ImageSortList)
		if len(menuBannerList) == 0 {
			return nil
		}

		return tx.Create(&menuBannerList).Error
	})
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("菜单添加失败", c)
		return
	}

	res.OkWithMessage("菜单添加成功", c)
}
