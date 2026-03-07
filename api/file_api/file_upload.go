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

// FileUploadItem 是附件上传成功后返回给前端的精简结构。
// 前端通常只关心：附件 ID、原始名称、下载地址、大小和扩展名。
type FileUploadItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
	Ext  string `json:"ext"`
}

// 文件附件白名单。
// 这里和图片上传分开，是因为文章附件允许 PDF、Office、压缩包等非图片文件。
var fileSuffixWhiteList = []string{
	"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx",
	"txt", "zip", "rar", "7z", "csv", "md",
}

// FileUploadView 上传文章附件。
func (FileApi) FileUploadView(c *gin.Context) {
	// 第一步：从请求体里取出 `file` 字段。
	file, err := c.FormFile("file")
	if err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	// 第二步：拿到当前登录用户身份。
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	if claims.Role == int(ctype.PermissionVisitor) {
		res.FailWithMessage("游客用户不可上传附件", c)
		return
	}

	// 第三步：检查文件名和扩展名。
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

	// 第四步：限制上传体积。
	sizeMB := float64(file.Size) / float64(1024*1024)
	if sizeMB >= float64(global.Config.Upload.Max_Size) {
		res.FailWithMessage(fmt.Sprintf("文件大小超过限制，当前 %.2fMB，限制 %dMB", sizeMB, global.Config.Upload.Max_Size), c)
		return
	}

	// 第五步：把文件完整读进来，用于后续计算哈希和写盘。
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

	// 第六步：按“文件内容哈希 + 用户 ID”做去重。
	// 同一个用户重复上传同一份文件时，直接复用旧记录。
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

	// 第七步：构造安全目录。
	// 附件统一放到 `upload.path/attachments/<用户昵称>` 下面。
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

	// 第八步：生成新的本地文件名，避免重名覆盖。
	baseName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	if baseName == "" {
		baseName = "file"
	}
	newFileName := fmt.Sprintf("%s_%s.%s", baseName, time.Now().Format("20060102150405"), suffix)
	localPath := filepath.Join(baseDir, newFileName)

	// 第九步：把附件写到磁盘。
	if err = os.WriteFile(localPath, byteData, 0644); err != nil {
		res.FailWithMessage("保存附件失败", c)
		return
	}

	// 第十步：把 URL 风格路径记录进数据库。
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

	// 返回给前端的是“下载接口地址”，而不是原始静态直链。
	// 这样可以继续在下载接口里做权限控制。
	res.OkWithData(FileUploadItem{
		ID:   record.ID,
		Name: record.Name,
		URL:  fmt.Sprintf("/api/files/%d/download", record.ID),
		Size: record.Size,
		Ext:  record.Ext,
	}, c)
}
