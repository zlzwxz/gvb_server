package article_api

import (
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/service/board_ser"
	"gvb-server/service/redis_ser"
	"gvb-server/utils/jwts"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

func optionalClaims(c *gin.Context) *jwts.CustomClaims {
	if _claims, ok := c.Get("claims"); ok {
		if claims, ok2 := _claims.(*jwts.CustomClaims); ok2 {
			return claims
		}
	}

	token := strings.TrimSpace(c.GetHeader("token"))
	if token == "" {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
	}
	if token == "" {
		return nil
	}

	claims, err := jwts.ParseToken(token)
	if err != nil || redis_ser.CheckLogout(token) {
		return nil
	}
	return claims
}

func isAdmin(claims *jwts.CustomClaims) bool {
	return claims != nil && claims.Role == int(ctype.PermissionAdmin)
}

func canManageArticle(article models.ArticleModel, claims *jwts.CustomClaims) bool {
	if claims == nil {
		return false
	}
	if isAdmin(claims) {
		return true
	}
	if article.UserID == claims.UserID {
		return true
	}
	if article.BoardID > 0 {
		if board, err := board_ser.GetBoardByID(article.BoardID); err == nil {
			if board_ser.IsUserBoardManager(board, claims.UserID) {
				return true
			}
		}
	}
	return false
}

func canReviewArticle(article models.ArticleModel, claims *jwts.CustomClaims) bool {
	if claims == nil {
		return false
	}
	if isAdmin(claims) {
		return true
	}
	if article.BoardID == 0 {
		return false
	}
	board, err := board_ser.GetBoardByID(article.BoardID)
	if err != nil {
		return false
	}
	return board_ser.IsUserBoardManager(board, claims.UserID)
}

func canViewArticle(article models.ArticleModel, claims *jwts.CustomClaims) bool {
	if article.ReviewStatus.IsPublicVisible() && !article.IsPrivate {
		return true
	}
	return canManageArticle(article, claims)
}

func publicVisibleQuery() elastic.Query {
	reviewQuery := elastic.NewBoolQuery().
		Should(
			elastic.NewTermQuery("review_status", int(ctype.ArticleReviewApproved)),
			elastic.NewTermQuery("review_status", int(ctype.ArticleReviewLegacy)),
			elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("review_status")),
		).
		MinimumNumberShouldMatch(1)
	privateQuery := elastic.NewBoolQuery().
		Should(
			elastic.NewTermQuery("is_private", false),
			elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("is_private")),
		).
		MinimumNumberShouldMatch(1)
	return elastic.NewBoolQuery().Must(reviewQuery, privateQuery)
}
