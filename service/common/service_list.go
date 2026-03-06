package common

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/plugins/log_stash"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

type Option struct {
	models.PageInfo                 // 分页参数
	Debug           bool            // 是否开启调试模式
	Likes           []string        // 模糊查询的字段
	Level           log_stash.Level // 日志的等级
	Where           string          // 自定义 WHERE 条件
	WhereArgs       []interface{}   // WHERE 条件参数

}

func ComList[T any](model T, option Option) (list []T, count int64, err error) {
	fmt.Println("option.Limit:", option.Limit, "option.Page:", option.Page, "likes:", option.Likes)
	DB := global.DB
	if option.Debug {
		// 开启调试模式
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

	//  添加模糊查询支持
	if option.Key != "" {
		// 获取模型的所有字符串字段进行模糊查询
		query = addFuzzySearch(query, model, option.Key)
	}

	// 添加level筛选条件
	if option.Level != 0 {
		query = query.Where("level = ?", option.Level)
	}
	// 添加自定义 WHERE 条件
	if option.Where != "" {
		query = query.Where(option.Where, option.WhereArgs...)
	}
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

// addFuzzySearch 添加模糊查询条件
func addFuzzySearch(query *gorm.DB, model interface{}, keyword string) *gorm.DB {
	// 使用反射获取模型的字段信息
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 构建OR查询条件
	var conditions []string
	var args []interface{}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 只处理字符串类型的字段
		if field.Type.Kind() == reflect.String {
			// 检查是否有json标签
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				// 提取字段名（处理json标签中的选项）
				fieldName := strings.Split(jsonTag, ",")[0]
				if fieldName != "" {
					conditions = append(conditions, fmt.Sprintf("%s LIKE ?", fieldName))
					args = append(args, "%"+keyword+"%")
				}
			}
		}
	}

	// 如果有匹配的字段，添加OR查询
	if len(conditions) > 0 {
		whereClause := strings.Join(conditions, " OR ")
		query = query.Where(whereClause, args...)
	}

	return query
}
