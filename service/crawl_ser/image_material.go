package crawl_ser

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "golang.org/x/image/webp"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/utils"
)

const (
	maxSyncImageNum          = 30
	maxImageScanArticleNum   = 20
	maxImagePreviewCandidate = 40
)

type SyncImageOptions struct {
	ImageURLs []string `json:"image_urls"`
	SyncAll   bool     `json:"sync_all"`
}

type CrawlImageCandidate struct {
	URL          string `json:"url"`
	Category     string `json:"category"`
	ArticleID    string `json:"article_id"`
	ArticleTitle string `json:"article_title"`
}

type PreviewImageResult struct {
	SourceTotal    int                   `json:"source_total"`
	LatestScanned  int                   `json:"latest_scanned"`
	SyncLimit      int                   `json:"sync_limit"`
	NewCandidate   int                   `json:"new_candidate"`
	DuplicateCount int                   `json:"duplicate_count"`
	InvalidCount   int                   `json:"invalid_count"`
	Candidates     []CrawlImageCandidate `json:"candidates"`
}

type SyncImageResult struct {
	SourceTotal    int `json:"source_total"`
	LatestScanned  int `json:"latest_scanned"`
	SyncLimit      int `json:"sync_limit"`
	Created        int `json:"created"`
	DuplicateCount int `json:"duplicate_count"`
	FailedCount    int `json:"failed_count"`
	Skipped        int `json:"skipped"`
	SelectedCount  int `json:"selected_count"`
}

func PreviewFengfengImages() (PreviewImageResult, error) {
	syncLock.Lock()
	defer syncLock.Unlock()
	return previewFengfengImagesNoLock()
}

func previewFengfengImagesNoLock() (PreviewImageResult, error) {
	linkList, err := fetchFengfengArticleLinks()
	if err != nil {
		return PreviewImageResult{}, err
	}

	result := PreviewImageResult{
		SourceTotal:   len(linkList),
		SyncLimit:     maxSyncImageNum,
		Candidates:    make([]CrawlImageCandidate, 0),
		LatestScanned: 0,
	}
	if len(linkList) == 0 {
		return result, nil
	}

	articleCandidates := buildCrawlCandidates(linkList, maxImageScanArticleNum)
	imageCandidates := collectFengfengImageCandidates(articleCandidates)
	result.LatestScanned = len(imageCandidates)
	if len(imageCandidates) == 0 {
		return result, nil
	}

	for _, candidate := range imageCandidates {
		if strings.TrimSpace(candidate.URL) == "" {
			result.InvalidCount++
			continue
		}
		if _, exists := findBannerBySourceURL(candidate.URL); exists {
			result.DuplicateCount++
			continue
		}
		result.NewCandidate++
		if len(result.Candidates) < maxImagePreviewCandidate {
			result.Candidates = append(result.Candidates, candidate)
		}
	}
	return result, nil
}

func SyncFengfengImages() (SyncImageResult, error) {
	return SyncFengfengImagesWithOptions(SyncImageOptions{SyncAll: true})
}

func SyncFengfengImagesWithOptions(options SyncImageOptions) (SyncImageResult, error) {
	syncLock.Lock()
	defer syncLock.Unlock()

	preview, err := previewFengfengImagesNoLock()
	if err != nil {
		return SyncImageResult{}, err
	}

	result := SyncImageResult{
		SourceTotal:   preview.SourceTotal,
		LatestScanned: preview.LatestScanned,
		SyncLimit:     maxSyncImageNum,
	}

	urlFilter := make(map[string]struct{}, len(options.ImageURLs))
	for _, raw := range options.ImageURLs {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		urlFilter[value] = struct{}{}
	}
	syncAll := options.SyncAll || len(urlFilter) == 0
	result.SelectedCount = len(urlFilter)
	if syncAll {
		result.SelectedCount = 0
	}

	processed := 0
	for _, candidate := range preview.Candidates {
		if processed >= maxSyncImageNum {
			break
		}
		if !syncAll {
			if _, ok := urlFilter[candidate.URL]; !ok {
				continue
			}
		}
		processed++

		if _, exists := findBannerBySourceURL(candidate.URL); exists {
			result.DuplicateCount++
			result.Skipped++
			continue
		}
		if _, created, downloadErr := downloadRemoteImageToBanner(candidate.URL, candidate.Category); downloadErr != nil {
			result.FailedCount++
			result.Skipped++
			global.Log.Warnf("抓取图片失败: %s err=%v", candidate.URL, downloadErr)
			continue
		} else if created {
			result.Created++
		} else {
			result.DuplicateCount++
			result.Skipped++
		}
	}
	return result, nil
}

