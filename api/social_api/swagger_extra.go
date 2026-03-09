package social_api

import "gvb-server/models/res"

var _ = res.Response{}

type socialBlockItemDoc struct {
	ID        uint   `json:"id" example:"1"`
	UserID    uint   `json:"user_id" example:"2"`
	NickName  string `json:"nick_name" example:"被拉黑用户"`
	UserName  string `json:"user_name" example:"blocked_user"`
	Avatar    string `json:"avatar" example:"/uploads/avatar/blocked.png"`
	Reason    string `json:"reason" example:"频繁骚扰"`
	CreatedAt string `json:"created_at" example:"2026-03-08 10:30:00"`
}

type socialGroupOverviewItemDoc struct {
	ID              uint   `json:"id" example:"1"`
	GroupNo         string `json:"group_no" example:"G000001"`
	Name            string `json:"name" example:"产品讨论群"`
	Avatar          string `json:"avatar" example:"/uploads/group/default.png"`
	Notice          string `json:"notice" example:"仅讨论产品需求和 Bug"`
	OwnerUserID     uint   `json:"owner_user_id" example:"1"`
	MemberCount     int    `json:"member_count" example:"3"`
	ConversationKey string `json:"conversation_key" example:"group:1"`
	MemberIDs       []uint `json:"member_ids" example:"1,2,3"`
	CreatedAt       string `json:"created_at,omitempty" example:"2026-03-08T10:00:00+08:00"`
}

type socialJoinGroupResponseDoc struct {
	ID              uint   `json:"id" example:"1"`
	GroupNo         string `json:"group_no" example:"G000001"`
	Name            string `json:"name" example:"产品讨论群"`
	ConversationKey string `json:"conversation_key" example:"group:1"`
}

type socialDiscoveryResponseDoc struct {
	Users  []socialDiscoveryUserItem  `json:"users"`
	Groups []socialDiscoveryGroupItem `json:"groups"`
}

type socialGroupDetailResponseDoc struct {
	ID              uint                    `json:"id" example:"1"`
	GroupNo         string                  `json:"group_no" example:"G000001"`
	Name            string                  `json:"name" example:"产品讨论群"`
	Avatar          string                  `json:"avatar" example:"/uploads/group/default.png"`
	Notice          string                  `json:"notice" example:"仅讨论产品需求和 Bug"`
	OwnerUserID     uint                    `json:"owner_user_id" example:"1"`
	ViewerRole      string                  `json:"viewer_role" example:"owner"`
	ViewerRoleLabel string                  `json:"viewer_role_label" example:"群主"`
	CanManage       bool                    `json:"can_manage" example:"true"`
	Members         []socialGroupMemberItem `json:"members"`
	ConversationKey string                  `json:"conversation_key" example:"group:1"`
}

type socialGroupMembersResponseDoc struct {
	Members []socialGroupMemberItem `json:"members"`
}

// AdminSummaryViewDoc 获取后台社交总览。
// @Tags 社交系统
// @Summary 获取社交后台总览
// @Description 返回关注数、双向好友数、黑名单数、群组数、群成员数、在线人数等指标，适合作为后台面板顶部统计卡片数据。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=socialManageOverviewResponse} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/manage/summary [get]
func AdminSummaryViewDoc() {}

// AdminFollowListViewDoc 获取后台关注和好友关系列表。
// @Tags 社交系统
// @Summary 获取社交关系列表
// @Description mode=friend 返回双向好友，mode=follow 返回单向关注；支持分页和关键字搜索。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量，最大 100"
// @Param key query string false "按用户昵称、用户名、用户 ID 搜索"
// @Param mode query string false "关系模式：friend 或 follow"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]socialManageFollowItem}} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/manage/follows [get]
func AdminFollowListViewDoc() {}

// AdminBlockListViewDoc 获取后台黑名单列表。
// @Tags 社交系统
// @Summary 获取黑名单列表
// @Description 支持分页和关键字搜索，便于后台审查拉黑关系和拉黑原因。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量，最大 100"
// @Param key query string false "按用户昵称、用户名、用户 ID、原因搜索"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]socialManageBlockItem}} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/manage/blocks [get]
func AdminBlockListViewDoc() {}

