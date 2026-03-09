package es_ser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"

	"github.com/olivere/elastic/v7"
)

const (
	defaultArticleBackupPath = "backup/es/article_index_backup.json"
	esScrollBatchSize        = 200
	esBulkBatchSize          = 200
	backupFormatVersion      = 1
)

// ESDocumentBackupItem 描述备份文件中的一条 ES 文档。
// `ID` 保存 ES 文档主键，`Source` 保存原始 JSON 内容，便于原样恢复。
type ESDocumentBackupItem struct {
	ID     string          `json:"id"`
	Source json.RawMessage `json:"source"`
}

// ESIndexBackupPayload 是任意索引导出后的通用备份文件结构。
type ESIndexBackupPayload struct {
	Version     int                    `json:"version"`
	ExportedAt  string                 `json:"exported_at"`
	Index       string                 `json:"index"`
	Count       int                    `json:"count"`
	Documents   []ESDocumentBackupItem `json:"documents"`
	Description string                 `json:"description"`
}

// 下面两个别名保留给命令行导入导出逻辑使用，避免已有代码名义上“全失效”。
type ArticleBackupItem = ESDocumentBackupItem
type ArticleBackupPayload = ESIndexBackupPayload

// ResolveArticleBackupPath 统一处理备份文件路径。
func ResolveArticleBackupPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return filepath.Clean(defaultArticleBackupPath)
	}
	return filepath.Clean(path)
}

// ExportArticleBackup 将文章索引完整导出到本地 JSON 文件。
func ExportArticleBackup(path string) error {
	if err := ensureESReady(); err != nil {
		return err
	}
	documents, err := loadAllArticleDocuments()
	if err != nil {
		return err
	}

	payload := buildIndexBackupPayload(
		models.ArticleModel{}.Index(),
		"article_index backup exported by gvb_server",
		documents,
	)

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 ES 备份文件失败: %w", err)
	}

	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建备份目录失败: %w", err)
	}
	if err = os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写入备份文件失败: %w", err)
	}
	return nil
}

// ImportArticleBackup 从备份文件恢复文章索引，并重新生成全文搜索索引。
func ImportArticleBackup(path string) error {
	if err := ensureESReady(); err != nil {
		return err
	}
	payload, err := readArticleBackup(path)
	if err != nil {
		return err
	}

	articleModel := models.ArticleModel{}
	if err = articleModel.CreateIndex(); err != nil {
		return fmt.Errorf("重建文章索引失败: %w", err)
	}
	if err = bulkRestoreArticleDocuments(payload.Documents); err != nil {
		return err
	}
	if err = rebuildFullTextFromBackup(payload.Documents); err != nil {
		return err
	}
	return nil
}

// SyncFullTextFromArticleIndex 直接读取当前文章索引，重新构建全文索引。
func SyncFullTextFromArticleIndex() error {
	if err := ensureESReady(); err != nil {
		return err
	}
	documents, err := loadAllArticleDocuments()
	if err != nil {
		return err
	}
	return rebuildFullTextFromBackup(documents)
}

func ensureESReady() error {
	if global.ESClient == nil {
		return errors.New("ES 客户端未初始化，请先检查配置并启动依赖")
	}
	return nil
}

func readArticleBackup(path string) (*ArticleBackupPayload, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 ES 备份文件失败: %w", err)
	}

	var payload ArticleBackupPayload
	if err = json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("解析 ES 备份文件失败: %w", err)
	}
	if payload.Version == 0 {
		return nil, errors.New("ES 备份文件缺少版本号，无法确认格式")
	}
	if len(payload.Documents) == 0 {
		global.Log.Warning("备份文件中没有文章数据，将恢复为空索引")
	}
	return &payload, nil
}

func loadAllArticleDocuments() ([]ArticleBackupItem, error) {
	indexName := models.ArticleModel{}.Index()
	return loadAllIndexDocuments(indexName)
}

