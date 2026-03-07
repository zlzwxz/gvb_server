package social_api

import (
	"strings"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type socialManageQuery struct {
	models.PageInfo
	Mode string `form:"mode"`
}

type socialManageOverviewResponse struct {
	FollowCount       int `json:"follow_count"`
	MutualFriendCount int `json:"mutual_friend_count"`
	BlockCount        int `json:"block_count"`
	GroupCount        int `json:"group_count"`
	GroupMemberCount  int `json:"group_member_count"`
	OnlineUserCount   int `json:"online_user_count"`
	PresenceUserCount int `json:"presence_user_count"`
}

type socialManageFollowItem struct {
	ID               uint   `json:"id"`
	UserID           uint   `json:"user_id"`
	UserNickName     string `json:"user_nick_name"`
	UserName         string `json:"user_name"`
	UserAvatar       string `json:"user_avatar"`
	FollowUserID     uint   `json:"follow_user_id"`
	FollowNickName   string `json:"follow_nick_name"`
	FollowUserName   string `json:"follow_user_name"`
	FollowUserAvatar string `json:"follow_user_avatar"`
	IsFriend         bool   `json:"is_friend"`
	CreatedAt        string `json:"created_at"`
}

type socialManageBlockItem struct {
	ID              uint   `json:"id"`
	UserID          uint   `json:"user_id"`
	UserNickName    string `json:"user_nick_name"`
	UserName        string `json:"user_name"`
	UserAvatar      string `json:"user_avatar"`
	BlockUserID     uint   `json:"block_user_id"`
	BlockNickName   string `json:"block_nick_name"`
	BlockUserName   string `json:"block_user_name"`
	BlockUserAvatar string `json:"block_user_avatar"`
	Reason          string `json:"reason"`
	CreatedAt       string `json:"created_at"`
}

type socialManageGroupItem struct {
	ID          uint   `json:"id"`
	GroupNo     string `json:"group_no"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Notice      string `json:"notice"`
	OwnerUserID uint   `json:"owner_user_id"`
	OwnerName   string `json:"owner_name"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func sanitizeManagePageInfo(cr *socialManageQuery, defaultLimit, maxLimit int) {
	if cr.Page <= 0 {
		cr.Page = 1
	}
	if cr.Limit <= 0 {
		cr.Limit = defaultLimit
	}
	if cr.Limit > maxLimit {
		cr.Limit = maxLimit
	}
}

func applySocialManageKeyFilter(query *gorm.DB, key string, fields ...string) *gorm.DB {
	key = strings.TrimSpace(key)
	if key == "" || len(fields) == 0 {
		return query
	}
	like := "%" + key + "%"
	conditions := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		conditions = append(conditions, field+" LIKE ?")
		args = append(args, like)
	}
	if len(conditions) == 0 {
		return query
	}
	return query.Where(strings.Join(conditions, " OR "), args...)
}

// AdminSummaryView 返回后台社交关系总览。
func (SocialApi) AdminSummaryView(c *gin.Context) {
	var followCount, blockCount, groupCount, groupMemberCount, presenceUserCount int64
	global.DB.Model(&models.UserFollowModel{}).Count(&followCount)
	global.DB.Model(&models.UserBlockModel{}).Count(&blockCount)
	global.DB.Model(&models.SocialGroupModel{}).Count(&groupCount)
	global.DB.Model(&models.SocialGroupMemberModel{}).Count(&groupMemberCount)
	global.DB.Model(&models.UserPresenceModel{}).Count(&presenceUserCount)

	var mutualFriendCount int64
	global.DB.Table("user_follow_models AS uf").
		Joins("JOIN user_follow_models AS uf2 ON uf.user_id = uf2.follow_user_id AND uf.follow_user_id = uf2.user_id").
		Where("uf.user_id < uf.follow_user_id").
		Count(&mutualFriendCount)

	res.OkWithData(socialManageOverviewResponse{
		FollowCount:       int(followCount),
		MutualFriendCount: int(mutualFriendCount),
		BlockCount:        int(blockCount),
		GroupCount:        int(groupCount),
		GroupMemberCount:  int(groupMemberCount),
		OnlineUserCount:   len(snapshotOnlinePresence()),
		PresenceUserCount: int(presenceUserCount),
	}, c)
}

// AdminFollowListView 返回后台关注/好友关系列表。
func (SocialApi) AdminFollowListView(c *gin.Context) {
	var cr socialManageQuery
	_ = c.ShouldBindQuery(&cr)
	sanitizeManagePageInfo(&cr, 12, 100)

	mode := strings.ToLower(strings.TrimSpace(cr.Mode))
	if mode == "" {
		mode = "friend"
	}

	query := global.DB.Table("user_follow_models AS uf").
		Select(`
			uf.id,
			uf.user_id,
			u1.nick_name AS user_nick_name,
			u1.user_name AS user_name,
			u1.avatar AS user_avatar,
			uf.follow_user_id,
			u2.nick_name AS follow_nick_name,
			u2.user_name AS follow_user_name,
			u2.avatar AS follow_user_avatar,
			CASE WHEN uf2.id IS NOT NULL THEN true ELSE false END AS is_friend,
			DATE_FORMAT(uf.created_at, '%Y-%m-%d %H:%i:%s') AS created_at
		`).
		Joins("LEFT JOIN user_models AS u1 ON u1.id = uf.user_id").
		Joins("LEFT JOIN user_models AS u2 ON u2.id = uf.follow_user_id").
		Joins("LEFT JOIN user_follow_models AS uf2 ON uf2.user_id = uf.follow_user_id AND uf2.follow_user_id = uf.user_id")

	if mode == "friend" {
		query = query.Where("uf2.id IS NOT NULL AND uf.user_id < uf.follow_user_id")
	} else if mode == "follow" {
		query = query.Where("uf2.id IS NULL")
	}

	query = applySocialManageKeyFilter(query, cr.Key,
		"u1.nick_name", "u1.user_name", "u2.nick_name", "u2.user_name",
		"CAST(uf.user_id AS CHAR)", "CAST(uf.follow_user_id AS CHAR)",
	)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取好友关系数量失败", c)
		return
	}

	var list []socialManageFollowItem
	if err := query.Order("uf.created_at desc").
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Scan(&list).Error; err != nil {
		res.FailWithMessage("获取好友关系列表失败", c)
		return
	}
	res.OkWithList(list, count, c)
}

