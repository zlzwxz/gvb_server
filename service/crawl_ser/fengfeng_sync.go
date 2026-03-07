package crawl_ser

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/service/board_ser"
	"gvb-server/service/es_ser"
	"gvb-server/service/user_ser"
	random2 "gvb-server/utils/random"
)

const (
	defaultCrawlerUser      = "system_crawler"
	defaultCrawlerNick      = "系统员"
	fengfengHomeURL         = "https://www.fengfengzhidao.com/"
	fengfengSitemapURL      = "https://www.fengfengzhidao.com/sitemap/article.xml"
	fengfengListAPIURL      = "https://www.fengfengzhidao.com/api/articles"
	fengfengAPIURLFmt       = "https://www.fengfengzhidao.com/api/articles/%s"
	fengfengDecryptKey      = "12341234123412341234123412341234"
	fengfengListPageSize    = 200
	maxFengfengListPages    = 80
	defaultArticleScanLimit = 200
	maxArticleScanLimit     = 10000
)

type SyncArticleOptions struct {
	// 仅抓取指定文章 ID；为空时由 SyncAll 决定是否抓取全部候选。
	ArticleIDs []string `json:"article_ids"`
	// 一键抓取：抓取预览中的全部“系统不存在”候选文章。
	SyncAll bool `json:"sync_all"`
	// 对系统已存在的同源文章执行更新。
	IncludeUpdate bool `json:"include_update"`
	// 扫描上限：<0 使用默认值；=0 代表全量扫描；>0 使用指定值。
	Limit int `json:"limit"`
}

type CrawlArticleCandidate struct {
	ArticleID string `json:"article_id"`
	Link      string `json:"link"`
	Title     string `json:"title"`
	Abstract  string `json:"abstract"`
	CoverURL  string `json:"cover_url"`
}

type SyncResult struct {
	SourceTotal    int `json:"source_total"`
	LatestScanned  int `json:"latest_scanned"`
	SyncLimit      int `json:"sync_limit"`
	Created        int `json:"created"`
	UpdatedCount   int `json:"updated_count"`
	DuplicateCount int `json:"duplicate_count"`
	FailedCount    int `json:"failed_count"`
	Skipped        int `json:"skipped"`
	SelectedCount  int `json:"selected_count"`
}

type PreviewResult struct {
	SourceTotal    int                     `json:"source_total"`
	LatestScanned  int                     `json:"latest_scanned"`
	SyncLimit      int                     `json:"sync_limit"`
	NewCandidate   int                     `json:"new_candidate"`
	DuplicateCount int                     `json:"duplicate_count"`
	InvalidCount   int                     `json:"invalid_count"`
	Candidates     []CrawlArticleCandidate `json:"candidates"`
}

type crawlCandidate struct {
	Link      string
	ArticleID string
}

var syncLock sync.Mutex

// EnsureCrawlerAccount 确保自动抓取使用的“系统员”账号存在。
func EnsureCrawlerAccount() (models.UserModel, error) {
	return ensureCrawlerAccount()
}

// PreviewFengfengArticles 检索最新文章数量并返回可新增/重复统计，不执行入库。
func PreviewFengfengArticles() (PreviewResult, error) {
	return PreviewFengfengArticlesWithLimit(-1)
}

// PreviewFengfengArticlesWithLimit 支持指定扫描上限，避免只能检索固定 20 条。
func PreviewFengfengArticlesWithLimit(limit int) (PreviewResult, error) {
	syncLock.Lock()
	defer syncLock.Unlock()
	return previewFengfengArticlesNoLock(limit)
}

func previewFengfengArticlesNoLock(limit int) (PreviewResult, error) {
	linkList, err := fetchFengfengArticleLinks()
	if err != nil {
		return PreviewResult{}, err
	}
	scanLimit := normalizeArticleScanLimit(limit, len(linkList))

	result := PreviewResult{
		SourceTotal:   len(linkList),
		SyncLimit:     scanLimit,
		LatestScanned: 0,
		Candidates:    make([]CrawlArticleCandidate, 0),
	}
	if len(linkList) == 0 {
		return result, nil
	}

	candidates := buildCrawlCandidates(linkList, scanLimit)
	result.LatestScanned = len(candidates)
	var previewMap map[string]CrawlArticleCandidate
	previewMapLoaded := false

	for _, candidate := range candidates {
		if candidate.ArticleID == "" || candidate.Link == "" {
			result.InvalidCount++
			continue
		}
		if existsArticleByFengfengID(candidate.ArticleID) || existsArticleByLink(candidate.Link) {
			result.DuplicateCount++
			continue
		}
		preview := CrawlArticleCandidate{
			ArticleID: candidate.ArticleID,
			Link:      candidate.Link,
		}
		if !previewMapLoaded {
			previewMapLoaded = true
			if cachedMap, previewErr := fetchArticlePreviewMapFromAPI(result.LatestScanned); previewErr == nil {
				previewMap = cachedMap
			} else {
				global.Log.Warnf("读取枫枫文章摘要失败，回退详情抓取: %v", previewErr)
			}
		}
		if cached, ok := previewMap[candidate.ArticleID]; ok {
			preview.Title = cached.Title
			preview.Abstract = cached.Abstract
			preview.CoverURL = cached.CoverURL
		}
		if strings.TrimSpace(preview.Title) == "" || strings.TrimSpace(preview.Abstract) == "" {
			if detail, detailErr := fetchArticleDetail(candidate.Link); detailErr == nil {
				preview.Title = detail.Title
				preview.Abstract = detail.Abstract
				preview.CoverURL = detail.CoverURL
			}
		}
		if strings.TrimSpace(preview.Title) == "" {
			preview.Title = fmt.Sprintf("候选文章 %s", candidate.ArticleID)
		}
		result.Candidates = append(result.Candidates, preview)
		result.NewCandidate++
	}
	return result, nil
}

