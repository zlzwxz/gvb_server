package user_api

import (
	"context"
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/service/es_ser"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
)

// UserRemoveView 批量删除用户
// @Summary 批量删除用户
// @Description 根据用户ID列表批量删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body models.RemoveRequest true "用户ID列表"
// @Success 200 {object} res.Response{msg=string} "删除成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权"
// @Failure 404 {object} res.Response "用户不存在"
// @Router /api/users [delete]
func (UserApi) UserRemoveView(c *gin.Context) {
	var cr models.RemoveRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var userList []models.UserModel
	count := global.DB.Find(&userList, cr.IDList).RowsAffected
	if count == 0 {
		res.FailWithMessage("用户不存在", c)
		return
	}

	// 事务
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		ids := make([]uint, 0, len(userList))
		for _, user := range userList {
			ids = append(ids, user.ID)
		}

		if len(ids) == 0 {
			return nil
		}

		if err = tx.Where("send_user_id IN ? OR rev_user_id IN ?", ids, ids).Delete(&models.MessageModel{}).Error; err != nil {
			global.Log.Error(err)
			return err
		}
		if err = tx.Where("user_id IN ?", ids).Delete(&models.CommentModel{}).Error; err != nil {
			global.Log.Error(err)
			return err
		}
		if err = tx.Where("user_id IN ?", ids).Delete(&models.UserCollectModel{}).Error; err != nil {
			global.Log.Error(err)
			return err
		}
		if err = tx.Where("user_id IN ?", ids).Delete(&models.LoginDataModel{}).Error; err != nil {
			global.Log.Error(err)
			return err
		}
		if err = tx.Where("user_id IN ?", ids).Delete(&models.UserCheckInModel{}).Error; err != nil {
			global.Log.Error(err)
			return err
		}
		if err = tx.Where("user_id IN ?", ids).Delete(&models.ArticleFileModel{}).Error; err != nil {
			global.Log.Error(err)
			return err
		}

		if err = removeUserArticlesFromES(ids); err != nil {
			global.Log.Error(err)
			return err
		}

		if err = tx.Delete(&userList).Error; err != nil {
			global.Log.Error(err)
			return err
		}
		return nil
	})
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("删除用户失败", c)
		return
	}
	res.OkWithMessage(fmt.Sprintf("共删除 %d 个用户", count), c)

}

func removeUserArticlesFromES(userIDs []uint) error {
	if len(userIDs) == 0 || global.ESClient == nil {
		return nil
	}

	userIDValues := make([]interface{}, 0, len(userIDs))
	for _, userID := range userIDs {
		userIDValues = append(userIDValues, int(userID))
	}

	searchResult, err := global.ESClient.Search(models.ArticleModel{}.Index()).
		Query(elastic.NewTermsQuery("user_id", userIDValues...)).
		Size(10000).
		Do(context.Background())
	if err != nil {
		return err
	}

	if searchResult.Hits == nil || len(searchResult.Hits.Hits) == 0 {
		return nil
	}

	articleIDs := make([]string, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		articleIDs = append(articleIDs, hit.Id)
	}

	bulk := global.ESClient.Bulk().Index(models.ArticleModel{}.Index())
	for _, articleID := range articleIDs {
		bulk.Add(elastic.NewBulkDeleteRequest().Id(articleID))
	}
	if _, err = bulk.Do(context.Background()); err != nil {
		return err
	}
	for _, articleID := range articleIDs {
		es_ser.DeleteFullTextByArticleID(articleID)
	}
	return nil
}
