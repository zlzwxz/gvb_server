package redis_ser

import (
	"strconv"

	"gvb-server/global"
)

// CountDB 封装一类“按 ID 累加计数”的 Redis Hash 操作。
// 文章浏览量、文章点赞数、评论点赞数本质上都是这一类数据，所以统一抽成一个通用结构。
type CountDB struct {
	Index string // Redis Hash 的 key 前缀
}

// Set 对某个 ID 的计数执行 +1。
func (c CountDB) Set(id string) error {
	num, _ := global.Redis.HGet(c.Index, id).Int()
	num++
	return global.Redis.HSet(c.Index, id, num).Err()
}

// SetCount 对某个 ID 的计数执行自定义增量。
func (c CountDB) SetCount(id string, count int) error {
	num, _ := global.Redis.HGet(c.Index, id).Int()
	num += count
	return global.Redis.HSet(c.Index, id, num).Err()
}

// Get 读取单个 ID 当前的计数值。
func (c CountDB) Get(id string) int {
	num, _ := global.Redis.HGet(c.Index, id).Int()
	return num
}

// GetInfo 把整个 Hash 读成 map，方便批量汇总到数据库或 ES。
func (c CountDB) GetInfo() map[string]int {
	info := map[string]int{}
	maps := global.Redis.HGetAll(c.Index).Val()
	for id, val := range maps {
		num, _ := strconv.Atoi(val)
		info[id] = num
	}
	return info
}

// Clear 清空当前计数 Hash。
func (c CountDB) Clear() {
	global.Redis.Del(c.Index)
}