// SyncFengfengArticles 自动抓取枫枫知道文章列表并写入本地文章索引。
func SyncFengfengArticles() (SyncResult, error) {
	return SyncFengfengArticlesWithOptions(SyncArticleOptions{
		SyncAll:       true,
		IncludeUpdate: true,
		Limit:         -1,
	})
}

// SyncFengfengArticlesWithOptions 支持“抓取选中”与“一键抓取缺失文章”。
func SyncFengfengArticlesWithOptions(options SyncArticleOptions) (SyncResult, error) {
	syncLock.Lock()
	defer syncLock.Unlock()

	owner, err := ensureCrawlerAccount()
	if err != nil {
		return SyncResult{}, err
	}
	linkList, err := fetchFengfengArticleLinks()
	if err != nil {
		return SyncResult{}, err
	}
	scanLimit := normalizeArticleScanLimit(options.Limit, len(linkList))

	result := SyncResult{
		SourceTotal:   len(linkList),
		SyncLimit:     scanLimit,
		LatestScanned: 0,
	}
	if len(linkList) == 0 {
		return result, nil
	}

	candidates := buildCrawlCandidates(linkList, scanLimit)
	result.LatestScanned = len(candidates)
	defaultBoardID, defaultBoardName := resolveDefaultCrawlBoard()
	idFilter := make(map[string]struct{}, len(options.ArticleIDs))
	for _, id := range options.ArticleIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		idFilter[id] = struct{}{}
	}
	syncAll := options.SyncAll || len(idFilter) == 0
	result.SelectedCount = len(idFilter)
	if syncAll {
		result.SelectedCount = 0
	}
	includeUpdate := options.IncludeUpdate || syncAll

	for _, candidate := range candidates {
		if !syncAll {
			if _, ok := idFilter[candidate.ArticleID]; !ok {
				continue
			}
		}
		if candidate.ArticleID == "" || candidate.Link == "" {
			result.FailedCount++
			result.Skipped++
			continue
		}
		existingDocID := findExistingArticleDocID(candidate.Link, candidate.ArticleID)
		if existingDocID != "" {
			result.DuplicateCount++
			if !includeUpdate {
				result.Skipped++
				continue
			}
			articleData, detailErr := fetchArticleDetail(candidate.Link)
			if detailErr != nil || articleData.Title == "" || articleData.Content == "" {
				if detailErr != nil {
					global.Log.Warnf("抓取重复文章详情失败: %s err=%v", candidate.Link, detailErr)
				}
				result.FailedCount++
				result.Skipped++
				continue
			}
			if updateErr := updateExistingArticleByDocID(existingDocID, candidate, articleData); updateErr != nil {
				global.Log.Warnf("更新重复文章失败: %s err=%v", candidate.Link, updateErr)
				result.FailedCount++
				result.Skipped++
				continue
			}
			result.UpdatedCount++
			continue
		}

		articleData, detailErr := fetchArticleDetail(candidate.Link)
		if detailErr != nil {
			global.Log.Warnf("抓取文章详情失败: %s err=%v", candidate.Link, detailErr)
			result.FailedCount++
			result.Skipped++
			continue
		}
		if articleData.Title == "" || articleData.Content == "" {
			result.FailedCount++
			result.Skipped++
			continue
		}

		now := time.Now().Format("2006-01-02 15:04:05")
		bannerID, bannerURL := randomBanner()
		if strings.TrimSpace(articleData.CoverURL) != "" {
			category := inferImageCategory(strings.Join([]string{articleData.Title, articleData.CoverURL}, " "))
			if banner, _, bannerErr := downloadRemoteImageToBanner(articleData.CoverURL, category); bannerErr == nil && banner.ID > 0 {
				bannerID = banner.ID
				bannerURL = banner.Path
			}
		}

		article := models.ArticleModel{
			Title:            articleData.Title,
			Keyword:          articleData.Title,
			Abstract:         articleData.Abstract,
			Content:          articleData.Content,
			Category:         defaultBoardName,
			BoardID:          defaultBoardID,
			BoardName:        defaultBoardName,
			Source:           "https://www.fengfengzhidao.com/",
			Link:             candidate.Link,
			BannerID:         bannerID,
			BannerUrl:        bannerURL,
			Tags:             ctype.Array{"自动采集", "枫枫知道"},
			UserID:           owner.ID,
			UserNickName:     owner.NickName,
			UserAvatar:       owner.Avatar,
			ReviewStatus:     ctype.ArticleReviewApproved,
			ReviewedAt:       now,
			ReviewerID:       owner.ID,
			ReviewerNickName: owner.NickName,
			IsPrivate:        false,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if createErr := article.Create(); createErr != nil {
			global.Log.Warnf("写入ES失败 title=%s err=%v", article.Title, createErr)
			result.FailedCount++
			result.Skipped++
			continue
		}
		if rate, duplicateID, duplicateTitle, duplicateErr := calculateArticleDuplicateRate(article.ID, article.Title, article.Content); duplicateErr == nil {
			article.DuplicateRate = rate
			article.DuplicateTargetID = duplicateID
			article.DuplicateTargetTitle = duplicateTitle
			_ = es_ser.ArticleUpdate(article.ID, map[string]any{
				"duplicate_rate":         rate,
				"duplicate_target_id":    duplicateID,
				"duplicate_target_title": duplicateTitle,
			})
		}

		es_ser.AsyncArticleByFullText(es_ser.SearchData{
			Key:   article.ID,
			Body:  article.Content,
			Slug:  es_ser.GetSlug(article.Title),
			Title: article.Title,
		})
		result.Created++
	}
	return result, nil
}