func collectFengfengImageCandidates(articleCandidates []crawlCandidate) []CrawlImageCandidate {
	seen := make(map[string]struct{}, len(articleCandidates)*2)
	list := make([]CrawlImageCandidate, 0, len(articleCandidates)*2)

	for _, candidate := range articleCandidates {
		if strings.TrimSpace(candidate.Link) == "" {
			continue
		}
		detail, err := fetchArticleDetail(candidate.Link)
		if err != nil {
			continue
		}
		articleTitle := strings.TrimSpace(detail.Title)
		if articleTitle == "" {
			articleTitle = fmt.Sprintf("文章 %s", candidate.ArticleID)
		}
		imageURLs := extractImageURLsFromContent(detail.Content)
		if strings.TrimSpace(detail.CoverURL) != "" {
			imageURLs = append(imageURLs, detail.CoverURL)
		}
		for _, imageURL := range imageURLs {
			normalized := normalizeImageURL(candidate.Link, imageURL)
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}

			category := inferImageCategory(strings.Join([]string{articleTitle, normalized}, " "))
			list = append(list, CrawlImageCandidate{
				URL:          normalized,
				Category:     category,
				ArticleID:    candidate.ArticleID,
				ArticleTitle: articleTitle,
			})
		}
	}
	return list
}

func normalizeImageURL(articleURL string, rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "data:") || strings.HasPrefix(value, "javascript:") {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if !parsed.IsAbs() {
		base, parseErr := url.Parse(strings.TrimSpace(articleURL))
		if parseErr != nil {
			return ""
		}
		parsed = base.ResolveReference(parsed)
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return ""
	}
	parsed.Fragment = ""
	return parsed.String()
}

var markdownImageRegex = regexp.MustCompile(`!\[[^\]]*]\(([^)\s]+)`)
var htmlImageRegex = regexp.MustCompile(`(?is)<img[^>]+src=["']([^"']+)["']`)

func extractImageURLsFromContent(content string) []string {
	value := strings.TrimSpace(content)
	if value == "" {
		return nil
	}
	seen := map[string]struct{}{}
	list := make([]string, 0, 8)

	for _, match := range markdownImageRegex.FindAllStringSubmatch(value, -1) {
		if len(match) < 2 {
			continue
		}
		urlValue := strings.TrimSpace(match[1])
		if urlValue == "" {
			continue
		}
		if _, ok := seen[urlValue]; ok {
			continue
		}
		seen[urlValue] = struct{}{}
		list = append(list, urlValue)
	}

	for _, match := range htmlImageRegex.FindAllStringSubmatch(value, -1) {
		if len(match) < 2 {
			continue
		}
		urlValue := strings.TrimSpace(match[1])
		if urlValue == "" {
			continue
		}
		if _, ok := seen[urlValue]; ok {
			continue
		}
		seen[urlValue] = struct{}{}
		list = append(list, urlValue)
	}
	return list
}

