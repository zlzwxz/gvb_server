package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gvb-server/core"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/utils/pwd"

	"github.com/olivere/elastic/v7"
)

const (
	targetArticleCount = 1200
	targetCommentCount = 1200
	targetMessageCount = 600
	targetChatCount    = 300
	targetTagCount     = 24
	targetAdvertCount  = 8
)

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if err := initDeps(); err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

	users, err := ensureUsers()
	if err != nil {
		fmt.Printf("用户初始化失败: %v\n", err)
		return
	}
	if len(users) == 0 {
		fmt.Println("没有可用用户，终止造数")
		return
	}

	localBanners, err := ensureLocalUserBanners(users, r)
	if err != nil {
		fmt.Printf("本地图片初始化失败: %v\n", err)
		return
	}

	remoteBanners, err := ensureRemoteShowcaseBanners()
	if err != nil {
		fmt.Printf("远程轮播图初始化失败: %v\n", err)
		return
	}

	allBanners, err := listAllBanners()
	if err != nil {
		fmt.Printf("读取图片列表失败: %v\n", err)
		return
	}
	if len(allBanners) == 0 {
		allBanners = append(allBanners, append(localBanners, remoteBanners...)...)
	}

	if err = ensureMenus(remoteBanners); err != nil {
		fmt.Printf("菜单初始化失败: %v\n", err)
		return
	}

	tags, err := ensureTags()
	if err != nil {
		fmt.Printf("标签初始化失败: %v\n", err)
		return
	}

	createdArticleIDs, err := seedArticles(users, allBanners, tags, r)
	if err != nil {
		fmt.Printf("文章造数失败: %v\n", err)
		return
	}

	articleIDs, err := listRecentArticleIDs(800)
	if err != nil {
		fmt.Printf("读取文章ID失败: %v\n", err)
		return
	}
	if len(articleIDs) == 0 {
		articleIDs = createdArticleIDs
	}

	if err = ensureComments(articleIDs, users, r); err != nil {
		fmt.Printf("评论造数失败: %v\n", err)
		return
	}
	if err = ensurePrivateMessages(users, r); err != nil {
		fmt.Printf("私信造数失败: %v\n", err)
		return
	}
	if err = ensureChatMessages(users, r); err != nil {
		fmt.Printf("聊天室造数失败: %v\n", err)
		return
	}
	if err = ensureAdverts(remoteBanners, r); err != nil {
		fmt.Printf("广告造数失败: %v\n", err)
		return
	}

	articleCount, _ := global.ESClient.Count(models.ArticleModel{}.Index()).Do(context.Background())
	var commentCount int64
	_ = global.DB.Model(&models.CommentModel{}).Count(&commentCount).Error
	var messageCount int64
	_ = global.DB.Model(&models.MessageModel{}).Count(&messageCount).Error
	fmt.Printf("造数完成：文章 %d，评论 %d，私信 %d\n", articleCount, commentCount, messageCount)
}

func initDeps() error {
	core.InitConf()
	global.Log = core.InitLogger()
	global.DB = core.InitGorm()
	if global.DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	esClient, err := core.EsConnect()
	if err != nil {
		return err
	}
	global.ESClient = esClient

	articleModel := models.ArticleModel{}
	if !articleModel.IndexExists() {
		if err = articleModel.CreateIndex(); err != nil {
			return err
		}
	}
	return nil
}