// SyncFengfengArticlesJob 定时任务入口：仅当后台开关开启时自动执行。
func SyncFengfengArticlesJob() {
	if !global.Config.SiteInfo.AutoCrawl {
		return
	}
	result, err := SyncFengfengArticles()
	if err != nil {
		global.Log.Errorf("自动抓取枫枫知道文章失败: %v", err)
		return
	}
	if result.Created > 0 || result.Skipped > 0 {
		global.Log.Infof(
			"自动抓取完成: 来源=%d 扫描=%d 新增=%d 更新=%d 重复=%d 失败=%d 跳过=%d",
			result.SourceTotal,
			result.LatestScanned,
			result.Created,
			result.UpdatedCount,
			result.DuplicateCount,
			result.FailedCount,
			result.Skipped,
		)
	}
}

func ensureCrawlerAccount() (models.UserModel, error) {
	userName := strings.TrimSpace(global.Config.SiteInfo.CrawlerUser)
	if userName == "" {
		userName = defaultCrawlerUser
	}
	nickName := strings.TrimSpace(global.Config.SiteInfo.CrawlerNick)
	if nickName == "" {
		nickName = defaultCrawlerNick
	}

	var user models.UserModel
	err := global.DB.Where("user_name = ?", userName).Take(&user).Error
	if err == nil {
		return user, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return user, err
	}

	password := random2.RandString(24) + "Aa1!"
	createErr := user_ser.UserService{}.CreateUser(
		userName,
		nickName,
		password,
		ctype.PermissionUser,
		"",
		"127.0.0.1",
	)
	if createErr != nil {
		return user, createErr
	}
	if err = global.DB.Where("user_name = ?", userName).Take(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func fetchFengfengArticleLinks() ([]string, error) {
	merged := make([]string, 0, 512)
	seen := make(map[string]struct{}, 512)
	appendLinks := func(list []string) {
		for _, link := range list {
			canonical := canonicalizeArticleLink(link)
			if canonical == "" {
				continue
			}
			if _, ok := seen[canonical]; ok {
				continue
			}
			seen[canonical] = struct{}{}
			merged = append(merged, canonical)
		}
	}

	if apiLinks, err := fetchArticleLinksFromAPI(); err == nil && len(apiLinks) > 0 {
		appendLinks(apiLinks)
	} else if err != nil {
		global.Log.Warnf("从 articles API 拉取文章链接失败，尝试 sitemap: %v", err)
	}

	if sitemapLinks, err := fetchArticleLinksFromSitemap(fengfengSitemapURL); err == nil && len(sitemapLinks) > 0 {
		appendLinks(sitemapLinks)
	} else if err != nil {
		global.Log.Warnf("从 sitemap 拉取文章链接失败，尝试页面兜底: %v", err)
	}

	fallbackLinks, fallbackErr := fetchArticleLinksFromHome()
	if fallbackErr != nil {
		if len(merged) > 0 {
			return merged, nil
		}
		return nil, fmt.Errorf("API/sitemap均失败且页面兜底失败: %w", fallbackErr)
	}
	appendLinks(fallbackLinks)
	if len(merged) == 0 {
		return nil, errors.New("未发现可用文章链接")
	}
	return merged, nil
}

func buildCrawlCandidates(links []string, limit int) []crawlCandidate {
	if limit <= 0 {
		limit = len(links)
	}

	seen := make(map[string]struct{}, len(links))
	candidates := make([]crawlCandidate, 0, limit)

	for _, rawLink := range links {
		canonicalLink := canonicalizeArticleLink(rawLink)
		if canonicalLink == "" {
			continue
		}
		articleID, err := extractArticleID(canonicalLink)
		if err != nil || strings.TrimSpace(articleID) == "" {
			continue
		}
		uniqueKey := "id:" + articleID
		if _, ok := seen[uniqueKey]; ok {
			continue
		}
		seen[uniqueKey] = struct{}{}
		seen["link:"+canonicalLink] = struct{}{}

		candidates = append(candidates, crawlCandidate{
			Link:      canonicalLink,
			ArticleID: articleID,
		})
		if len(candidates) >= limit {
			break
		}
	}
	return candidates
}

func normalizeArticleScanLimit(limit int, sourceTotal int) int {
	if sourceTotal <= 0 {
		return 0
	}
	if limit < 0 {
		return defaultArticleScanLimit
	}
	if limit == 0 {
		return sourceTotal
	}
	if limit > maxArticleScanLimit {
		return maxArticleScanLimit
	}
	if limit > sourceTotal {
		return sourceTotal
	}
	return limit
}

type sitemapURLSet struct {
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

type sitemapIndex struct {
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

func fetchArticleLinksFromSitemap(sitemapURL string) ([]string, error) {
	body, err := fetchURLContent(sitemapURL)
	if err != nil {
		return nil, err
	}

	index := sitemapIndex{}
	if unmarshalErr := xml.Unmarshal(body, &index); unmarshalErr == nil && len(index.Sitemaps) > 0 {
		seen := map[string]struct{}{}
		links := make([]string, 0, 64)
		for _, item := range index.Sitemaps {
			childURL := strings.TrimSpace(item.Loc)
			if childURL == "" {
				continue
			}
			childLinks, childErr := fetchArticleLinksFromSitemap(childURL)
			if childErr != nil {
				global.Log.Warnf("读取子 sitemap 失败: %s err=%v", childURL, childErr)
				continue
			}
			for _, link := range childLinks {
				if _, ok := seen[link]; ok {
					continue
				}
				seen[link] = struct{}{}
				links = append(links, link)
			}
		}
		if len(links) > 0 {
			return links, nil
		}
	}

	urlSet := sitemapURLSet{}
	if err = xml.Unmarshal(body, &urlSet); err != nil {
		return nil, err
	}

	baseURL, _ := url.Parse(fengfengHomeURL)
	seen := map[string]struct{}{}
	links := make([]string, 0, len(urlSet.URLs))
	for _, item := range urlSet.URLs {
		normalized := normalizeURL(baseURL, item.Loc)
		if normalized == "" {
			continue
		}
		path := strings.ToLower(strings.TrimSpace(mustParseURL(normalized).Path))
		if !looksLikeArticlePath(path) {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		links = append(links, normalized)
	}
	if len(links) == 0 {
		return nil, errors.New("sitemap未发现有效文章链接")
	}
	return links, nil
}

func fetchArticleLinksFromHome() ([]string, error) {
	doc, err := fetchDocument(fengfengHomeURL)
	if err != nil {
		return nil, err
	}

	baseURL, _ := url.Parse(fengfengHomeURL)
	seen := map[string]struct{}{}
	links := make([]string, 0, 32)

	doc.Find("a[href]").Each(func(_ int, selection *goquery.Selection) {
		href, ok := selection.Attr("href")
		if !ok {
			return
		}
		normalized := normalizeURL(baseURL, href)
		if normalized == "" {
			return
		}
		path := strings.ToLower(strings.TrimSpace(mustParseURL(normalized).Path))
		if !looksLikeArticlePath(path) {
			return
		}
		if _, ok = seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		links = append(links, normalized)
	})

	return links, nil
}

type articleDetail struct {
	Title    string
	Abstract string
	Content  string
	CoverURL string
}

type fengfengArticleAPIResponse struct {
	Code int    `json:"code"`
	Data string `json:"data"`
	Msg  string `json:"msg"`
}

type fengfengArticleListResponse struct {
	Count int                      `json:"count"`
	List  []map[string]interface{} `json:"list"`
}

func fetchArticleLinksFromAPI() ([]string, error) {
	seen := make(map[string]struct{}, 512)
	links := make([]string, 0, 512)
	totalCount := -1

	for page := 1; page <= maxFengfengListPages; page++ {
		pageData, err := fetchArticleListPageFromAPI(page, fengfengListPageSize)
		if err != nil {
			if page == 1 {
				return nil, err
			}
			break
		}
		if totalCount < 0 {
			totalCount = pageData.Count
		}
		if len(pageData.List) == 0 {
			break
		}

		for _, item := range pageData.List {
			link := extractArticleLinkFromListItem(item)
			if link == "" {
				continue
			}
			if _, ok := seen[link]; ok {
				continue
			}
			seen[link] = struct{}{}
			links = append(links, link)
		}

		if len(pageData.List) < fengfengListPageSize {
			break
		}
		if totalCount > 0 && len(links) >= totalCount {
			break
		}
	}

	if len(links) == 0 {
		return nil, errors.New("articles API 未返回可用文章链接")
	}
	return links, nil
}

func fetchArticlePreviewMapFromAPI(limit int) (map[string]CrawlArticleCandidate, error) {
	previewMap := make(map[string]CrawlArticleCandidate, 256)
	totalCount := -1
	collected := 0
	if limit <= 0 {
		limit = fengfengListPageSize
	}

	for page := 1; page <= maxFengfengListPages; page++ {
		pageData, err := fetchArticleListPageFromAPI(page, fengfengListPageSize)
		if err != nil {
			if page == 1 {
				return nil, err
			}
			break
		}
		if totalCount < 0 {
			totalCount = pageData.Count
		}
		if len(pageData.List) == 0 {
			break
		}

		for _, item := range pageData.List {
			preview := extractArticlePreviewFromListItem(item)
			if preview.ArticleID == "" || preview.Link == "" {
				continue
			}
			if _, ok := previewMap[preview.ArticleID]; ok {
				continue
			}
			previewMap[preview.ArticleID] = preview
			collected++
			if collected >= limit {
				return previewMap, nil
			}
		}

		if len(pageData.List) < fengfengListPageSize {
			break
		}
		if totalCount > 0 && collected >= totalCount {
			break
		}
	}

	if len(previewMap) == 0 {
		return nil, errors.New("articles API 未返回可用文章摘要")
	}
	return previewMap, nil
}

func fetchArticleListPageFromAPI(page int, limit int) (fengfengArticleListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = fengfengListPageSize
	}
	apiURL, err := url.Parse(fengfengListAPIURL)
	if err != nil {
		return fengfengArticleListResponse{}, err
	}
	query := apiURL.Query()
	query.Set("page", strconv.Itoa(page))
	query.Set("limit", strconv.Itoa(limit))
	apiURL.RawQuery = query.Encode()

	body, err := fetchURLContent(apiURL.String())
	if err != nil {
		return fengfengArticleListResponse{}, err
	}
	var resp fengfengArticleAPIResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return fengfengArticleListResponse{}, err
	}
	if resp.Code != 0 {
		return fengfengArticleListResponse{}, fmt.Errorf("articles API 返回失败 code=%d msg=%s", resp.Code, resp.Msg)
	}
	if strings.TrimSpace(resp.Data) == "" {
		return fengfengArticleListResponse{}, errors.New("articles API data为空")
	}

	decryptedJSON, err := decryptFengfengPayload(resp.Data)
	if err != nil {
		return fengfengArticleListResponse{}, err
	}

	var payload fengfengArticleListResponse
	if err = json.Unmarshal([]byte(decryptedJSON), &payload); err != nil {
		return fengfengArticleListResponse{}, err
	}
	return payload, nil
}

func extractArticleLinkFromListItem(item map[string]interface{}) string {
	if item == nil {
		return ""
	}
	if value, ok := pickFieldValueByKeys(item, "link", "url", "href"); ok {
		link := normalizeURL(mustParseURL(fengfengHomeURL), interfaceToString(value))
		if link != "" {
			return link
		}
	}

	if value, ok := pickFieldValueByKeys(item, "id", "article_id", "articleId"); ok {
		articleID := strings.TrimSpace(interfaceToString(value))
		if articleID == "" {
			return ""
		}
		return canonicalizeArticleLink(fengfengHomeURL + "article/" + articleID)
	}
	return ""
}

func extractArticlePreviewFromListItem(item map[string]interface{}) CrawlArticleCandidate {
	preview := CrawlArticleCandidate{
		Link: extractArticleLinkFromListItem(item),
	}
	if preview.Link != "" {
		preview.ArticleID, _ = extractArticleID(preview.Link)
	}
	if preview.ArticleID == "" {
		if value, ok := pickFieldValueByKeys(item, "id", "article_id", "articleId"); ok {
			preview.ArticleID = strings.TrimSpace(interfaceToString(value))
		}
	}

	titleValue, _ := pickFieldValueByKeys(item,
		"title",
		"article_title",
		"articleTitle",
		"name",
	)
	abstractValue, _ := pickFieldValueByKeys(item,
		"abstract",
		"summary",
		"description",
		"desc",
		"intro",
		"sub_title",
		"subTitle",
	)
	coverValue, _ := pickFieldValueByKeys(item,
		"cover",
		"cover_url",
		"coverUrl",
		"banner",
		"banner_url",
		"bannerUrl",
		"image",
		"image_url",
		"imageUrl",
		"thumb",
		"thumbnail",
	)

	preview.Title = sanitizeArticleTitle(interfaceToString(titleValue))
	preview.Abstract = interfaceToString(abstractValue)
	preview.CoverURL = normalizeURL(mustParseURL(fengfengHomeURL), interfaceToString(coverValue))
	return preview
}

func fetchArticleDetail(articleURL string) (articleDetail, error) {
	detail, err := fetchArticleDetailFromAPI(articleURL)
	if err == nil && strings.TrimSpace(detail.Title) != "" && strings.TrimSpace(detail.Content) != "" {
		return detail, nil
	}
	if err != nil {
		global.Log.Warnf("API抓取失败，回退HTML解析: %s err=%v", articleURL, err)
	}
	return fetchArticleDetailFromPage(articleURL)
}

func fetchArticleDetailFromAPI(articleURL string) (articleDetail, error) {
	articleID, err := extractArticleID(articleURL)
	if err != nil {
		return articleDetail{}, err
	}

	apiURL := fmt.Sprintf(fengfengAPIURLFmt, url.PathEscape(articleID))
	body, err := fetchURLContent(apiURL)
	if err != nil {
		return articleDetail{}, err
	}

	var resp fengfengArticleAPIResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return articleDetail{}, err
	}
	if resp.Code != 0 {
		return articleDetail{}, fmt.Errorf("远端返回失败 code=%d msg=%s", resp.Code, resp.Msg)
	}
	if strings.TrimSpace(resp.Data) == "" {
		return articleDetail{}, errors.New("远端返回内容为空")
	}

	decryptedJSON, err := decryptFengfengPayload(resp.Data)
	if err != nil {
		return articleDetail{}, err
	}
	return parseFengfengArticleJSON(decryptedJSON)
}

func fetchArticleDetailFromPage(articleURL string) (articleDetail, error) {
	doc, err := fetchDocument(articleURL)
	if err != nil {
		return articleDetail{}, err
	}

	title := strings.TrimSpace(doc.Find("h1").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("title").First().Text())
	}
	title = strings.ReplaceAll(title, "| 枫枫知道", "")
	title = strings.TrimSpace(title)

	contentSelection := firstNonEmptySelection(doc,
		"article .post-content",
		".article-content",
		".post-content",
		".entry-content",
		"article",
		"main",
		".content",
	)

	contentHTML, _ := contentSelection.Html()
	contentMarkdown := strings.TrimSpace(contentSelection.Text())
	if strings.TrimSpace(contentHTML) != "" {
		converter := md.NewConverter("", true, nil)
		if markdown, convertErr := converter.ConvertString(contentHTML); convertErr == nil {
			contentMarkdown = strings.TrimSpace(markdown)
		}
	}
	if contentMarkdown == "" {
		contentMarkdown = strings.TrimSpace(doc.Find("body").Text())
	}

	abs := []rune(strings.TrimSpace(contentSelection.Text()))
	abstract := ""
	if len(abs) > 0 {
		if len(abs) > 120 {
			abstract = string(abs[:120])
		} else {
			abstract = string(abs)
		}
	}

	coverURL := ""
	doc.Find("img[src]").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		src, ok := selection.Attr("src")
		if !ok {
			return true
		}
		normalized := normalizeURL(mustParseURL(articleURL), src)
		if normalized == "" {
			return true
		}
		coverURL = normalized
		return false
	})

	return articleDetail{
		Title:    title,
		Abstract: abstract,
		Content:  contentMarkdown,
		CoverURL: coverURL,
	}, nil
}