func downloadRemoteImageToBanner(rawURL string, category string) (models.BannerModel, bool, error) {
	cleanURL := strings.TrimSpace(rawURL)
	if cleanURL == "" {
		return models.BannerModel{}, false, fmt.Errorf("图片链接为空")
	}
	if existing, ok := findBannerBySourceURL(cleanURL); ok {
		return existing, false, nil
	}

	body, contentType, err := fetchRemoteBinary(cleanURL)
	if err != nil {
		return models.BannerModel{}, false, err
	}
	if len(body) == 0 {
		return models.BannerModel{}, false, fmt.Errorf("图片内容为空")
	}
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return models.BannerModel{}, false, fmt.Errorf("远端不是图片类型: %s", contentType)
	}
	if _, _, decodeErr := image.DecodeConfig(bytes.NewReader(body)); decodeErr != nil {
		return models.BannerModel{}, false, fmt.Errorf("图片解析失败: %w", decodeErr)
	}

	hash := utils.Md5(body)
	var byHash models.BannerModel
	if err = global.DB.Take(&byHash, "hash = ?", hash).Error; err == nil && byHash.ID > 0 {
		updateMap := map[string]any{}
		if strings.TrimSpace(byHash.SourceURL) == "" {
			updateMap["source_url"] = cleanURL
		}
		if strings.TrimSpace(byHash.ImageCategory) == "" && strings.TrimSpace(category) != "" {
			updateMap["image_category"] = category
		}
		if len(updateMap) > 0 {
			_ = global.DB.Model(&byHash).Updates(updateMap).Error
		}
		return byHash, false, nil
	}

	suffix := resolveImageSuffix(cleanURL, contentType)
	if suffix == "" {
		suffix = "jpg"
	}
	category = strings.TrimSpace(category)
	if category == "" {
		category = inferImageCategory(cleanURL)
	}
	category = sanitizeImageNameSegment(category, "图片")

	basePath := filepath.Clean(global.Config.Upload.Path)
	saveDir := filepath.Join(basePath, "materials", category)
	if err = os.MkdirAll(saveDir, 0755); err != nil {
		return models.BannerModel{}, false, err
	}

	fileName := fmt.Sprintf("%s_%s_%d.%s", category, time.Now().Format("20060102150405"), rand.Intn(100000), suffix)
	targetPath := filepath.Join(saveDir, fileName)
	if err = os.WriteFile(targetPath, body, 0644); err != nil {
		return models.BannerModel{}, false, err
	}

	dbPath := "/" + filepath.ToSlash(targetPath)
	model := models.BannerModel{
		Path:          dbPath,
		Hash:          hash,
		Name:          fileName,
		ImageType:     ctype.Local,
		SourceURL:     cleanURL,
		ImageCategory: category,
	}
	if err = global.DB.Create(&model).Error; err != nil {
		return models.BannerModel{}, false, err
	}
	return model, true, nil
}

func findBannerBySourceURL(sourceURL string) (models.BannerModel, bool) {
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return models.BannerModel{}, false
	}
	var model models.BannerModel
	if err := global.DB.Take(&model, "source_url = ?", sourceURL).Error; err != nil {
		return models.BannerModel{}, false
	}
	return model, true
}

func fetchRemoteBinary(rawURL string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; gvb-crawler/1.0)")
	req.Header.Set("Accept", "image/*,*/*")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("图片请求失败: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	return body, contentType, nil
}

func resolveImageSuffix(rawURL string, contentType string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(strings.Split(rawURL, "?")[0]), "."))
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp":
		if ext == "jpeg" {
			return "jpg"
		}
		return ext
	}
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.Contains(contentType, "png"):
		return "png"
	case strings.Contains(contentType, "gif"):
		return "gif"
	case strings.Contains(contentType, "webp"):
		return "webp"
	default:
		return "jpg"
	}
}

func sanitizeImageNameSegment(value string, fallback string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "/", "_")
	value = strings.ReplaceAll(value, "\\", "_")
	value = strings.Trim(value, "._-")
	if value == "" {
		return fallback
	}
	if len([]rune(value)) > 24 {
		value = string([]rune(value)[:24])
	}
	return value
}

var imageCategoryRules = map[string][]string{
	"天空": {"sky", "cloud", "sunset", "sunrise", "星空", "蓝天", "云"},
	"海洋": {"ocean", "sea", "beach", "pool", "swim", "海", "泳池", "水"},
	"AI": {"ai", "人工智能", "机器学习", "模型", "robot", "github", "代码", "程序"},
	"城市": {"city", "street", "building", "城市", "街道", "建筑"},
	"自然": {"nature", "forest", "mountain", "草地", "山", "森林"},
}

func inferImageCategory(text string) string {
	value := strings.ToLower(strings.TrimSpace(text))
	for category, keywords := range imageCategoryRules {
		for _, keyword := range keywords {
			if strings.Contains(value, strings.ToLower(keyword)) {
				return category
			}
		}
	}
	return "通用"
}