func ensureUsers() ([]models.UserModel, error) {
	var users []models.UserModel
	if err := global.DB.Where("role in ?", []int{int(ctype.PermissionAdmin), int(ctype.PermissionUser)}).Find(&users).Error; err != nil {
		return nil, err
	}

	if len(users) >= 8 {
		return users, nil
	}

	for i := len(users); i < 8; i++ {
		role := ctype.PermissionUser
		if i == 0 {
			role = ctype.PermissionAdmin
		}
		userName := fmt.Sprintf("seed_user_%03d", i+1)
		nickName := fmt.Sprintf("演示用户%03d", i+1)

		var existing models.UserModel
		if err := global.DB.Where("user_name = ?", userName).Take(&existing).Error; err == nil {
			continue
		}

		user := models.UserModel{
			UserName: userName,
			NickName: nickName,
			Password: pwd.HashPwd("123456"),
			Role:     role,
			Avatar:   "",
			Addr:     "北京",
		}
		if err := global.DB.Create(&user).Error; err != nil {
			return nil, err
		}
	}

	users = make([]models.UserModel, 0)
	if err := global.DB.Where("role in ?", []int{int(ctype.PermissionAdmin), int(ctype.PermissionUser)}).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func ensureLocalUserBanners(users []models.UserModel, r *rand.Rand) ([]models.BannerModel, error) {
	sourceDir := filepath.Join("testdata", "uploads", "chat_avatar")
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}

	avatars := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") {
			avatars = append(avatars, filepath.Join(sourceDir, entry.Name()))
		}
	}
	sort.Strings(avatars)
	if len(avatars) == 0 {
		return nil, fmt.Errorf("未找到可复用头像素材")
	}

	result := make([]models.BannerModel, 0)
	for _, user := range users {
		nickName := strings.TrimSpace(user.NickName)
		if nickName == "" {
			nickName = user.UserName
		}
		if nickName == "" {
			continue
		}

		targetDir := filepath.Join("uploads", "file", nickName)
		if err = os.MkdirAll(targetDir, 0755); err != nil {
			return nil, err
		}

		for i := 0; i < 4; i++ {
			srcPath := avatars[r.Intn(len(avatars))]
			srcData, readErr := os.ReadFile(srcPath)
			if readErr != nil {
				continue
			}
			fileName := fmt.Sprintf("seed_%02d_%03d.png", i+1, user.ID)
			dstPath := filepath.Join(targetDir, fileName)
			if writeErr := os.WriteFile(dstPath, srcData, 0644); writeErr != nil {
				continue
			}

			relPath := "/" + filepath.ToSlash(dstPath)
			hash := md5Hex(srcData)
			banner := models.BannerModel{
				Path:      relPath,
				Hash:      hash,
				Name:      fmt.Sprintf("%s-%s", nickName, fileName),
				ImageType: ctype.Local,
			}

			var existing models.BannerModel
			if err = global.DB.Where("path = ?", relPath).Take(&existing).Error; err == nil {
				result = append(result, existing)
				continue
			}
			if err = global.DB.Create(&banner).Error; err != nil {
				continue
			}
			result = append(result, banner)
		}
	}
	return result, nil
}

func ensureRemoteShowcaseBanners() ([]models.BannerModel, error) {
	paths := []string{
		"https://picsum.photos/seed/gvb-community-1/1600/900",
		"https://picsum.photos/seed/gvb-community-2/1600/900",
		"https://picsum.photos/seed/gvb-community-3/1600/900",
		"https://picsum.photos/seed/gvb-community-4/1600/900",
		"https://picsum.photos/seed/gvb-community-5/1600/900",
		"https://picsum.photos/seed/gvb-community-6/1600/900",
		"https://picsum.photos/seed/gvb-community-7/1600/900",
		"https://picsum.photos/seed/gvb-community-8/1600/900",
	}

	result := make([]models.BannerModel, 0, len(paths))
	for i, path := range paths {
		var existing models.BannerModel
		if err := global.DB.Where("path = ?", path).Take(&existing).Error; err == nil {
			result = append(result, existing)
			continue
		}

		banner := models.BannerModel{
			Path:      path,
			Hash:      md5Hex([]byte(path)),
			Name:      fmt.Sprintf("community_showcase_%02d", i+1),
			ImageType: ctype.QiNiu,
		}
		if err := global.DB.Create(&banner).Error; err != nil {
			return nil, err
		}
		result = append(result, banner)
	}
	return result, nil
}