func parseFengfengArticleJSON(rawJSON string) (articleDetail, error) {
	decoder := json.NewDecoder(strings.NewReader(rawJSON))
	decoder.UseNumber()

	payload := map[string]interface{}{}
	if err := decoder.Decode(&payload); err != nil {
		return articleDetail{}, err
	}

	titleValue, _ := pickFieldValueByKeys(payload,
		"title",
		"article_title",
		"articleTitle",
		"name",
	)
	abstractValue, _ := pickFieldValueByKeys(payload,
		"abstract",
		"summary",
		"description",
		"desc",
		"intro",
		"sub_title",
		"subTitle",
	)
	contentValue, _ := pickFieldValueByKeys(payload,
		"content",
		"article_content",
		"articleContent",
		"markdown",
		"md",
		"html",
		"body",
		"detail",
		"text",
	)
	coverValue, _ := pickFieldValueByKeys(payload,
		"cover",
		"cover_url",
		"coverUrl",
		"banner",
		"banner_url",
		"bannerUrl",
		"image",
		"image_url",
		"imageUrl",
		"thumb",
		"thumbnail",
	)

	title := sanitizeArticleTitle(interfaceToString(titleValue))
	abstract := interfaceToString(abstractValue)
	content := normalizeArticleContentValue(contentValue)
	coverURL := strings.TrimSpace(interfaceToString(coverValue))
	coverURL = normalizeURL(mustParseURL(fengfengHomeURL), coverURL)
	if title == "" || content == "" {
		return articleDetail{}, errors.New("解析API正文失败：标题或正文为空")
	}
	if abstract == "" {
		abstract = buildAbstractFromText(content, 120)
	}

	return articleDetail{
		Title:    title,
		Abstract: abstract,
		Content:  content,
		CoverURL: coverURL,
	}, nil
}

