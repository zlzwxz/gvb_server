package common

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"gorm.io/gorm"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/plugins/log_stash"
)

// Option 描述通用列表查询需要的筛选条件。
// 很多后台分页接口都复用了这套参数，避免每个模块都重复写分页和模糊搜索逻辑。
type Option struct {
	models.PageInfo
	Debug     bool            // 是否启用 SQL 调试日志
	Likes     []string        // 预留的模糊查询字段列表
	Level     log_stash.Level // 日志等级筛选
	Where     string          // 额外的 WHERE 条件
	WhereArgs []interface{}   // 额外 WHERE 对应的参数
}

// ComList 提供一个通用的 Gorm 列表查询模板。
// 它负责统一处理分页、排序、关键字搜索和计数，让各模块 API 只需要传模型和筛选参数。
func ComList[T any](model T, option Option) (list []T, count int64, err error) {
	DB := global.DB
	if option.Debug {
		DB = global.DB.Session(&gorm.Session{Logger: global.MysqlLog})
	}
	option.Sort = sanitizeSortClause(option.Sort, model, "created_at desc")
	if option.Page < 1 {
		option.Page = 1
	}
	if option.Limit < 0 {
		option.Limit = 0
	}

	query := DB.Model(&model)
	if option.Key != "" {
		query = addFuzzySearch(query, model, option.Key)
	}
	if option.Level != 0 {
		query = query.Where("level = ?", option.Level)
	}
	if option.Where != "" {
		query = query.Where(option.Where, option.WhereArgs...)
	}
	if err = query.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	offset := (option.Page - 1) * option.Limit
	if option.Limit > 0 {
		query = query.Limit(option.Limit).Offset(offset)
	}

	err = query.Order(option.Sort).Find(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}
	return list, count, err
}

// addFuzzySearch 根据模型里的字符串字段自动拼接 LIKE 查询。
// 对新手来说，可以把它理解成“自动遍历字符串字段，生成 title LIKE ? OR content LIKE ? 这一类条件”。
func addFuzzySearch(query *gorm.DB, model interface{}, keyword string) *gorm.DB {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	var conditions []string
	var args []interface{}
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.Type.Kind() != reflect.String {
			continue
		}
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", fieldName))
		args = append(args, "%"+keyword+"%")
	}

	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " OR "), args...)
	}
	return query
}

var sortClauseReg = regexp.MustCompile(`^([a-zA-Z0-9_]+)(?:\s+(asc|desc))?$`)

func sanitizeSortClause(rawSort string, model any, fallback string) string {
	allowedFieldSet := collectSortableFieldSet(model)
	if len(allowedFieldSet) == 0 {
		return fallback
	}

	rawSort = strings.TrimSpace(strings.ToLower(rawSort))
	if rawSort == "" {
		return fallback
	}

	safeClauses := make([]string, 0, 2)
	for _, part := range strings.Split(rawSort, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		matches := sortClauseReg.FindStringSubmatch(part)
		if len(matches) == 0 {
			continue
		}

		field := strings.TrimSpace(matches[1])
		if _, ok := allowedFieldSet[field]; !ok {
			continue
		}

		direction := strings.TrimSpace(matches[2])
		if direction != "asc" {
			direction = "desc"
		}

		safeClauses = append(safeClauses, fmt.Sprintf("%s %s", field, direction))
	}

	if len(safeClauses) == 0 {
		return fallback
	}
	return strings.Join(safeClauses, ", ")
}

func collectSortableFieldSet(model any) map[string]struct{} {
	fieldSet := map[string]struct{}{
		"id":         {},
		"created_at": {},
		"updated_at": {},
	}

	modelType := reflect.TypeOf(model)
	if modelType == nil {
		return fieldSet
	}
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return fieldSet
	}

	walkSortableFields(modelType, fieldSet)
	return fieldSet
}

func walkSortableFields(modelType reflect.Type, fieldSet map[string]struct{}) {
	for index := 0; index < modelType.NumField(); index++ {
		field := modelType.Field(index)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			walkSortableFields(field.Type, fieldSet)
			continue
		}

		if field.PkgPath != "" {
			continue
		}

		fieldName := resolveSortableFieldName(field)
		if fieldName == "" {
			continue
		}
		fieldSet[fieldName] = struct{}{}
	}
}

func resolveSortableFieldName(field reflect.StructField) string {
	if jsonTag := strings.TrimSpace(field.Tag.Get("json")); jsonTag != "" && jsonTag != "-" {
		name := strings.TrimSpace(strings.Split(jsonTag, ",")[0])
		if name != "" {
			return name
		}
	}

	gormTag := strings.TrimSpace(field.Tag.Get("gorm"))
	if gormTag != "" {
		for _, item := range strings.Split(gormTag, ";") {
			item = strings.TrimSpace(item)
			if strings.HasPrefix(strings.ToLower(item), "column:") {
				name := strings.TrimSpace(item[len("column:"):])
				if name != "" {
					return strings.ToLower(name)
				}
			}
		}
	}

	return toSnakeCase(field.Name)
}

func toSnakeCase(value string) string {
	if value == "" {
		return ""
	}
	var builder strings.Builder
	runes := []rune(value)
	for index, currentRune := range runes {
		if unicode.IsUpper(currentRune) {
			if index > 0 {
				prevRune := runes[index-1]
				if unicode.IsLower(prevRune) || unicode.IsDigit(prevRune) {
					builder.WriteRune('_')
				}
			}
			builder.WriteRune(unicode.ToLower(currentRune))
			continue
		}
		builder.WriteRune(unicode.ToLower(currentRune))
	}
	return builder.String()
}