// AdminGroupListViewDoc 获取后台群组列表。
// @Tags 社交系统
// @Summary 获取后台群组列表
// @Description 支持分页和关键字搜索，可查看群号、群主、成员数量和最后更新时间。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param page query int false "页码"
// @Param limit query int false "每页数量，最大 100"
// @Param key query string false "按群名、群号、群主信息搜索"
// @Success 200 {object} res.Response{data=object{count=int64,list=[]socialManageGroupItem}} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/manage/groups [get]
func AdminGroupListViewDoc() {}

// SummaryViewDoc 获取当前用户社交摘要。
// @Tags 社交系统
// @Summary 获取社交摘要
// @Description 返回当前用户在线状态、好友数、在线好友数和黑名单数量。好友面板初始化时建议先调这个接口。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=socialSummaryResponse} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/summary [get]
func SummaryViewDoc() {}

// DiscoveryViewDoc 搜索用户和群组。
// @Tags 社交系统
// @Summary 搜索用户和群组
// @Description 根据博客号、昵称、用户名或群组号进行搜索，返回用户关系状态和群组加入状态。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param key query string true "关键词，可输入博客号、昵称、用户名或群组号"
// @Success 200 {object} res.Response{data=socialDiscoveryResponseDoc} "搜索成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/discovery [get]
func DiscoveryViewDoc() {}

// FriendListViewDoc 获取好友列表。
// @Tags 社交系统
// @Summary 获取好友列表
// @Description 返回当前用户的双向好友列表，并附带在线状态、最近消息和未读数。可选 key 做前端搜索过滤。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param key query string false "按好友昵称或用户名过滤"
// @Success 200 {object} res.Response{data=[]socialFollowCard} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/friends [get]
func FriendListViewDoc() {}

// RelationViewDoc 获取与指定用户的社交关系。
// @Tags 社交系统
// @Summary 获取社交关系
// @Description 返回是否已关注、是否互相关注、是否好友、是否存在拉黑关系，以及能否私聊、建群、发文件和语音通话。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "目标用户 ID"
// @Success 200 {object} res.Response{data=socialRelationResponse} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/relations/{id} [get]
func RelationViewDoc() {}

// FollowViewDoc 关注用户。
// @Tags 社交系统
// @Summary 关注用户
// @Description 对目标用户建立单向关注。若双方互相关注且不存在拉黑关系，则会自动变成好友关系。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "目标用户 ID"
// @Success 200 {object} res.Response{data=socialRelationResponse} "关注成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/follows/{id} [post]
func FollowViewDoc() {}

// UnfollowViewDoc 取消关注用户。
// @Tags 社交系统
// @Summary 取消关注用户
// @Description 取消当前用户对目标用户的关注关系；如果原本是好友，也会同步失去好友关系。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "目标用户 ID"
// @Success 200 {object} res.Response{data=socialRelationResponse} "取消成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/follows/{id} [delete]
func UnfollowViewDoc() {}

// BlockListViewDoc 获取黑名单。
// @Tags 社交系统
// @Summary 获取黑名单
// @Description 返回当前用户拉黑的用户列表和拉黑原因。被拉黑用户之间无法继续私聊、关注或发起通话。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=[]socialBlockItemDoc} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/blocks [get]
func BlockListViewDoc() {}

// BlockViewDoc 拉黑用户。
// @Tags 社交系统
// @Summary 拉黑用户
// @Description 建立拉黑关系后，会自动清理双方之间的关注关系，后续也不能继续私聊和通话。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "目标用户 ID"
// @Param data body socialBlockRequest false "拉黑原因"
// @Success 200 {object} res.Response{data=socialRelationResponse} "拉黑成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/blocks/{id} [post]
func BlockViewDoc() {}

// UnblockViewDoc 取消拉黑用户。
// @Tags 社交系统
// @Summary 取消拉黑用户
// @Description 移除黑名单关系后，双方仍需要重新关注才能恢复好友身份。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "目标用户 ID"
// @Success 200 {object} res.Response{data=socialRelationResponse} "取消成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/blocks/{id} [delete]
func UnblockViewDoc() {}

// ConversationListViewDoc 获取会话列表。
// @Tags 社交系统
// @Summary 获取会话列表
// @Description 合并返回单聊和群聊会话，包含最新消息预览、未读数、在线状态、群成员数量等字段。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=[]socialConversationItem} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/conversations [get]
func ConversationListViewDoc() {}