func extractArticleID(articleURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(articleURL))
	if err != nil {
		return "", err
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for idx := 0; idx < len(segments)-1; idx++ {
		if strings.EqualFold(segments[idx], "article") {
			articleID := strings.TrimSpace(segments[idx+1])
			if articleID != "" {
				return articleID, nil
			}
		}
	}
	return "", fmt.Errorf("无法从URL中提取文章ID: %s", articleURL)
}

func decryptFengfengPayload(encryptedText string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encryptedText))
	if err != nil {
		return "", fmt.Errorf("密文base64解码失败: %w", err)
	}
	if len(cipherBytes) == 0 || len(cipherBytes)%aes.BlockSize != 0 {
		return "", fmt.Errorf("密文长度非法: %d", len(cipherBytes))
	}

	block, err := aes.NewCipher([]byte(fengfengDecryptKey))
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(fengfengDecryptKey))
	iv := []byte(hex.EncodeToString(hash[:]))
	if len(iv) < aes.BlockSize {
		return "", errors.New("IV长度不足")
	}
	iv = iv[:aes.BlockSize]

	plain := make([]byte, len(cipherBytes))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, cipherBytes)
	plain = trimAESPadding(plain)
	intermediate := strings.TrimSpace(string(plain))
	if intermediate == "" {
		return "", errors.New("AES解密结果为空")
	}

	decoded, err := base64.StdEncoding.DecodeString(intermediate)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(intermediate)
		if err != nil {
			return "", fmt.Errorf("二层base64解码失败: %w", err)
		}
	}
	return strings.TrimSpace(string(decoded)), nil
}

