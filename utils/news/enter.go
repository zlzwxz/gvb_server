package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CategoryItem 表示一个热榜来源。
// 这里单独定义成具名结构体，是为了让上层 handler 更容易复用字段。
type CategoryItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Icon     string `json:"icon"`
	Category string `json:"category"`
}

// CategoryResponse 分类列表（拿最新ID）
type CategoryResponse struct {
	Code int            `json:"code"`
	Data []CategoryItem `json:"data"`
}

// 固定接口：获取所有平台 + 最新ID
const categoryAPI = "https://api.codelife.cc/api/top/category?lang=cn"

var categoryHTTPClient = &http.Client{Timeout: 3 * time.Second}

// GetNewsId 获取热搜榜类型的id
func GetNewsId() CategoryResponse {
	resp, err := categoryHTTPClient.Get(categoryAPI)
	if err != nil {
		fmt.Println(categoryAPI, err)
		return CategoryResponse{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		return CategoryResponse{}
	}

	var categoryResp CategoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&categoryResp); err != nil {
		fmt.Println(err)
		return CategoryResponse{}
	}
	return categoryResp
}
