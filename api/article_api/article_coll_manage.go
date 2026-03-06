package article_api

import (
	"context"
	"encoding/json"
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"
	"gvb-server/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type CollectManageQuery struct {
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
	Scope     string `form:"scope"`
	UserID    uint   `form:"user_id"`
	ArticleID string `form:"article_id"`
}

type CollectManageItem struct {
	UserID       uint   `json:"user_id"`
	UserName     string `json:"user_name"`
	NickName     string `json:"nick_name"`
	ArticleID    string `json:"article_id"`
	ArticleTitle string `json:"article_title"`
	CreatedAt    string `json:"created_at"`
}

type CollectManageBatchRemoveRequest struct {
	Items []CollectManageRemoveItem `json:"items" binding:"required"`
}

type CollectManageRemoveItem struct {
	UserID    uint   `json:"user_id" binding:"required"`
	ArticleID string `json:"article_id" binding:"required"`
}

// ArticleCollManageListView 后台收藏管理列表
// @Tags 文章管理
// @Summary 后台收藏管理列表
// @Description 获取收藏列表；管理员可通过 scope=all 查看全站，普通用户仅可查看自己的收藏
// @Param token header string true "token"
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Param scope query string false "all 或 me"
// @Param user_id query int false "用户ID（仅 scope=all 且管理员可用）"
// @Param article_id query string false "文章ID"
// @Router /api/articles/collects/manage [get]
// @Produce json
// @Success 200 {object} res.Response{data=object{count=int64,list=[]CollectManageItem}} "获取成功"
func (ArticleApi) ArticleCollManageListView(c *gin.Context) {
	var cr CollectManageQuery
	_ = c.ShouldBindQuery(&cr)

	if cr.Page < 1 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = 20
	}
	if cr.Limit > 100 {
		cr.Limit = 100
	}

	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	isAdmin := claims.Role == int(ctype.PermissionAdmin)

	query := global.DB.Model(&models.UserCollectModel{})

	if !isAdmin || strings.ToLower(cr.Scope) != "all" {
		query = query.Where("user_id = ?", claims.UserID)
	} else if cr.UserID > 0 {
		query = query.Where("user_id = ?", cr.UserID)
	}

	if cr.ArticleID != "" {
		query = query.Where("article_id = ?", cr.ArticleID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取收藏总数失败", c)
		return
	}

	var collects []models.UserCollectModel
	err := query.Order("created_at desc").Limit(cr.Limit).Offset((cr.Page - 1) * cr.Limit).Find(&collects).Error
	if err != nil {
		res.FailWithMessage("获取收藏列表失败", c)
		return
	}

	list := buildCollectManageItems(collects)
	res.OkWithList(list, count, c)
}

// ArticleCollManageBatchRemoveView 后台批量取消收藏
// @Tags 文章管理
// @Summary 后台批量取消收藏
// @Description 批量取消收藏；管理员可取消任意用户收藏，普通用户仅可取消自己的收藏
// @Param token header string true "token"
// @Param data body CollectManageBatchRemoveRequest true "收藏记录列表"
// @Router /api/articles/collects/manage [delete]
// @Produce json
// @Success 200 {object} res.Response{data=object{removed=int}} "取消成功"
func (ArticleApi) ArticleCollManageBatchRemoveView(c *gin.Context) {
	var cr CollectManageBatchRemoveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	isAdmin := claims.Role == int(ctype.PermissionAdmin)

	itemMap := map[string]CollectManageRemoveItem{}
	for _, item := range cr.Items {
		if item.ArticleID == "" {
			continue
		}
		if !isAdmin && item.UserID != claims.UserID {
			res.FailWithMessage("无权操作他人收藏", c)
			return
		}
		key := fmt.Sprintf("%d:%s", item.UserID, item.ArticleID)
		itemMap[key] = item
	}

	if len(itemMap) == 0 {
		res.FailWithMessage("请选择有效的收藏记录", c)
		return
	}

	var records []models.UserCollectModel
	for _, item := range itemMap {
		var coll models.UserCollectModel
		err := global.DB.Take(&coll, "user_id = ? and article_id = ?", item.UserID, item.ArticleID).Error
		if err == nil {
			records = append(records, coll)
		}
	}

	if len(records) == 0 {
		res.FailWithMessage("没有可取消的收藏记录", c)
		return
	}

	tx := global.DB.Begin()
	for _, record := range records {
		if err := tx.Delete(&models.UserCollectModel{}, "user_id = ? and article_id = ?", record.UserID, record.ArticleID).Error; err != nil {
			tx.Rollback()
			res.FailWithMessage("取消收藏失败", c)
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		res.FailWithMessage("取消收藏失败", c)
		return
	}

	articleDelta := map[string]int{}
	for _, record := range records {
		articleDelta[record.ArticleID]++
	}

	for articleID, reduceCount := range articleDelta {
		article, err := es_ser.CommDetail(articleID)
		if err != nil {
			continue
		}
		newCount := article.CollectsCount - reduceCount
		if newCount < 0 {
			newCount = 0
		}
		_ = es_ser.ArticleUpdate(articleID, map[string]any{
			"collects_count": newCount,
		})
	}

	res.OkWithData(map[string]any{
		"removed": len(records),
	}, c)
}

func buildCollectManageItems(collects []models.UserCollectModel) []CollectManageItem {
	if len(collects) == 0 {
		return []CollectManageItem{}
	}

	userIDMap := map[uint]struct{}{}
	articleIDMap := map[string]struct{}{}
	for _, coll := range collects {
		userIDMap[coll.UserID] = struct{}{}
		articleIDMap[coll.ArticleID] = struct{}{}
	}

	var userIDs []uint
	for userID := range userIDMap {
		userIDs = append(userIDs, userID)
	}

	var articleIDs []string
	for articleID := range articleIDMap {
		articleIDs = append(articleIDs, articleID)
	}

	userMap := map[uint]models.UserModel{}
	if len(userIDs) > 0 {
		var users []models.UserModel
		err := global.DB.Select("id", "user_name", "nick_name").Find(&users, "id in ?", userIDs).Error
		if err == nil {
			for _, user := range users {
				userMap[user.ID] = user
			}
		}
	}

	articleTitleMap := getArticleTitleMap(articleIDs)

	list := make([]CollectManageItem, 0, len(collects))
	for _, coll := range collects {
		user := userMap[coll.UserID]
		list = append(list, CollectManageItem{
			UserID:       coll.UserID,
			UserName:     user.UserName,
			NickName:     user.NickName,
			ArticleID:    coll.ArticleID,
			ArticleTitle: articleTitleMap[coll.ArticleID],
			CreatedAt:    coll.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return list
}

func getArticleTitleMap(articleIDs []string) map[string]string {
	titleMap := map[string]string{}
	if len(articleIDs) == 0 {
		return titleMap
	}

	idList := make([]interface{}, 0, len(articleIDs))
	for _, articleID := range articleIDs {
		idList = append(idList, articleID)
	}

	result, err := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(elastic.NewTermsQuery("_id", idList...)).
		Size(len(articleIDs)).
		Do(context.Background())
	if err != nil {
		return titleMap
	}

	for _, hit := range result.Hits.Hits {
		var article models.ArticleModel
		if err = json.Unmarshal(hit.Source, &article); err != nil {
			continue
		}
		titleMap[hit.Id] = article.Title
	}

	return titleMap
}