func trimAESPadding(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	padLen := int(data[len(data)-1])
	if padLen > 0 && padLen <= aes.BlockSize && padLen <= len(data) {
		valid := true
		for _, b := range data[len(data)-padLen:] {
			if int(b) != padLen {
				valid = false
				break
			}
		}
		if valid {
			data = data[:len(data)-padLen]
		}
	}
	return bytes.TrimRight(data, "\x00")
}

func pickFieldValueByKeys(data map[string]interface{}, keys ...string) (interface{}, bool) {
	for _, key := range keys {
		target := normalizeFieldKey(key)
		for fieldKey, fieldValue := range data {
			if normalizeFieldKey(fieldKey) == target {
				return fieldValue, true
			}
		}
	}
	for _, key := range keys {
		if value, ok := findFieldValueByKey(data, normalizeFieldKey(key)); ok {
			return value, true
		}
	}
	return nil, false
}

func findFieldValueByKey(data interface{}, targetKey string) (interface{}, bool) {
	switch node := data.(type) {
	case map[string]interface{}:
		for key, value := range node {
			if normalizeFieldKey(key) == targetKey {
				return value, true
			}
		}
		for _, value := range node {
			if text, ok := findFieldValueByKey(value, targetKey); ok {
				return text, true
			}
		}
	case []interface{}:
		for _, value := range node {
			if text, ok := findFieldValueByKey(value, targetKey); ok {
				return text, true
			}
		}
	}
	return nil, false
}

func interfaceToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return v.String()
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return strings.TrimSpace(fmt.Sprintf("%f", v))
	case float32:
		if v == float32(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return strings.TrimSpace(fmt.Sprintf("%f", v))
	case int:
		return fmt.Sprintf("%d", v)
	case int8:
		return fmt.Sprintf("%d", v)
	case int16:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case uint:
		return fmt.Sprintf("%d", v)
	case uint8:
		return fmt.Sprintf("%d", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			text := interfaceToString(item)
			if text == "" {
				continue
			}
			parts = append(parts, text)
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	default:
		return ""
	}
}

func normalizeFieldKey(key string) string {
	normalized := strings.TrimSpace(strings.ToLower(key))
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	return normalized
}

func sanitizeArticleTitle(title string) string {
	normalized := strings.TrimSpace(title)
	normalized = strings.ReplaceAll(normalized, "| 枫枫知道", "")
	return strings.TrimSpace(normalized)
}

func normalizeArticleContentValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return normalizeRawContent(v)
	case map[string]interface{}, []interface{}:
		richText := collectRichText(value, 0)
		return normalizeRawContent(richText)
	default:
		return normalizeRawContent(interfaceToString(v))
	}
}

func normalizeRawContent(content string) string {
	text := strings.TrimSpace(content)
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\u00a0", " ")
	if strings.Contains(text, "\\n") && !strings.Contains(text, "\n") {
		text = strings.ReplaceAll(text, "\\n", "\n")
	}
	text = normalizeEscapedMarkdown(text)

	if maybeJSONText(text) {
		if decoded := parseEmbeddedJSONContent(text); decoded != "" {
			return collapseEmptyLines(decoded, 2)
		}
	}
	if strings.Contains(text, "<") && strings.Contains(text, ">") {
		converter := md.NewConverter("", true, nil)
		if markdown, err := converter.ConvertString(text); err == nil && strings.TrimSpace(markdown) != "" {
			return collapseEmptyLines(markdown, 2)
		}
	}
	return collapseEmptyLines(text, 2)
}

