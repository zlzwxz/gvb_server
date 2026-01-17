package menu_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
)

type MenuNameResponse struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// MenuNameList 获取菜单名称列表
// @Summary 获取菜单名称列表
// @Description 获取所有菜单的基础信息（ID、标题、路径）
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Success 200 {object} res.Response{data=[]MenuNameResponse} "返回菜单名称列表"
// @Router /api/menus/names [get]
func (MenuApi) MenuNameList(c *gin.Context) {
	var menuNameList []MenuNameResponse
	global.DB.Model(models.MenuModel{}).Select("id", "title", "path").Scan(&menuNameList)
	res.OkWithData(menuNameList, c)
}
