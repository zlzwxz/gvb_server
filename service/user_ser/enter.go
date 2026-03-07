package user_ser

import (
	"crypto/rand"
	"errors"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/service/redis_ser"
	"gvb-server/utils"
	"gvb-server/utils/jwts"
	"gvb-server/utils/pwd"
	random2 "gvb-server/utils/random"
)

// UserService 聚合用户领域的业务逻辑。
type UserService struct{}

// Logout 把当前 token 放进 Redis 黑名单，直到它自然过期。
func (UserService) Logout(claims *jwts.CustomClaims, token string) error {
	exp := claims.ExpiresAt
	now := time.Now()
	diff := exp.Time.Sub(now)
	return redis_ser.Logout(token, diff)
}

// DefaultAvatar 是兜底头像路径。
const DefaultAvatar = "uploads/avatar/default.jpeg"

// RandomAvatar 随机选择一张默认头像。
func RandomAvatar() string {
	matches, err := filepath.Glob(filepath.Join("uploads", "chat_avatar", "*.png"))
	if err != nil || len(matches) == 0 {
		return DefaultAvatar
	}
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(matches))))
	if err != nil {
		return DefaultAvatar
	}
	return filepath.ToSlash(matches[idx.Int64()])
}

// GenerateUniqueUserName 根据给定前缀生成一个尽量可读且不重复的用户名。
func GenerateUniqueUserName(prefix string) string {
	base := strings.TrimSpace(prefix)
	if base == "" {
		base = "user"
	}
	base = strings.ToLower(base)
	base = strings.ReplaceAll(base, " ", "")
	base = strings.ReplaceAll(base, "-", "")
	base = strings.ReplaceAll(base, "_", "")
	if len(base) > 12 {
		base = base[:12]
	}
	if base == "" {
		base = "user"
	}

	for i := 0; i < 10; i++ {
		candidate := base + random2.RandString(6)
		var count int64
		global.DB.Model(&models.UserModel{}).Where("user_name = ?", candidate).Count(&count)
		if count == 0 {
			return candidate
		}
	}
	return "user" + random2.RandString(10)
}

// CreateUser 创建用户并补齐密码加密、地址解析、默认头像等衍生信息。
func (UserService) CreateUser(userName, nickName, password string, role ctype.Role, email string, ip string) error {
	var userModel models.UserModel
	err := global.DB.Take(&userModel, "user_name = ?", userName).Error
	if err == nil {
		return errors.New("用户名已存在")
	}
	if email != "" {
		err = global.DB.Take(&userModel, "email = ?", email).Error
		if err == nil {
			return errors.New("邮箱已存在")
		}
	}

	hashPwd := pwd.HashPwd(password)
	if strings.TrimSpace(hashPwd) == "" {
		return errors.New("密码加密失败")
	}
	addr := utils.GetAddr(ip)
	return global.DB.Create(&models.UserModel{
		NickName:   nickName,
		UserName:   userName,
		Password:   hashPwd,
		Email:      email,
		Role:       role,
		Avatar:     RandomAvatar(),
		IP:         ip,
		Addr:       addr,
		SignStatus: ctype.SignEmail,
		Points:     0,
		Experience: 0,
		Level:      1,
	}).Error
}
