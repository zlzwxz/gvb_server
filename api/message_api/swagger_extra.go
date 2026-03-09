package message_api

import (
	"gvb-server/models"
	"gvb-server/models/res"
)

var (
	_ = res.Response{}
	_ = models.MessageModel{}
)

// MessageListAllAliasViewDoc 获取所有消息列表（兼容旧路径）。
// @Tags 消息管理
// @Summary 获取所有消息列表（兼容旧路径）
// @Description 兼容旧前端和第三方调用方的管理员消息列表入口，实际逻辑与 /api/messages/all 相同。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]models.MessageModel}} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/messages_all [get]
func MessageListAllAliasViewDoc() {}

// MessageRecordAliasViewDoc 获取用户消息记录（兼容旧路径）。
// @Tags 消息管理
// @Summary 获取用户消息记录（兼容旧路径）
// @Description 兼容旧前端和第三方调用方的私信记录入口，实际逻辑与 /api/messages/record 相同。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param user_id query int true "用户 ID"
// @Success 200 {object} res.Response{data=[]models.MessageModel} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/messages_record [get]
func MessageRecordAliasViewDoc() {}
