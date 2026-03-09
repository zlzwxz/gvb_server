package es_ser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
	"unicode"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/service/board_ser"
	"gvb-server/utils/jwts"
	"gvb-server/utils/sanitize"

	"github.com/PuerkitoBio/goquery"
	"github.com/olivere/elastic/v7"
	"github.com/russross/blackfriday"
)

var (
	articleIndexName  = (models.ArticleModel{}).Index()
	fullTextIndexName = (models.FullTextModel{}).Index()
)

type ESIndexSummary struct {
	Name            string `json:"name"`
	DocsCount       int64  `json:"docs_count"`
	ExportSupported bool   `json:"export_supported"`
	ImportSupported bool   `json:"import_supported"`
	Description     string `json:"description"`
}

type ArticleImportSuccess struct {
	Row      int    `json:"row"`
	SourceID string `json:"source_id"`
	NewID    string `json:"new_id"`
	Title    string `json:"title"`
}

type ArticleImportFailure struct {
	Row      int    `json:"row"`
	SourceID string `json:"source_id"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}

type ArticleImportResult struct {
	Index     string                 `json:"index"`
	Total     int                    `json:"total"`
	Success   int                    `json:"success"`
	Failed    int                    `json:"failed"`
	Created   []ArticleImportSuccess `json:"created"`
	Failures  []ArticleImportFailure `json:"failures"`
	Completed string                 `json:"completed"`
}

type articleImportContext struct {
	operator       *jwts.CustomClaims
	userCache      map[uint]models.UserModel
	bannerPathByID map[uint]string
	bannerIDs      []uint
}

// ListVisibleIndices 返回 ES 中所有可见索引，供后台页面选择导出目标。
func ListVisibleIndices() ([]ESIndexSummary, error) {
	if err := ensureESReady(); err != nil {
		return nil, err
	}

	indexNames, err := global.ESClient.IndexNames()
	if err != nil {
		return nil, fmt.Errorf("获取索引列表失败: %w", err)
	}

	list := make([]ESIndexSummary, 0, len(indexNames))
	for _, name := range indexNames {
		name = strings.TrimSpace(name)
		if name == "" || strings.HasPrefix(name, ".") {
			continue
		}
		count, countErr := global.ESClient.Count(name).Do(ctxBg())
		if countErr != nil {
			return nil, fmt.Errorf("获取索引 %s 文档数量失败: %w", name, countErr)
		}
		desc, importSupported := describeIndex(name)
		list = append(list, ESIndexSummary{
			Name:            name,
			DocsCount:       count,
			ExportSupported: true,
			ImportSupported: importSupported,
			Description:     desc,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		wi, wj := indexWeight(list[i].Name), indexWeight(list[j].Name)
		if wi != wj {
			return wi < wj
		}
		return list[i].Name < list[j].Name
	})
	return list, nil
}

// ExportIndexPayload 生成某个索引的导出结构，供 HTTP 接口或命令行复用。
func ExportIndexPayload(indexName string) (*ESIndexBackupPayload, error) {
	if err := ensureESReady(); err != nil {
		return nil, err
	}
	indexName = strings.TrimSpace(indexName)
	if indexName == "" {
		return nil, errors.New("索引名称不能为空")
	}

	documents, err := loadAllIndexDocuments(indexName)
	if err != nil {
		return nil, err
	}
	desc, _ := describeIndex(indexName)
	payload := buildIndexBackupPayload(indexName, desc, documents)
	return &payload, nil
}

// ExportIndexPayloadJSON 将选中索引直接序列化为 JSON 下载内容。
func ExportIndexPayloadJSON(indexName string) ([]byte, error) {
	payload, err := ExportIndexPayload(indexName)
	if err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化索引导出数据失败: %w", err)
	}
	return data, nil
}

// ParseIndexBackupPayload 解析前端上传的 ES 导出文件。
func ParseIndexBackupPayload(data []byte) (*ESIndexBackupPayload, error) {
	if len(data) == 0 {
		return nil, errors.New("上传文件为空")
	}
	var payload ESIndexBackupPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("导入文件不是合法 JSON: %w", err)
	}
	if payload.Version <= 0 {
		return nil, errors.New("导入文件缺少有效版本号")
	}
	if strings.TrimSpace(payload.Index) == "" {
		return nil, errors.New("导入文件缺少 index 字段")
	}
	if payload.Documents == nil {
		return nil, errors.New("导入文件缺少 documents 数组")
	}
	return &payload, nil
}

// ImportArticleIndexPayload 将 article_index 导出文件按“创建文章”逻辑重新导入。
func ImportArticleIndexPayload(data []byte, selectedIndex string, operator *jwts.CustomClaims) (ArticleImportResult, error) {
	result := ArticleImportResult{
		Index:     articleIndexName,
		Created:   []ArticleImportSuccess{},
		Failures:  []ArticleImportFailure{},
		Completed: time.Now().Format("2006-01-02 15:04:05"),
	}

	payload, err := ParseIndexBackupPayload(data)
	if err != nil {
		return result, err
	}
	selectedIndex = strings.TrimSpace(selectedIndex)
	if selectedIndex != "" && selectedIndex != payload.Index {
		return result, fmt.Errorf("当前选择的索引是 %s，但导入文件声明的索引是 %s", selectedIndex, payload.Index)
	}
	if payload.Index != articleIndexName {
		return result, fmt.Errorf("当前仅支持导入 %s 导出的文件", articleIndexName)
	}

	ctx := &articleImportContext{
		operator:       operator,
		userCache:      map[uint]models.UserModel{},
		bannerPathByID: map[uint]string{},
	}
	result.Index = payload.Index
	result.Total = len(payload.Documents)

	for idx, item := range payload.Documents {
		row := idx + 1
		article, decodeErr := decodeArticleBackupItem(item)
		if decodeErr != nil {
			result.Failures = append(result.Failures, ArticleImportFailure{
				Row:      row,
				SourceID: strings.TrimSpace(item.ID),
				Message:  "source 字段无法解析为文章结构: " + decodeErr.Error(),
			})
			continue
		}

		created, importErr := ctx.importArticle(article)
		if importErr != nil {
			result.Failures = append(result.Failures, ArticleImportFailure{
				Row:      row,
				SourceID: strings.TrimSpace(item.ID),
				Title:    strings.TrimSpace(article.Title),
				Message:  importErr.Error(),
			})
			continue
		}

		result.Created = append(result.Created, ArticleImportSuccess{
			Row:      row,
			SourceID: strings.TrimSpace(item.ID),
			NewID:    created.ID,
			Title:    created.Title,
		})
		result.Success++
	}

	result.Failed = len(result.Failures)
	return result, nil
}

func (ctx *articleImportContext) importArticle(source models.ArticleModel) (models.ArticleModel, error) {
	title := strings.TrimSpace(source.Title)
	if title == "" {
		return models.ArticleModel{}, errors.New("文章标题不能为空")
	}
	if (models.ArticleModel{Title: title}).ISExistData() {
		return models.ArticleModel{}, errors.New("文章标题已存在")
	}

	content := sanitize.CleanMarkdownInput(source.Content)
	if content == "" {
		return models.ArticleModel{}, errors.New("文章内容不能为空或清洗后为空")
	}

	board, err := board_ser.GetEnabledBoardByID(source.BoardID)
	if err != nil {
		return models.ArticleModel{}, fmt.Errorf("板块不存在或已停用: %d", source.BoardID)
	}

	author, err := ctx.resolveAuthor(source.UserID)
	if err != nil {
		return models.ArticleModel{}, err
	}
	bannerID, bannerURL, err := ctx.resolveBanner(source.BannerID)
	if err != nil {
		return models.ArticleModel{}, err
	}

	link := strings.TrimSpace(source.Link)
	if link != "" {
		link = sanitize.CleanURL(link, true)
		if link == "" {
			return models.ArticleModel{}, errors.New("文章链接格式错误，仅支持 http/https 或站内相对路径")
		}
	}

	nowText := time.Now().Format("2006-01-02 15:04:05")
	createdAt := normalizeImportTime(source.CreatedAt, nowText)
	updatedAt := normalizeImportTime(source.UpdatedAt, createdAt)
	reviewStatus, reviewReason, reviewedAt, reviewerID, reviewerName := normalizeReviewMeta(source, ctx.operator)

	article := models.ArticleModel{
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		Title:            title,
		Keyword:          title,
		Abstract:         buildImportAbstract(source.Abstract, content),
		Content:          content,
		LookCount:        normalizeCounter(source.LookCount),
		CommentCount:     normalizeCounter(source.CommentCount),
		DiggCount:        normalizeCounter(source.DiggCount),
		CollectsCount:    normalizeCounter(source.CollectsCount),
		UserID:           author.ID,
		UserNickName:     strings.TrimSpace(author.NickName),
		UserAvatar:       strings.TrimSpace(author.Avatar),
		Category:         board.Name,
		Source:           strings.TrimSpace(source.Source),
		Link:             link,
		BoardID:          board.ID,
		BoardName:        board.Name,
		BannerID:         bannerID,
		BannerUrl:        bannerURL,
		Tags:             normalizeTags(source.Tags),
		Attachments:      normalizeAttachments(source.Attachments),
		IsPrivate:        source.IsPrivate,
		ReviewStatus:     reviewStatus,
		ReviewReason:     reviewReason,
		ReviewedAt:       reviewedAt,
		ReviewerID:       reviewerID,
		ReviewerNickName: reviewerName,
	}

	if err = article.Create(); err != nil {
		return models.ArticleModel{}, fmt.Errorf("写入 ES 失败: %w", err)
	}

	if rate, duplicateID, duplicateTitle, duplicateErr := calculateImportDuplicateRate(article.ID, article.Title, article.Content); duplicateErr == nil {
		article.DuplicateRate = rate
		article.DuplicateTargetID = duplicateID
		article.DuplicateTargetTitle = duplicateTitle
		_ = ArticleUpdate(article.ID, map[string]any{
			"duplicate_rate":         rate,
			"duplicate_target_id":    duplicateID,
			"duplicate_target_title": duplicateTitle,
		})
	}
	if allowFullTextRebuild(article) {
		AsyncArticleByFullText(SearchData{
			Key:   article.ID,
			Body:  article.Content,
			Slug:  GetSlug(article.Title),
			Title: article.Title,
		})
	}
	return article, nil
}

func (ctx *articleImportContext) resolveAuthor(userID uint) (models.UserModel, error) {
	if userID == 0 && ctx.operator != nil {
		userID = ctx.operator.UserID
	}
	if userID == 0 {
		return models.UserModel{}, errors.New("作者用户 ID 不能为空")
	}
	if cached, ok := ctx.userCache[userID]; ok {
		return cached, nil
	}
	var user models.UserModel
	if err := global.DB.Take(&user, userID).Error; err != nil {
		return models.UserModel{}, fmt.Errorf("作者用户不存在: %d", userID)
	}
	ctx.userCache[userID] = user
	return user, nil
}

func (ctx *articleImportContext) resolveBanner(bannerID uint) (uint, string, error) {
	if bannerID != 0 {
		if path, ok := ctx.bannerPathByID[bannerID]; ok {
			return bannerID, path, nil
		}
		var banner models.BannerModel
		if err := global.DB.Take(&banner, bannerID).Error; err != nil {
			return 0, "", fmt.Errorf("文章封面不存在: %d", bannerID)
		}
		path := strings.TrimSpace(banner.Path)
		if path == "" {
			return 0, "", fmt.Errorf("文章封面路径为空: %d", bannerID)
		}
		ctx.bannerPathByID[bannerID] = path
		return bannerID, path, nil
	}

	if len(ctx.bannerIDs) == 0 {
		if err := global.DB.Model(&models.BannerModel{}).Select("id").Scan(&ctx.bannerIDs).Error; err != nil {
			return 0, "", fmt.Errorf("读取封面列表失败: %w", err)
		}
	}
	if len(ctx.bannerIDs) == 0 {
		return 0, "", errors.New("系统里没有可用封面图，无法导入文章")
	}

	selectedID := ctx.bannerIDs[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(ctx.bannerIDs))]
	path, ok := ctx.bannerPathByID[selectedID]
	if ok {
		return selectedID, path, nil
	}
	var banner models.BannerModel
	if err := global.DB.Take(&banner, selectedID).Error; err != nil {
		return 0, "", fmt.Errorf("读取随机封面失败: %d", selectedID)
	}
	path = strings.TrimSpace(banner.Path)
	if path == "" {
		return 0, "", fmt.Errorf("随机封面路径为空: %d", selectedID)
	}
	ctx.bannerPathByID[selectedID] = path
	return selectedID, path, nil
}

func describeIndex(indexName string) (string, bool) {
	switch strings.TrimSpace(indexName) {
	case articleIndexName:
		return "文章主索引，可导出，也支持按业务规则导入", true
	case fullTextIndexName:
		return "全文搜索派生索引，只建议导出排查，不支持页面导入", false
	default:
		return "普通 ES 索引，可导出备份", false
	}
}

func indexWeight(indexName string) int {
	switch strings.TrimSpace(indexName) {
	case articleIndexName:
		return 0
	case fullTextIndexName:
		return 1
	default:
		return 10
	}
}

func normalizeImportTime(value, fallback string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return fallback
	}
	if _, err := time.Parse("2006-01-02 15:04:05", text); err != nil {
		return fallback
	}
	return text
}

func normalizeReviewMeta(source models.ArticleModel, operator *jwts.CustomClaims) (ctype.ArticleReviewStatus, string, string, uint, string) {
	nowText := time.Now().Format("2006-01-02 15:04:05")
	switch source.ReviewStatus {
	case ctype.ArticleReviewPending:
		return ctype.ArticleReviewPending, "", "", 0, ""
	case ctype.ArticleReviewRejected:
		reviewerID, reviewerName := pickReviewer(source, operator)
		reason := strings.TrimSpace(source.ReviewReason)
		if reason == "" {
			reason = "导入时保留驳回状态，但原文件未提供驳回原因"
		}
		return ctype.ArticleReviewRejected, reason, normalizeImportTime(source.ReviewedAt, nowText), reviewerID, reviewerName
	case ctype.ArticleReviewApproved, ctype.ArticleReviewLegacy:
		reviewerID, reviewerName := pickReviewer(source, operator)
		return source.ReviewStatus, "", normalizeImportTime(source.ReviewedAt, nowText), reviewerID, reviewerName
	default:
		if operator != nil {
			return ctype.ArticleReviewApproved, "", nowText, operator.UserID, strings.TrimSpace(operator.NickName)
		}
		return ctype.ArticleReviewApproved, "", nowText, 0, ""
	}
}

func pickReviewer(source models.ArticleModel, operator *jwts.CustomClaims) (uint, string) {
	if source.ReviewerID != 0 || strings.TrimSpace(source.ReviewerNickName) != "" {
		return source.ReviewerID, strings.TrimSpace(source.ReviewerNickName)
	}
	if operator != nil {
		return operator.UserID, strings.TrimSpace(operator.NickName)
	}
	return 0, ""
}

func buildImportAbstract(rawAbstract, content string) string {
	abstract := strings.TrimSpace(rawAbstract)
	if abstract != "" {
		return abstract
	}
	unsafe := blackfriday.MarkdownCommon([]byte(content))
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(unsafe)))
	text := []rune(strings.TrimSpace(doc.Text()))
	if len(text) > 100 {
		return string(text[:100])
	}
	return string(text)
}

func normalizeCounter(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func normalizeTags(tags ctype.Array) ctype.Array {
	list := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, tag := range tags {
		text := strings.TrimSpace(tag)
		if text == "" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		list = append(list, text)
	}
	return ctype.Array(list)
}

func normalizeAttachments(attachments []models.ArticleAttachment) []models.ArticleAttachment {
	list := make([]models.ArticleAttachment, 0, len(attachments))
	for _, item := range attachments {
		name := strings.TrimSpace(item.Name)
		url := strings.TrimSpace(item.URL)
		if item.FileID == 0 && name == "" && url == "" {
			continue
		}
		if item.Size < 0 {
			item.Size = 0
		}
		item.Name = name
		item.URL = url
		list = append(list, item)
	}
	return list
}

func ctxBg() context.Context {
	return context.Background()
}

func calculateImportDuplicateRate(articleID string, title string, content string) (float64, string, string, error) {
	if global.ESClient == nil {
		return 0, "", "", errors.New("ES未初始化")
	}
	seed := strings.TrimSpace(strings.Join([]string{title, content}, "\n"))
	if seed == "" {
		return 0, "", "", nil
	}

	mltQuery := elastic.NewMoreLikeThisQuery().
		Field("title").
		Field("content").
		LikeText(seed).
		MinDocFreq(1).
		MinTermFreq(1).
		Analyzer("standard")

	boolQuery := elastic.NewBoolQuery().Must(mltQuery)
	if strings.TrimSpace(articleID) != "" {
		boolQuery.MustNot(elastic.NewIdsQuery().Ids(articleID))
	}

	result, err := global.ESClient.
		Search(articleIndexName).
		Query(boolQuery).
		Size(8).
		Do(context.Background())
	if err != nil {
		return 0, "", "", err
	}
	if result.Hits == nil || len(result.Hits.Hits) == 0 {
		return 0, "", "", nil
	}

	contentSet := buildImportTextShingleSet(content, 2)
	titleSet := buildImportTextShingleSet(title, 2)
	bestRate := 0.0
	bestID := ""
	bestTitle := ""

	for _, hit := range result.Hits.Hits {
		if strings.TrimSpace(hit.Id) == "" {
			continue
		}
		var article models.ArticleModel
		if unmarshalErr := json.Unmarshal(hit.Source, &article); unmarshalErr != nil {
			continue
		}

		contentRate := jaccardImportRate(contentSet, buildImportTextShingleSet(article.Content, 2))
		titleRate := jaccardImportRate(titleSet, buildImportTextShingleSet(article.Title, 2))
		score := contentRate*0.85 + titleRate*0.15
		if score > bestRate {
			bestRate = score
			bestID = hit.Id
			bestTitle = strings.TrimSpace(article.Title)
		}
	}

	return roundImportFloat(bestRate*100, 2), bestID, bestTitle, nil
}

func buildImportTextShingleSet(raw string, n int) map[string]struct{} {
	set := map[string]struct{}{}
	normalized := normalizeImportDuplicateText(raw)
	if normalized == "" {
		return set
	}
	runes := []rune(normalized)
	if len(runes) <= n {
		set[normalized] = struct{}{}
		return set
	}
	if len(runes) > 5000 {
		runes = runes[:5000]
	}
	for idx := 0; idx <= len(runes)-n; idx++ {
		part := string(runes[idx : idx+n])
		if strings.TrimSpace(part) == "" {
			continue
		}
		set[part] = struct{}{}
	}
	return set
}

func normalizeImportDuplicateText(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		switch {
		case unicode.Is(unicode.Han, r):
			builder.WriteRune(r)
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(r)
		case unicode.IsSpace(r):
			builder.WriteRune(' ')
		default:
			// ignore punctuation
		}
	}
	normalized := strings.Join(strings.Fields(builder.String()), " ")
	return strings.TrimSpace(normalized)
}

func jaccardImportRate(a map[string]struct{}, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for key := range a {
		if _, ok := b[key]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union <= 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func roundImportFloat(value float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(value*pow) / pow
}