// DirectMessageListViewDoc 获取单聊消息列表。
// @Tags 社交系统
// @Summary 获取单聊消息列表
// @Description 按 user_id 获取当前用户与目标用户的完整单聊消息，并自动推进当前用户的已读游标。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param user_id query int true "对方用户 ID"
// @Success 200 {object} res.Response{data=[]socialMessageResponse} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/direct/messages [get]
func DirectMessageListViewDoc() {}

// DirectMessageCreateViewDoc 发送单聊消息。
// @Tags 社交系统
// @Summary 发送单聊消息
// @Description msg_type 为空时按文本消息处理；文件消息需要先通过社交文件上传接口拿到 file_id，而且双方必须是好友。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body socialDirectMessageRequest true "消息内容"
// @Success 200 {object} res.Response{data=socialMessageResponse} "发送成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/direct/messages [post]
func DirectMessageCreateViewDoc() {}

// MessageSearchViewDoc 搜索会话消息。
// @Tags 社交系统
// @Summary 搜索会话消息
// @Description 二选一传 user_id 或 group_id。只搜索当前用户有权访问的会话，已撤回消息不会出现在结果中。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param user_id query int false "单聊目标用户 ID"
// @Param group_id query int false "群组 ID"
// @Param keyword query string true "搜索关键词"
// @Param limit query int false "返回条数，默认 20，最大 100"
// @Success 200 {object} res.Response{data=[]socialMessageResponse} "搜索成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/messages/search [get]
func MessageSearchViewDoc() {}

// MessageRecallViewDoc 撤回消息。
// @Tags 社交系统
// @Summary 撤回消息
// @Description 发送人可撤回自己的消息；群主和群管理员可撤回群内消息；平台管理员可全局撤回。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "消息 ID"
// @Success 200 {object} res.Response{data=socialMessageResponse} "撤回成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/messages/{id}/recall [post]
func MessageRecallViewDoc() {}

// CallLogListViewDoc 获取通话记录。
// @Tags 社交系统
// @Summary 获取通话记录
// @Description 返回当前用户的语音通话记录，可按状态筛选，默认最多返回 30 条。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param status query string false "通话状态：ringing、rejected、missed、completed、canceled"
// @Param limit query int false "返回条数，默认 30，最大 100"
// @Success 200 {object} res.Response{data=[]socialCallLogItem} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/calls [get]
func CallLogListViewDoc() {}

// GroupListViewDoc 获取我的群组列表。
// @Tags 社交系统
// @Summary 获取我的群组列表
// @Description 返回当前用户加入的群组信息、成员数量和会话 key，适合群聊列表页初始化。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Success 200 {object} res.Response{data=[]socialGroupOverviewItemDoc} "获取成功"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups [get]
func GroupListViewDoc() {}

// GroupCreateViewDoc 创建群组。
// @Tags 社交系统
// @Summary 创建群组
// @Description 只能邀请自己的好友入群，群总人数上限 30。创建成功后会自动把群主自己加入成员列表。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body socialGroupCreateRequest true "群组信息"
// @Success 200 {object} res.Response{data=socialGroupOverviewItemDoc} "创建成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups [post]
func GroupCreateViewDoc() {}

// GroupJoinViewDoc 通过群号加群。
// @Tags 社交系统
// @Summary 通过群号加入群组
// @Description 根据 group_no 加入目标群组。若已经在群里、群不存在或群人数已达上限，则会返回失败。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body socialGroupJoinRequest true "群号"
// @Success 200 {object} res.Response{data=socialJoinGroupResponseDoc} "加入成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups/join [post]
func GroupJoinViewDoc() {}

// GroupDetailViewDoc 获取群组详情。
// @Tags 社交系统
// @Summary 获取群组详情
// @Description 返回群基本信息、当前用户在群内的角色和完整成员列表。只有已入群成员才可查看。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "群组 ID"
// @Success 200 {object} res.Response{data=socialGroupDetailResponseDoc} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups/{id} [get]
func GroupDetailViewDoc() {}

// GroupMessageListViewDoc 获取群消息。
// @Tags 社交系统
// @Summary 获取群消息
// @Description 返回指定群组的消息列表，并自动把当前用户已读游标推进到最后一条消息。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "群组 ID"
// @Success 200 {object} res.Response{data=[]socialMessageResponse} "获取成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups/{id}/messages [get]
func GroupMessageListViewDoc() {}