func normalizeEscapedMarkdown(text string) string {
	if text == "" {
		return ""
	}
	if !strings.Contains(text, "\\`") &&
		!strings.Contains(text, "\\#") &&
		!strings.Contains(text, "\\[") &&
		!strings.Contains(text, "\\*") &&
		!strings.Contains(text, "\\_") {
		return text
	}
	replacer := strings.NewReplacer(
		"\\`", "`",
		"\\#", "#",
		"\\[", "[",
		"\\]", "]",
		"\\(", "(",
		"\\)", ")",
		"\\*", "*",
		"\\_", "_",
		"\\-", "-",
		"\\+", "+",
	)
	return replacer.Replace(text)
}

func maybeJSONText(text string) bool {
	trimmed := strings.TrimSpace(text)
	return (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
}

func parseEmbeddedJSONContent(content string) string {
	decoder := json.NewDecoder(strings.NewReader(content))
	decoder.UseNumber()
	var payload interface{}
	if err := decoder.Decode(&payload); err != nil {
		return ""
	}
	return collectRichText(payload, 0)
}

func collectRichText(data interface{}, depth int) string {
	if depth > 32 {
		return ""
	}

	switch node := data.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(node)
	case []interface{}:
		parts := make([]string, 0, len(node))
		for _, item := range node {
			text := collectRichText(item, depth+1)
			if strings.TrimSpace(text) == "" {
				continue
			}
			parts = append(parts, strings.TrimSpace(text))
		}
		return strings.Join(parts, "\n\n")
	case map[string]interface{}:
		nodeType := normalizeFieldKey(interfaceToString(node["type"]))
		text := strings.TrimSpace(interfaceToString(node["text"]))
		if text == "" {
			text = strings.TrimSpace(interfaceToString(node["value"]))
		}

		childText := ""
		for _, key := range []string{"children", "content", "nodes", "items"} {
			if child, ok := node[key]; ok {
				candidate := strings.TrimSpace(collectRichText(child, depth+1))
				if candidate != "" {
					if childText == "" {
						childText = candidate
					} else {
						childText += "\n\n" + candidate
					}
				}
			}
		}

		base := strings.TrimSpace(text)
		if base == "" {
			base = childText
		} else if childText != "" {
			base = strings.TrimSpace(base + "\n\n" + childText)
		}
		if base == "" {
			base = strings.TrimSpace(interfaceToString(node["html"]))
			if base == "" {
				base = strings.TrimSpace(interfaceToString(node["md"]))
			}
		}
		if base == "" {
			return ""
		}
		return formatRichTextByType(nodeType, base)
	default:
		return strings.TrimSpace(interfaceToString(node))
	}
}

func formatRichTextByType(nodeType, text string) string {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return ""
	}
	switch nodeType {
	case "h1", "heading1", "heading":
		return "# " + normalized
	case "h2", "heading2":
		return "## " + normalized
	case "h3", "heading3":
		return "### " + normalized
	case "h4", "heading4":
		return "#### " + normalized
	case "h5", "heading5":
		return "##### " + normalized
	case "h6", "heading6":
		return "###### " + normalized
	case "blockquote", "quote":
		lines := strings.Split(normalized, "\n")
		for idx, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				lines[idx] = ">"
			} else {
				lines[idx] = "> " + trimmed
			}
		}
		return strings.Join(lines, "\n")
	case "codeline", "codeblock", "code":
		return "```\n" + normalized + "\n```"
	case "orderedlistitem", "olitem", "numberedlistitem":
		return "1. " + normalized
	case "unorderedlistitem", "listitem", "bulletlistitem":
		return "- " + normalized
	default:
		return normalized
	}
}

func collapseEmptyLines(content string, maxConsecutive int) string {
	if maxConsecutive < 1 {
		maxConsecutive = 1
	}
	lines := strings.Split(content, "\n")
	var builder strings.Builder
	emptyCount := 0
	for idx, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if strings.TrimSpace(trimmed) == "" {
			emptyCount++
			if emptyCount > maxConsecutive {
				continue
			}
		} else {
			emptyCount = 0
		}
		builder.WriteString(trimmed)
		if idx < len(lines)-1 {
			builder.WriteByte('\n')
		}
	}
	return strings.TrimSpace(builder.String())
}

func buildAbstractFromText(content string, maxLen int) string {
	text := strings.TrimSpace(content)
	if text == "" {
		return ""
	}
	if strings.Contains(text, "<") && strings.Contains(text, ">") {
		if doc, err := goquery.NewDocumentFromReader(strings.NewReader(text)); err == nil {
			text = strings.TrimSpace(doc.Text())
		}
	}
	text = strings.Join(strings.Fields(text), " ")
	runes := []rune(text)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return text
}

func fetchDocument(rawURL string) (*goquery.Document, error) {
	body, err := fetchURLContent(rawURL)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytes.NewReader(body))
}

func fetchURLContent(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; gvb-crawler/1.0)")
	req.Header.Set("Accept", "application/json,text/html,application/xml,*/*")
	req.Header.Set("Referer", fengfengHomeURL)

	client := http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("请求失败: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func firstNonEmptySelection(doc *goquery.Document, selectors ...string) *goquery.Selection {
	for _, selector := range selectors {
		sel := doc.Find(selector).First()
		if strings.TrimSpace(sel.Text()) != "" {
			return sel
		}
	}
	return doc.Find("body").First()
}

