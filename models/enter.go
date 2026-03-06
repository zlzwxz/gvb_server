package models

import "time"

// MODEL 提供所有数据模型都会复用的公共字段。
// 这里相当于一个“基础表结构”，业务模型只需要嵌入它即可。
type MODEL struct {
	ID        uint      `gorm:"primarykey" json:"id,select(c)"` // 主键 ID
	CreatedAt time.Time `json:"created_at"`                     // 创建时间
	UpdatedAt time.Time `json:"-"`                              // 更新时间
}

// RemoveRequest 统一接收批量删除时的 ID 列表。
type RemoveRequest struct {
	IDList []uint `json:"id_list"`
}

// PageInfo 统一接收常见的分页、排序和关键字搜索参数。
type PageInfo struct {
	Page  int    `form:"page"`
	Key   string `form:"key"`
	Limit int    `form:"limit"`
	Sort  string `form:"sort"`
	Tag   string `form:"tag"`
	Id    string `form:"id"`
}

// ESIDRequest 用于接收单个 ES 文档 ID。
type ESIDRequest struct {
	ID string `json:"id" binding:"required" uri:"id"`
}

// ESIDListRequest 用于接收多个 ES 文档 ID。
type ESIDListRequest struct {
	IDList []string `json:"id_list" binding:"required" form:"id_list"`
}