// GroupMessageCreateViewDoc 发送群消息。
// @Tags 社交系统
// @Summary 发送群消息
// @Description 支持文本和文件消息。文件消息需要先上传到社交文件接口拿到 file_id。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "群组 ID"
// @Param data body socialGroupMessageRequest true "消息内容"
// @Success 200 {object} res.Response{data=socialMessageResponse} "发送成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups/{id}/messages [post]
func GroupMessageCreateViewDoc() {}

// GroupMemberAddViewDoc 邀请好友入群。
// @Tags 社交系统
// @Summary 邀请好友入群
// @Description 只有群主和管理员可以操作，并且仅允许邀请自己的好友入群。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "群组 ID"
// @Param data body socialGroupMemberSaveRequest true "要加入的成员 ID 列表"
// @Success 200 {object} res.Response{data=socialGroupMembersResponseDoc} "操作成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups/{id}/members [post]
func GroupMemberAddViewDoc() {}

// GroupMemberRoleUpdateViewDoc 设置群管理员。
// @Tags 社交系统
// @Summary 设置群成员角色
// @Description 只有群主可以把成员设置为 admin 或恢复为 member，不能直接修改群主自己的角色。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "群组 ID"
// @Param user_id path int true "目标用户 ID"
// @Param data body socialGroupMemberRoleRequest true "角色信息"
// @Success 200 {object} res.Response{data=socialGroupMembersResponseDoc} "设置成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups/{id}/members/{user_id}/role [put]
func GroupMemberRoleUpdateViewDoc() {}

// GroupMemberRemoveViewDoc 移除群成员。
// @Tags 社交系统
// @Summary 移除群成员或退群
// @Description 自己删除自己表示退群；管理员和群主可按权限移除其他成员；群主退出前必须先转让群主。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "群组 ID"
// @Param user_id path int true "目标用户 ID"
// @Success 200 {object} res.Response{data=socialGroupMembersResponseDoc} "移除成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups/{id}/members/{user_id} [delete]
func GroupMemberRemoveViewDoc() {}

// GroupTransferOwnerViewDoc 转让群主。
// @Tags 社交系统
// @Summary 转让群主
// @Description 只有当前群主可以把群主身份转给另一位现有群成员。转让成功后原群主会被降为管理员。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param id path int true "群组 ID"
// @Param data body socialGroupTransferRequest true "新群主用户 ID"
// @Success 200 {object} res.Response{data=socialGroupMembersResponseDoc} "转让成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/groups/{id}/transfer-owner [put]
func GroupTransferOwnerViewDoc() {}

// PresenceUpdateViewDoc 更新在线状态。
// @Tags 社交系统
// @Summary 更新在线状态
// @Description 更新当前用户的在线模式、状态文案和隐身设置。前端切换忙碌、离开、隐身等状态时调用。
// @Accept json
// @Produce json
// @Param token header string true "token"
// @Param data body socialPresenceRequest true "在线状态"
// @Success 200 {object} res.Response{data=socialPresenceResponse} "更新成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/presence [put]
func PresenceUpdateViewDoc() {}

// FileUploadViewDoc 上传社交文件。
// @Tags 社交系统
// @Summary 上传社交文件
// @Description 上传单聊或群聊中要发送的文件。返回 file_id 后，再在消息接口里用该 ID 发送文件消息。
// @Accept multipart/form-data
// @Produce json
// @Param token header string true "token"
// @Param file formData file true "文件内容"
// @Success 200 {object} res.Response{data=socialFileUploadItem} "上传成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/files [post]
func FileUploadViewDoc() {}

// FileDownloadViewDoc 下载社交文件。
// @Tags 社交系统
// @Summary 下载社交文件
// @Description 只有文件所有者、参与对应私聊会话的成员，或所在群成员，才能下载该文件。
// @Accept json
// @Produce application/octet-stream
// @Param token header string true "token"
// @Param id path int true "文件 ID"
// @Success 200 {file} binary "文件流"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/social/files/{id}/download [get]
func FileDownloadViewDoc() {}

// SocketViewDoc 建立社交 WebSocket 连接。
// @Tags 社交系统
// @Summary 建立社交 WebSocket 连接
// @Description 用于在线状态同步、实时私聊提醒、群消息通知和语音通话信令。token 可通过 query、token 头或 Authorization: Bearer 传入。
// @Accept json
// @Produce json
// @Param token query string false "token，也可放在 header"
// @Param token header string false "自定义 token 头"
// @Param Authorization header string false "Bearer token"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} res.Response "连接失败"
// @Router /api/social/ws [get]
func SocketViewDoc() {}