func listAllBanners() ([]models.BannerModel, error) {
	list := make([]models.BannerModel, 0)
	if err := global.DB.Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func ensureMenus(showcaseBanners []models.BannerModel) error {
	menuSeeds := []models.MenuModel{
		{Title: "首页", Path: "/", Slogan: "社区焦点与精选内容", Abstract: ctype.Array{"实时热点聚合", "高质量文章推荐", "互动内容运营"}, AbstractTime: 6, BannerTime: 5, Sort: 1},
		{Title: "资讯", Path: "/news", Slogan: "多源热榜资讯", Abstract: ctype.Array{"微博/知乎/百度等多平台切换", "后台可动态控制榜单展示"}, AbstractTime: 6, BannerTime: 6, Sort: 2},
		{Title: "社区", Path: "/chat", Slogan: "实时聊天社区", Abstract: ctype.Array{"群聊+私信双通道", "支持图片与文本消息"}, AbstractTime: 6, BannerTime: 7, Sort: 3},
	}

	for _, menu := range menuSeeds {
		var existing models.MenuModel
		if err := global.DB.Where("path = ?", menu.Path).Take(&existing).Error; err == nil {
			_ = global.DB.Model(&existing).Updates(map[string]any{
				"title":         menu.Title,
				"slogan":        menu.Slogan,
				"abstract":      menu.Abstract,
				"abstract_time": menu.AbstractTime,
				"banner_time":   menu.BannerTime,
				"sort":          menu.Sort,
			}).Error
			continue
		}
		if err := global.DB.Create(&menu).Error; err != nil {
			return err
		}
	}

	if len(showcaseBanners) == 0 {
		return nil
	}

	var homeMenu models.MenuModel
	if err := global.DB.Where("path = ?", "/").Take(&homeMenu).Error; err != nil {
		return err
	}

	_ = global.DB.Where("menu_id = ?", homeMenu.ID).Delete(&models.MenuBannerModel{}).Error
	for i, banner := range showcaseBanners {
		relation := models.MenuBannerModel{
			MenuID:   homeMenu.ID,
			BannerID: banner.ID,
			Sort:     i + 1,
		}
		_ = global.DB.Create(&relation).Error
		if i >= 4 {
			break
		}
	}
	return nil
}

func ensureTags() ([]string, error) {
	tagSeeds := []string{
		"后端", "前端", "Go", "Vue", "AI", "工程化", "性能优化", "系统设计",
		"微服务", "容器化", "数据库", "可观测性", "运营", "产品", "社区",
		"安全", "架构", "测试", "部署", "开源", "资讯", "热点", "教程", "实战",
	}

	var currentCount int64
	if err := global.DB.Model(&models.TagModel{}).Count(&currentCount).Error; err != nil {
		return nil, err
	}
	if currentCount >= targetTagCount {
		return tagSeeds, nil
	}

	for _, title := range tagSeeds {
		var existing models.TagModel
		if err := global.DB.Where("title = ?", title).Take(&existing).Error; err == nil {
			continue
		}
		tag := models.TagModel{Title: title}
		_ = global.DB.Create(&tag).Error
	}
	return tagSeeds, nil
}

func seedArticles(users []models.UserModel, banners []models.BannerModel, tags []string, r *rand.Rand) ([]string, error) {
	count, err := global.ESClient.Count(models.ArticleModel{}.Index()).Do(context.Background())
	if err != nil {
		return nil, err
	}
	need := targetArticleCount - int(count)
	if need <= 0 {
		return listRecentArticleIDs(300)
	}

	topics := []string{"社区热议", "技术周报", "开发实践", "架构观察", "前端趋势", "后端优化", "产品洞察", "运营复盘"}
	categories := []string{"后端开发", "前端工程", "产品运营", "技术管理", "社区观察"}
	createdIDs := make([]string, 0, need)

	for i := 0; i < need; i++ {
		user := users[r.Intn(len(users))]
		banner := banners[r.Intn(len(banners))]
		articleTags := pickTags(tags, 2+r.Intn(2), r)

		title := fmt.Sprintf("%s｜第%04d期", topics[r.Intn(len(topics))], int(count)+i+1)
		content := buildArticleContent(title, articleTags)
		now := time.Now().Add(-time.Duration(r.Intn(180*24)) * time.Hour).Format("2006-01-02 15:04:05")
		article := models.ArticleModel{
			CreatedAt:    now,
			UpdatedAt:    now,
			Title:        title,
			Keyword:      title,
			Abstract:     fmt.Sprintf("%s，覆盖热点拆解、关键数据和可执行建议。", title),
			Content:      content,
			LookCount:    20 + r.Intn(8000),
			CommentCount: r.Intn(20),
			DiggCount:    r.Intn(120),
			UserID:       user.ID,
			UserNickName: user.NickName,
			UserAvatar:   user.Avatar,
			Category:     categories[r.Intn(len(categories))],
			Source:       "GVB 社区",
			Link:         "",
			BannerID:     banner.ID,
			BannerUrl:    banner.Path,
			Tags:         ctype.Array(articleTags),
		}
		if err = article.Create(); err != nil {
			return createdIDs, err
		}
		createdIDs = append(createdIDs, article.ID)
		if (i+1)%100 == 0 {
			fmt.Printf("已写入文章 %d/%d\n", i+1, need)
		}
	}
	return createdIDs, nil
}

func listRecentArticleIDs(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	res, err := global.ESClient.Search(models.ArticleModel{}.Index()).
		Sort("created_at", false).
		Size(limit).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		ids = append(ids, hit.Id)
	}
	return ids, nil
}

