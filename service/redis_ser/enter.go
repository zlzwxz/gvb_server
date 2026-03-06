package redis_ser

const (
	articleLookPrefix         = "article_look"
	articleCommentCountPrefix = "article_comment_count"
	articleDiggPrefix         = "article_digg"
	commentDiggPrefix         = "comment_digg"
)

// NewDigg 返回文章点赞计数器。
func NewDigg() CountDB {
	return CountDB{Index: articleDiggPrefix}
}

// NewArticleLook 返回文章浏览量计数器。
func NewArticleLook() CountDB {
	return CountDB{Index: articleLookPrefix}
}

// NewCommentCount 返回文章评论量计数器。
func NewCommentCount() CountDB {
	return CountDB{Index: articleCommentCountPrefix}
}

// NewCommentDigg 返回评论点赞计数器。
func NewCommentDigg() CountDB {
	return CountDB{Index: commentDiggPrefix}
}
