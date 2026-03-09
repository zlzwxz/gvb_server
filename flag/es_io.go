package flag

import (
	"path/filepath"

	"gvb-server/global"
	"gvb-server/service/es_ser"
)

// EsExport 将文章索引导出到本地 JSON 备份文件。
func EsExport(path string) {
	resolvedPath := es_ser.ResolveArticleBackupPath(path)
	if err := es_ser.ExportArticleBackup(resolvedPath); err != nil {
		global.Log.Fatalf("导出 ES 文章数据失败: %v", err)
		return
	}
	global.Log.Infof("ES 文章数据已导出到: %s", filepath.Clean(resolvedPath))
}

// EsImport 从本地 JSON 备份恢复文章索引，并重建全文索引。
func EsImport(path string) {
	resolvedPath := es_ser.ResolveArticleBackupPath(path)
	if err := es_ser.ImportArticleBackup(resolvedPath); err != nil {
		global.Log.Fatalf("导入 ES 文章数据失败: %v", err)
		return
	}
	global.Log.Infof("ES 文章数据恢复完成，备份文件: %s", filepath.Clean(resolvedPath))
}

// EsSyncFullText 仅根据当前文章索引重建全文搜索索引。
func EsSyncFullText() {
	if err := es_ser.SyncFullTextFromArticleIndex(); err != nil {
		global.Log.Fatalf("重建全文索引失败: %v", err)
		return
	}
	global.Log.Info("全文索引重建完成")
}