func ensureComments(articleIDs []string, users []models.UserModel, r *rand.Rand) error {
	if len(articleIDs) == 0 {
		return nil
	}
	var count int64
	if err := global.DB.Model(&models.CommentModel{}).Count(&count).Error; err != nil {
		return err
	}
	need := targetCommentCount - int(count)
	if need <= 0 {
		return nil
	}

	commentTemplates := []string{
		"这个观点很有参考价值，已经收藏。",
		"补充一个一线实践：先做小范围灰度再全量。",
		"数据维度能再展开下吗？很想看后续。",
		"这个方案我们线上跑过，收益很明显。",
		"写得很清楚，适合团队内部分享。",
		"建议把风险点也列出来，便于评审。",
	}

	increase := map[string]int{}
	for i := 0; i < need; i++ {
		articleID := articleIDs[r.Intn(len(articleIDs))]
		user := users[r.Intn(len(users))]
		root := models.CommentModel{
			ArticleID: articleID,
			UserID:    user.ID,
			Content:   commentTemplates[r.Intn(len(commentTemplates))],
		}
		if err := global.DB.Create(&root).Error; err != nil {
			continue
		}
		increase[articleID]++

		if r.Intn(100) < 45 {
			replyUser := users[r.Intn(len(users))]
			reply := models.CommentModel{
				ArticleID:       articleID,
				UserID:          replyUser.ID,
				ParentCommentID: &root.ID,
				Content:         "同意楼上，补充了我们团队最近一周的监控结果。",
			}
			if err := global.DB.Create(&reply).Error; err == nil {
				increase[articleID]++
				_ = global.DB.Model(&root).Update("comment_count", 1).Error
			}
		}
	}

	for articleID, inc := range increase {
		script := elastic.NewScript("ctx._source.comment_count += params.inc").Param("inc", inc)
		_, _ = global.ESClient.Update().
			Index(models.ArticleModel{}.Index()).
			Id(articleID).
			Script(script).
			Do(context.Background())
	}
	return nil
}

