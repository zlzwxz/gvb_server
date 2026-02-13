package news

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 分类列表（拿最新ID）
type CategoryResponse struct {
	Code int `json:"code"`
	Data []struct {
		Id       string `json:"id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Icon     string `json:"icon"`
		Category string `json:"category"`
	} `json:"data"`
}

// 固定接口：获取所有平台 + 最新ID
const categoryAPI = "https://api.codelife.cc/api/top/category?lang=cn"
const hotAPI = "https://api.codelife.cc/api/top/list"

// GetNewsId 获取热搜榜类型的id
func GetNewsId() CategoryResponse {
	// 发送 GET 请求
	resp, err := http.Get(categoryAPI)
	if err != nil {
		fmt.Println(categoryAPI, err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
	}

	// 解析 JSON 响应
	var categoryResp CategoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&categoryResp); err != nil {
		fmt.Println(err)
	}
	return categoryResp
}
