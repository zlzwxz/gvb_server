package social_api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"gvb-server/utils"

	"github.com/gin-gonic/gin"
)

// FileUploadView 上传好友系统文件，允许任意格式。
func (SocialApi) FileUploadView(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	if file.Size <= 0 {
		res.FailWithMessage("文件不能为空", c)
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
	var existing models.SocialFileModel
	if err = global.DB.Take(&existing, "hash = ? AND user_id = ?", fileHash, claims.UserID).Error; err == nil {
		res.OkWithData(socialFileUploadItem{
			ID:   existing.ID,
			Name: existing.Name,
			URL:  fmt.Sprintf("/api/social/files/%d/download", existing.ID),
			Size: existing.Size,
			Ext:  existing.Ext,
			Mime: existing.Mime,
		}, c)
		return
	}

	fileName := strings.TrimSpace(file.Filename)
	if fileName == "" {
		fileName = fmt.Sprintf("file_%d", time.Now().Unix())
	}
	baseDir := filepath.Join(global.Config.Upload.Path, "im", fmt.Sprintf("%d", claims.UserID))
	if err = os.MkdirAll(baseDir, 0755); err != nil {
		res.FailWithMessage("创建文件目录失败", c)
		return
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	baseName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	if baseName == "" {
		baseName = "file"
	}
	safeName := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_").Replace(baseName)
	newFileName := fmt.Sprintf("%s_%s%s", safeName, time.Now().Format("20060102150405"), filepath.Ext(fileName))
	localPath := filepath.Join(baseDir, newFileName)
	if err = os.WriteFile(localPath, byteData, 0644); err != nil {
		res.FailWithMessage("保存文件失败", c)
		return
	}

	record := models.SocialFileModel{
		UserID: claims.UserID,
		Name:   fileName,
		Path:   "/" + filepath.ToSlash(localPath),
		Size:   file.Size,
		Ext:    ext,
		Mime:   guessMimeType(fileName),
		Hash:   fileHash,
	}
	if err = global.DB.Create(&record).Error; err != nil {
		res.FailWithMessage("文件入库失败", c)
		return
	}
	res.OkWithData(socialFileUploadItem{
		ID:   record.ID,
		Name: record.Name,
		URL:  fmt.Sprintf("/api/social/files/%d/download", record.ID),
		Size: record.Size,
		Ext:  record.Ext,
		Mime: record.Mime,
	}, c)
}

// FileDownloadView 下载好友系统文件。
func (SocialApi) FileDownloadView(c *gin.Context) {
	var uri socialFileURI
	if err := c.ShouldBindUri(&uri); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	claims := getClaims(c)
	var record models.SocialFileModel
	if err := global.DB.Take(&record, uri.ID).Error; err != nil {
		res.FailWithMessage("文件不存在", c)
		return
	}
	if record.UserID != claims.UserID {
		var count int64
		groupIDs := getCurrentUserGroupIDs(claims.UserID)
		query := global.DB.Model(&models.SocialMessageModel{}).Where("file_id = ?", record.ID)
		query = query.Where(
			"(conversation_type = ? AND (send_user_id = ? OR receive_user_id = ?)) OR (conversation_type = ? AND group_id IN ?)",
			string(models.SocialConversationDirect), claims.UserID, claims.UserID,
			string(models.SocialConversationGroup), groupIDs,
		)
		query.Count(&count)
		if count == 0 {
			res.FailWithMessage("无权下载该文件", c)
			return
		}
	}
	localPath := filepath.Clean(strings.TrimPrefix(record.Path, "/"))
	if strings.HasPrefix(localPath, "..") {
		res.FailWithMessage("非法文件路径", c)
		return
	}
	if _, err := os.Stat(localPath); err != nil {
		res.FailWithMessage("文件不存在", c)
		return
	}
	uploadRoot := filepath.Clean(global.Config.Upload.Path)
	if !strings.HasPrefix(filepath.ToSlash(localPath), filepath.ToSlash(uploadRoot)+"/") {
		res.FailWithMessage("文件路径非法", c)
		return
	}
	c.FileAttachment(localPath, record.Name)
}

func getCurrentUserGroupIDs(userID uint) []uint {
	var memberships []models.SocialGroupMemberModel
	global.DB.Select("group_id").Where("user_id = ?", userID).Find(&memberships)
	result := make([]uint, 0, len(memberships))
	for _, item := range memberships {
		result = append(result, item.GroupID)
	}
	return dedupeUintSlice(result)
}
