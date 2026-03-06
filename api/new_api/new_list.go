package new_api

import (
	"encoding/json"
	"fmt"
	"gvb-server/models/res"
	"gvb-server/service/redis_ser"
	"gvb-server/utils/requests"
	"io"
	"time"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
)

// params 新闻请求参数
type params struct {
	ID   string `json:"id" form:"id" swag:"description:新闻分类ID"`
	Size int    `json:"size" form:"size" swag:"description:获取数量"`
}

// header 新闻请求头
// 第三方接口对请求头要求不高，这里保留结构体主要是为了以后扩展时不用改 handler 结构。
type header struct {
	Signaturekey string `form:"signaturekey" structs:"signaturekey" swag:"description:签名key"`
	Version      string `form:"version" structs:"version" swag:"description:版本号"`
	UserAgent    string `form:"User-Agent" structs:"User-Agent" swag:"description:用户代理"`
}

// NewResponse 新闻响应结构
type NewResponse struct {
	Code int                 `json:"code"`
	Data []redis_ser.NewData `json:"data"`
	Msg  string              `json:"msg"`
}

const newAPI = "https://api.codelife.cc/api/top/list?"
const timeout = 2 * time.Second

// NewListView 获取新闻列表
// @Summary 获取新闻列表
// @Description 获取热搜榜新闻列表
// @Tags 新闻管理
// @Accept json
// @Produce json
// @Param id query string false "新闻分类ID"
// @Param size query int false "获取数量"
// @Success 200 {object} res.Response{data=[]redis_ser.NewData} "获取成功"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 500 {object} res.Response "获取失败"
// @Router /api/news [get]
func (NewApi) NewListView(c *gin.Context) {
	var cr params
	var headers header

	if err := c.ShouldBindQuery(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	_ = c.ShouldBindHeader(&headers)

	if cr.Size <= 0 {
		cr.Size = 20
	}
	if cr.Size > 50 {
		cr.Size = 50
	}

	sources := filterEnabledNewsSources(getNewsSources())
	cr.ID = pickNewsSourceID(cr.ID, sources)
	if cr.ID == "" {
		res.FailWithMessage("暂无可用的新闻来源，请先在系统配置里开启资讯榜单", c)
		return
	}

	key := fmt.Sprintf("%s-%d", cr.ID, cr.Size)
	if newsData, _ := redis_ser.GetNews(key); len(newsData) != 0 {
		res.OkWithData(newsData, c)
		return
	}

	httpResponse, err := requests.Post(newAPI+fmt.Sprintf("id=%s&size=%d", cr.ID, cr.Size), cr, structs.Map(headers), timeout)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	defer httpResponse.Body.Close()

	var response NewResponse
	byteData, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	if err = json.Unmarshal(byteData, &response); err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}
	if response.Code != 200 {
		res.FailWithMessage(response.Msg, c)
		return
	}

	res.OkWithData(response.Data, c)
	_ = redis_ser.SetNews(key, response.Data)
}