func normalizeURL(base *url.URL, href string) string {
	value := strings.TrimSpace(href)
	if value == "" || strings.HasPrefix(value, "javascript:") || strings.HasPrefix(value, "#") {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if !parsed.IsAbs() {
		parsed = base.ResolveReference(parsed)
	}
	host := strings.ToLower(parsed.Host)
	if !strings.Contains(host, "fengfengzhidao.com") {
		return ""
	}
	parsed.Fragment = ""
	return canonicalizeArticleLink(parsed.String())
}

func canonicalizeArticleLink(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Host))
	if host == "" || !strings.Contains(host, "fengfengzhidao.com") {
		return ""
	}
	parsed.Host = host
	parsed.Fragment = ""
	parsed.RawQuery = ""

	path := strings.TrimSpace(parsed.Path)
	if path == "" {
		path = "/"
	}
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}
	parsed.Path = path
	return parsed.String()
}

func looksLikeArticlePath(path string) bool {
	return strings.Contains(path, "/article/") ||
		strings.Contains(path, "/archives/") ||
		strings.Contains(path, "/post/")
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		return &url.URL{}
	}
	return u
}

func existsArticleByTitle(title string) bool {
	if global.ESClient == nil {
		return false
	}
	return models.ArticleModel{Title: title}.ISExistData()
}

func findExistingArticleDocID(link, articleID string) string {
	if global.ESClient == nil {
		return ""
	}

	query := buildExistingArticleQuery(link, articleID)
	if query == nil {
		return ""
	}

	result, err := global.ESClient.Search(models.ArticleModel{}.Index()).
		Query(query).
		Size(1).
		Do(context.Background())
	if err != nil || result.Hits == nil || len(result.Hits.Hits) == 0 {
		return ""
	}
	return result.Hits.Hits[0].Id
}

func buildExistingArticleQuery(link, articleID string) elastic.Query {
	boolQuery := elastic.NewBoolQuery()
	hasCondition := false

	for _, item := range buildLinkVariants(link) {
		boolQuery.Should(elastic.NewTermQuery("link", item))
		hasCondition = true
	}

	normalizedID := strings.TrimSpace(articleID)
	if normalizedID != "" {
		boolQuery.Should(elastic.NewWildcardQuery("link", fmt.Sprintf("*%s*", "/article/"+normalizedID)))
		hasCondition = true
	}

	if !hasCondition {
		return nil
	}
	boolQuery.MinimumNumberShouldMatch(1)
	return boolQuery
}

func buildLinkVariants(link string) []string {
	canonicalLink := canonicalizeArticleLink(link)
	if canonicalLink == "" {
		return nil
	}
	linkVariants := map[string]struct{}{
		canonicalLink: {},
	}
	if strings.HasSuffix(canonicalLink, "/") {
		linkVariants[strings.TrimSuffix(canonicalLink, "/")] = struct{}{}
	} else {
		linkVariants[canonicalLink+"/"] = struct{}{}
	}

	result := make([]string, 0, len(linkVariants))
	for item := range linkVariants {
		result = append(result, item)
	}
	return result
}

func existsArticleByFengfengID(articleID string) bool {
	return findExistingArticleDocID("", articleID) != ""
}

func existsArticleByLink(link string) bool {
	return findExistingArticleDocID(link, "") != ""
}

func updateExistingArticleByDocID(docID string, candidate crawlCandidate, detail articleDetail) error {
	if global.ESClient == nil {
		return errors.New("ES未初始化")
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	updatePayload := map[string]interface{}{
		"title":      detail.Title,
		"keyword":    detail.Title,
		"abstract":   detail.Abstract,
		"content":    detail.Content,
		"link":       candidate.Link,
		"source":     fengfengHomeURL,
		"updated_at": now,
	}
	if boardID, boardName := resolveDefaultCrawlBoard(); boardID > 0 {
		updatePayload["board_id"] = boardID
		updatePayload["board_name"] = boardName
		updatePayload["category"] = boardName
	}
	if strings.TrimSpace(detail.CoverURL) != "" {
		category := inferImageCategory(strings.Join([]string{detail.Title, detail.CoverURL}, " "))
		if banner, _, err := downloadRemoteImageToBanner(detail.CoverURL, category); err == nil && banner.ID > 0 {
			updatePayload["banner_id"] = banner.ID
			updatePayload["banner_url"] = banner.Path
		}
	}
	_, err := global.ESClient.Update().
		Index(models.ArticleModel{}.Index()).
		Id(docID).
		Doc(updatePayload).
		Do(context.Background())
	if err != nil {
		return err
	}

	if rate, duplicateID, duplicateTitle, duplicateErr := calculateArticleDuplicateRate(docID, detail.Title, detail.Content); duplicateErr == nil {
		_ = es_ser.ArticleUpdate(docID, map[string]any{
			"duplicate_rate":         rate,
			"duplicate_target_id":    duplicateID,
			"duplicate_target_title": duplicateTitle,
		})
	}
	return nil
}

func resolveDefaultCrawlBoard() (uint, string) {
	defaultName := "技术精选"
	if err := board_ser.EnsureDefaultBoards(); err != nil {
		return 0, defaultName
	}

	var board models.BoardModel
	err := global.DB.
		Where("is_enabled = ?", true).
		Order("sort asc").
		First(&board).Error
	if err != nil {
		return 0, defaultName
	}
	if strings.TrimSpace(board.Name) == "" {
		return board.ID, defaultName
	}
	return board.ID, strings.TrimSpace(board.Name)
}

func randomBanner() (uint, string) {
	var banners []models.BannerModel
	if err := global.DB.Select("id", "path").Find(&banners).Error; err != nil || len(banners) == 0 {
		return 0, ""
	}
	index := time.Now().UnixNano() % int64(len(banners))
	banner := banners[index]
	return banner.ID, banner.Path
}