func loadAllIndexDocuments(indexName string) ([]ESDocumentBackupItem, error) {
	indexName = strings.TrimSpace(indexName)
	if indexName == "" {
		return nil, errors.New("索引名称不能为空")
	}

	exists, err := global.ESClient.IndexExists(indexName).Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("检查索引 %s 是否存在失败: %w", indexName, err)
	}
	if !exists {
		return nil, fmt.Errorf("索引 %s 不存在，无法执行该命令", indexName)
	}

	scroll := global.ESClient.Scroll(indexName).Size(esScrollBatchSize)
	documents := make([]ESDocumentBackupItem, 0, esScrollBatchSize)
	for {
		result, scrollErr := scroll.Do(context.Background())
		if errors.Is(scrollErr, io.EOF) {
			break
		}
		if scrollErr != nil {
			return nil, fmt.Errorf("滚动读取索引 %s 失败: %w", indexName, scrollErr)
		}
		for _, hit := range result.Hits.Hits {
			sourceCopy := append(json.RawMessage(nil), hit.Source...)
			documents = append(documents, ESDocumentBackupItem{
				ID:     hit.Id,
				Source: sourceCopy,
			})
		}
	}
	return documents, nil
}

func bulkRestoreArticleDocuments(documents []ArticleBackupItem) error {
	if len(documents) == 0 {
		return nil
	}

	indexName := models.ArticleModel{}.Index()
	bulk := global.ESClient.Bulk().Index(indexName).Refresh("true")
	for idx, item := range documents {
		if strings.TrimSpace(item.ID) == "" {
			return fmt.Errorf("第 %d 条备份数据缺少文档 ID", idx+1)
		}
		if len(item.Source) == 0 {
			return fmt.Errorf("第 %d 条备份数据缺少 source 内容", idx+1)
		}
		bulk.Add(elastic.NewBulkIndexRequest().Id(item.ID).Doc(item.Source))

		if bulk.NumberOfActions() >= esBulkBatchSize {
			if _, err := bulk.Do(context.Background()); err != nil {
				return fmt.Errorf("批量恢复文章索引失败: %w", err)
			}
			bulk = global.ESClient.Bulk().Index(indexName).Refresh("true")
		}
	}

	if bulk.NumberOfActions() == 0 {
		return nil
	}
	if _, err := bulk.Do(context.Background()); err != nil {
		return fmt.Errorf("批量恢复文章索引失败: %w", err)
	}
	return nil
}

func rebuildFullTextFromBackup(documents []ArticleBackupItem) error {
	fullTextModel := models.FullTextModel{}
	if err := fullTextModel.CreateIndex(); err != nil {
		return fmt.Errorf("重建全文索引失败: %w", err)
	}
	if len(documents) == 0 {
		return nil
	}

	indexName := models.FullTextModel{}.Index()
	bulk := global.ESClient.Bulk().Index(indexName).Refresh("true")
	for idx, item := range documents {
		article, err := decodeArticleBackupItem(item)
		if err != nil {
			return fmt.Errorf("解析第 %d 条文章备份失败: %w", idx+1, err)
		}
		if !allowFullTextRebuild(article) {
			continue
		}

		searchDataList := GetSearchIndexDataByContent(article.ID, article.Title, article.Content)
		for _, searchData := range searchDataList {
			bulk.Add(elastic.NewBulkIndexRequest().Doc(searchData))
			if bulk.NumberOfActions() >= esBulkBatchSize {
				if _, err = bulk.Do(context.Background()); err != nil {
					return fmt.Errorf("批量写入全文索引失败: %w", err)
				}
				bulk = global.ESClient.Bulk().Index(indexName).Refresh("true")
			}
		}
	}

	if bulk.NumberOfActions() == 0 {
		return nil
	}
	if _, err := bulk.Do(context.Background()); err != nil {
		return fmt.Errorf("批量写入全文索引失败: %w", err)
	}
	return nil
}

func decodeArticleBackupItem(item ArticleBackupItem) (models.ArticleModel, error) {
	var article models.ArticleModel
	if err := json.Unmarshal(item.Source, &article); err != nil {
		return article, err
	}
	article.ID = item.ID
	return article, nil
}

func allowFullTextRebuild(article models.ArticleModel) bool {
	if article.ID == "" || strings.TrimSpace(article.Title) == "" || strings.TrimSpace(article.Content) == "" {
		return false
	}
	if article.IsPrivate {
		return false
	}
	return article.ReviewStatus == ctype.ArticleReviewApproved || article.ReviewStatus == ctype.ArticleReviewLegacy
}

func buildIndexBackupPayload(indexName, description string, documents []ESDocumentBackupItem) ESIndexBackupPayload {
	return ESIndexBackupPayload{
		Version:     backupFormatVersion,
		ExportedAt:  time.Now().Format("2006-01-02 15:04:05"),
		Index:       strings.TrimSpace(indexName),
		Count:       len(documents),
		Documents:   documents,
		Description: strings.TrimSpace(description),
	}
}
