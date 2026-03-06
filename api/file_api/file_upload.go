package file_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/utils"
	"gvb-server/utils/jwts"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type FileUploadItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
	Ext  string `json:"ext"`
}

var fileSuffixWhiteList = []string{
	"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx",
	"txt", "zip", "rar", "7z", "csv", "md",
}

// FileUploadView 上传文章附件
func (FileApi) FileUploadView(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	if claims.Role == int(ctype.PermissionVisitor) {
		res.FailWithMessage("游客用户不可上传附件", c)
		return
	}

	fileName := strings.TrimSpace(file.Filename)
	if fileName == "" {
		res.FailWithMessage("文件名不能为空", c)
		return
	}

	suffix := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	if !utils.InList(suffix, fileSuffixWhiteList) {
		res.FailWithMessage("不支持的文件类型", c)
		return
	}

	sizeMB := float64(file.Size) / float64(1024*1024)
	if sizeMB >= float64(global.Config.Upload.Max_Size) {
		res.FailWithMessage(fmt.Sprintf("文件大小超过限制，当前 %.2fMB，限制 %dMB", sizeMB, global.Config.Upload.Max_Size), c)
		return
	}

	src, err := file.Open()
	if err != nil {
		res.FailWithMessage("文件读取失败", c)
		return
	}
	defer src.Close()

	byteData, err := io.ReadAll(src)
	if err != nil {
		res.FailWithMessage("文件读取失败", c)
		return
	}

	fileHash := utils.Md5(byteData)
	var existing models.ArticleFileModel
	if err = global.DB.Take(&existing, "hash = ? and user_id = ?", fileHash, claims.UserID).Error; err == nil {
		res.OkWithData(FileUploadItem{
			ID:   existing.ID,
			Name: existing.Name,
			URL:  fmt.Sprintf("/api/files/%d/download", existing.ID),
			Size: existing.Size,
			Ext:  existing.Ext,
		}, c)
		return
	}

	safeNick := strings.ReplaceAll(strings.TrimSpace(claims.NickName), "/", "_")
	safeNick = strings.ReplaceAll(safeNick, "\\", "_")
	if safeNick == "" {
		safeNick = "user"
	}
	baseDir := filepath.Join(global.Config.Upload.Path, "attachments", safeNick)
	if err = os.MkdirAll(baseDir, 0755); err != nil {
		res.FailWithMessage("创建附件目录失败", c)
		return
	}

	baseName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	if baseName == "" {
		baseName = "file"
	}
	newFileName := fmt.Sprintf("%s_%s.%s", baseName, time.Now().Format("20060102150405"), suffix)
	localPath := filepath.Join(baseDir, newFileName)

	if err = os.WriteFile(localPath, byteData, 0644); err != nil {
		res.FailWithMessage("保存附件失败", c)
		return
	}

	urlPath := "/" + filepath.ToSlash(localPath)
	record := models.ArticleFileModel{
		UserID:       claims.UserID,
		UserNickName: claims.NickName,
		Name:         fileName,
		Hash:         fileHash,
		Path:         urlPath,
		Size:         file.Size,
		Ext:          suffix,
	}
	if err = global.DB.Create(&record).Error; err != nil {
		res.FailWithMessage("附件入库失败", c)
		return
	}

	res.OkWithData(FileUploadItem{
		ID:   record.ID,
		Name: record.Name,
		URL:  fmt.Sprintf("/api/files/%d/download", record.ID),
		Size: record.Size,
		Ext:  record.Ext,
	}, c)
}
