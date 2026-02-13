package redis_ser

const CommentPrefix = "comment"

/*// Comment 评论某一篇文章
func Comment(id string) error {
	num, _ := global.Redis.HGet(CommentPrefix, id).Int()
	num++
	err := global.Redis.HSet(CommentPrefix, id, num).Err()
	return err
}

// GetComment 获取某一篇文章下的评论数
func GetComment(id string) int {
	num, _ := global.Redis.HGet(CommentPrefix, id).Int()
	return num
}

// GetCommentInfo 取出评论数据
func GetCommentInfo() map[string]int {
	var CommentInfo = map[string]int{}
	maps := global.Redis.HGetAll(CommentPrefix).Val()
	for id, val := range maps {
		num, _ := strconv.Atoi(val)
		CommentInfo[id] = num
	}
	return CommentInfo
}

// CommentClear 清空评论数据
func CommentClear() {
	global.Redis.Del(CommentPrefix)
}*/
