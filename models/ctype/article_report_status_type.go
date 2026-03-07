package ctype

type ArticleReportStatus int

const (
	ArticleReportPending   ArticleReportStatus = 1 // 待处理
	ArticleReportReReview  ArticleReportStatus = 2 // 已转复审
	ArticleReportDismissed ArticleReportStatus = 3 // 已忽略
)

func (s ArticleReportStatus) IsHandled() bool {
	return s == ArticleReportReReview || s == ArticleReportDismissed
}
