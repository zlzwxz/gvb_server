package ctype

type ArticleReviewStatus int

const (
	ArticleReviewLegacy   ArticleReviewStatus = 0 // 历史数据（默认视为已通过）
	ArticleReviewPending  ArticleReviewStatus = 1 // 待审核
	ArticleReviewApproved ArticleReviewStatus = 2 // 审核通过
	ArticleReviewRejected ArticleReviewStatus = 3 // 审核驳回
)

func (s ArticleReviewStatus) IsPublicVisible() bool {
	return s == ArticleReviewLegacy || s == ArticleReviewApproved
}