// AdminBlockListView 返回后台黑名单列表。
func (SocialApi) AdminBlockListView(c *gin.Context) {
	var cr socialManageQuery
	_ = c.ShouldBindQuery(&cr)
	sanitizeManagePageInfo(&cr, 12, 100)

	query := global.DB.Table("user_block_models AS ub").
		Select(`
			ub.id,
			ub.user_id,
			u1.nick_name AS user_nick_name,
			u1.user_name AS user_name,
			u1.avatar AS user_avatar,
			ub.block_user_id,
			u2.nick_name AS block_nick_name,
			u2.user_name AS block_user_name,
			u2.avatar AS block_user_avatar,
			ub.reason,
			DATE_FORMAT(ub.created_at, '%Y-%m-%d %H:%i:%s') AS created_at
		`).
		Joins("LEFT JOIN user_models AS u1 ON u1.id = ub.user_id").
		Joins("LEFT JOIN user_models AS u2 ON u2.id = ub.block_user_id")

	query = applySocialManageKeyFilter(query, cr.Key,
		"u1.nick_name", "u1.user_name", "u2.nick_name", "u2.user_name", "ub.reason",
		"CAST(ub.user_id AS CHAR)", "CAST(ub.block_user_id AS CHAR)",
	)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取黑名单数量失败", c)
		return
	}

	var list []socialManageBlockItem
	if err := query.Order("ub.created_at desc").
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Scan(&list).Error; err != nil {
		res.FailWithMessage("获取黑名单列表失败", c)
		return
	}
	res.OkWithList(list, count, c)
}

// AdminGroupListView 返回后台好友群组列表。
func (SocialApi) AdminGroupListView(c *gin.Context) {
	var cr socialManageQuery
	_ = c.ShouldBindQuery(&cr)
	sanitizeManagePageInfo(&cr, 12, 100)

	query := global.DB.Table("social_group_models AS sg").
		Select(`
			sg.id,
			sg.group_no,
			sg.name,
			sg.avatar,
			sg.notice,
			sg.owner_user_id,
			COALESCE(NULLIF(owner.nick_name, ''), owner.user_name) AS owner_name,
			(SELECT COUNT(1) FROM social_group_member_models AS gm WHERE gm.group_id = sg.id) AS member_count,
			DATE_FORMAT(sg.created_at, '%Y-%m-%d %H:%i:%s') AS created_at,
			DATE_FORMAT(sg.updated_at, '%Y-%m-%d %H:%i:%s') AS updated_at
		`).
		Joins("LEFT JOIN user_models AS owner ON owner.id = sg.owner_user_id")

	query = applySocialManageKeyFilter(query, cr.Key,
		"sg.name", "sg.group_no", "sg.notice", "owner.nick_name", "owner.user_name",
		"CAST(sg.owner_user_id AS CHAR)",
	)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		res.FailWithMessage("获取群组数量失败", c)
		return
	}

	var list []socialManageGroupItem
	if err := query.Order("sg.updated_at desc").
		Limit(cr.Limit).
		Offset((cr.Page - 1) * cr.Limit).
		Scan(&list).Error; err != nil {
		res.FailWithMessage("获取群组列表失败", c)
		return
	}
	res.OkWithList(list, count, c)
}
