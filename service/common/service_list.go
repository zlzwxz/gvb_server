package common

import (
	"fmt"
	"gorm.io/gorm"
	"gvb-server/global"
	"gvb-server/models"
)

type Option struct {
	models.PageInfo
	Debug bool
}

func ComList[T any](model T, option Option) (list []T, count int64, err error) {
	fmt.Println("option.Limit:", option.Limit, "option.Page:", option.Page)
	DB := global.DB
	if option.Debug {
		DB = global.DB.Session(&gorm.Session{Logger: global.MysqlLog})
	}
	if option.Sort == "" {
		option.Sort = "created_at desc" // 默认按照时间往前排
	}

	// 1. 处理分页参数的默认值和边界值
	// 页码最小为1
	if option.Page < 1 {
		option.Page = 1
	}
	// 每页条数：0 表示不分页（查询所有），否则最小为1
	if option.Limit < 0 {
		option.Limit = 0
	}

	// 2. 构建查询链（先计数，再查询列表，保证筛选条件一致）
	query := DB.Model(&model) // 基于传入的 model 构建查询

	// 计数（如果有 WHERE 条件，要先加条件再 Count）
	if err = query.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 3. 处理分页逻辑
	offset := (option.Page - 1) * option.Limit
	fmt.Println("option.Limit:", option.Limit, "option.offset:", offset)

	// 只有当 Limit > 0 时，才添加 LIMIT 和 OFFSET
	if option.Limit > 0 {
		query = query.Limit(option.Limit).Offset(offset)
	}

	// 4. 执行查询
	err = query.Order(option.Sort).Debug().Find(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}

	return list, count, err
}