func ensurePrivateMessages(users []models.UserModel, r *rand.Rand) error {
	if len(users) < 2 {
		return nil
	}
	var count int64
	if err := global.DB.Model(&models.MessageModel{}).Count(&count).Error; err != nil {
		return err
	}
	need := targetMessageCount - int(count)
	if need <= 0 {
		return nil
	}

	msgTemplates := []string{
		"你好，这篇文章写得不错，方便交流下实现细节吗？",
		"今晚我们准备做版本回归，一起看下接口变更。",
		"你的评论很专业，能否补充一下排查步骤？",
		"我把文档更新了，麻烦抽空帮我过一下。",
		"这个需求优先级提高了，明早一起对齐。",
	}

	for i := 0; i < need; i++ {
		sendIdx := r.Intn(len(users))
		revIdx := r.Intn(len(users) - 1)
		if revIdx >= sendIdx {
			revIdx++
		}
		sendUser := users[sendIdx]
		revUser := users[revIdx]

		createdAt := time.Now().Add(-time.Duration(r.Intn(30*24)) * time.Hour)
		model := models.MessageModel{
			MODEL:            models.MODEL{CreatedAt: createdAt, UpdatedAt: createdAt},
			SendUserID:       sendUser.ID,
			SendUserNickName: sendUser.NickName,
			SendUserAvatar:   sendUser.Avatar,
			RevUserID:        revUser.ID,
			RevUserNickName:  revUser.NickName,
			RevUserAvatar:    revUser.Avatar,
			IsRead:           r.Intn(100) < 60,
			Content:          msgTemplates[r.Intn(len(msgTemplates))],
		}
		_ = global.DB.Create(&model).Error
	}
	return nil
}

func ensureChatMessages(users []models.UserModel, r *rand.Rand) error {
	var count int64
	if err := global.DB.Model(&models.ChatModel{}).Count(&count).Error; err != nil {
		return err
	}
	need := targetChatCount - int(count)
	if need <= 0 {
		return nil
	}
	chatTemplates := []string{
		"今天的热点更新了吗？",
		"有人试过新的推荐策略吗？",
		"这段代码可以再优化一下缓存层。",
		"我刚发了一篇复盘，欢迎拍砖。",
		"刚刚修了一个线上告警，终于稳了。",
	}
	for i := 0; i < need; i++ {
		user := users[r.Intn(len(users))]
		msg := models.ChatModel{
			NickName: user.NickName,
			Avatar:   user.Avatar,
			Content:  chatTemplates[r.Intn(len(chatTemplates))],
			IP:       "127.0.0.1",
			Addr:     "内网地址",
			IsGroup:  true,
			MsgType:  ctype.MsgType(1),
		}
		_ = global.DB.Create(&msg).Error
	}
	return nil
}

func ensureAdverts(banners []models.BannerModel, r *rand.Rand) error {
	var count int64
	if err := global.DB.Model(&models.AdvertModel{}).Count(&count).Error; err != nil {
		return err
	}
	need := targetAdvertCount - int(count)
	if need <= 0 {
		return nil
	}
	if len(banners) == 0 {
		return nil
	}

	for i := 0; i < need; i++ {
		banner := banners[r.Intn(len(banners))]
		ad := models.AdvertModel{
			Title:  fmt.Sprintf("社区品牌推荐 %02d", i+1),
			Href:   "https://github.com",
			Images: banner.Path,
			IsShow: true,
		}
		_ = global.DB.Create(&ad).Error
	}
	return nil
}

func pickTags(tags []string, size int, r *rand.Rand) []string {
	if len(tags) == 0 || size <= 0 {
		return []string{"社区"}
	}
	if size >= len(tags) {
		return append([]string{}, tags...)
	}
	idxSet := map[int]struct{}{}
	result := make([]string, 0, size)
	for len(result) < size {
		idx := r.Intn(len(tags))
		if _, ok := idxSet[idx]; ok {
			continue
		}
		idxSet[idx] = struct{}{}
		result = append(result, tags[idx])
	}
	return result
}

func buildArticleContent(title string, tags []string) string {
	tagLine := strings.Join(tags, "、")
	return fmt.Sprintf(`## %s

### 一、背景概览
本期围绕「%s」展开，目标是把热点信息转化为可执行的工程动作。

### 二、关键观察
1. 数据趋势在近 24 小时波动明显，建议建立小时级监控。
2. 内容传播呈现社区化扩散，评论区对结论可信度影响更大。
3. 技术侧需要优先保障稳定性，再逐步推进体验优化。

### 三、落地建议
- 建立来源对照表，按平台热度交叉验证。
- 对核心模块增加灰度开关和回滚预案。
- 输出日报模板，统一复盘维度。

### 四、标签
%s
`, title, title, tagLine)
}

func md5Hex(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// 仅用于调试 bulk 搜索结果结构时快速观察。
func debugMarshal(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
