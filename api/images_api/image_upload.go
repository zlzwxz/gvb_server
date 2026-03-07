package images_api

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/models/res"
	"gvb-server/service/image_ser"
	"gvb-server/utils"
	"gvb-server/utils/jwts"

	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// FileUploadResponse 文件上传响应结构。
type FileUploadResponse struct {
	FileName  string `json:"file_name" swag:"description:文件名"`
	IsSuccess bool   `json:"is_success" swag:"description:是否上传成功"`
	Msg       string `json:"msg" swag:"description:消息"`
}

// ImageUploadView 上传图片。
// @Summary 上传图片
// @Description 上传图片文件，支持 jpg、png、gif 等格式，并校验大小、后缀和重复文件。
// @Tags 图片管理
// @Accept multipart/form-data
// @Produce json
// @Param token header string true "token"
// @Param image formData file true "图片文件"
// @Success 200 {object} res.Response{data=string} "上传成功，返回文件路径"
// @Failure 400 {object} res.Response "请求错误"
// @Failure 401 {object} res.Response "未授权或游客不可上传"
// @Router /api/images [post]
func (ImagesApi) ImageUploadView(c *gin.Context) {
	// 第一步：从 multipart/form-data 请求里取出名为 `image` 的文件字段。
	file, err := c.FormFile("image")
	if err != nil {
		global.Log.Error(err)
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	// 第二步：从中间件提前塞入的 claims 里拿到当前登录用户身份。
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	if claims.Role == 3 {
		res.FailWithMessage("游客用户不可上传图片", c)
		return
	}

	// 第三步：做基础文件名与路径根目录准备。
	fileName := strings.TrimSpace(file.Filename)
	basePath := filepath.Clean(global.Config.Upload.Path)
	if fileName == "" {
		res.FailWithMessage("文件名不能为空", c)
		return
	}

	// 第四步：先检查扩展名是否在白名单内。
	// 注意：只检查扩展名并不安全，所以后面还会继续检测 MIME 和图片头信息。
	suffix := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	if !utils.InList(suffix, image_ser.WhiteImageList) {
		res.FailWithMessage("非法文件", c)
		return
	}

	// 第五步：按配置检查文件大小，避免超大图片直接把磁盘和带宽打满。
	size := float64(file.Size) / float64(1024*1024)
	if size >= float64(global.Config.Upload.Max_Size) {
		msg := fmt.Sprintf("图片大小超过设定大小，当前大小为:%.2fMB， 设定大小为：%dMB ", size, global.Config.Upload.Max_Size)
		res.FailWithMessage(msg, c)
		return
	}

	// 第六步：读取完整文件内容到内存，后续要做内容级校验和哈希去重。
	fileObj, err := file.Open()
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("文件打开失败", c)
		return
	}
	defer fileObj.Close()

	byteData, err := io.ReadAll(fileObj)
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("文件读取失败", c)
		return
	}
	if len(byteData) == 0 {
		res.FailWithMessage("空文件不可上传", c)
		return
	}

	// 第七步：从文件内容本身判断 MIME，防止“伪装后缀”的假图片。
	contentType := http.DetectContentType(byteData)
	if !strings.HasPrefix(contentType, "image/") {
		res.FailWithMessage("文件内容不是合法图片", c)
		return
	}

	// 第八步：真正尝试解析图片头，确保它确实是一张能被 Go 图像库识别的图片。
	if _, _, err = image.DecodeConfig(bytes.NewReader(byteData)); err != nil {
		res.FailWithMessage("图片解析失败，请检查文件格式", c)
		return
	}

	// 第九步：计算文件哈希，用于去重。
	imageHash := utils.Md5(byteData)

	var bannerModel models.BannerModel
	err = global.DB.Take(&bannerModel, "hash = ?", imageHash).Error
	if err == nil {
		// 如果数据库里已经有同一张图，并且当前用户有权限使用它，直接复用旧路径，不再重复落盘。
		if canOperateImage(claims, bannerModel) {
			res.OkWithData(bannerModel.Path, c)
			return
		}
	}

	// 第十步：为每个用户分配自己的图片子目录，避免不同用户文件混在一起。
	ownerDir := fmt.Sprintf("u_%d", claims.UserID)
	saveDir := filepath.Join(basePath, ownerDir)
	if err = os.MkdirAll(saveDir, 0755); err != nil {
		global.Log.Error(err)
		res.FailWithMessage("创建目录失败", c)
		return
	}

	// 第十一步：生成一个更安全、更稳定的新文件名。
	baseName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	baseName = sanitizePathSegment(baseName, "image")
	now := time.Now().Format("20060102150405")
	newFileName := fmt.Sprintf("%s_%s.%s", baseName, now, suffix)
	filePath := filepath.Join(saveDir, newFileName)

	// 再做一次路径逃逸检查，防止拼接后跑出允许目录之外。
	if !isSubPath(basePath, filePath) {
		res.FailWithMessage("非法文件路径", c)
		return
	}

	// 第十二步：真正把图片写到磁盘。
	if err = os.WriteFile(filePath, byteData, 0644); err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}

	// 数据库存储的路径统一改成 URL 风格，方便前端直接使用。
	dbPath := "/" + filepath.ToSlash(filePath)

	// 第十三步：把图片元信息写入数据库。
	err = global.DB.Create(&models.BannerModel{
		Path:      dbPath,
		Hash:      imageHash,
		Name:      newFileName,
		ImageType: ctype.Local,
	}).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("图片入库失败", c)
		return
	}

	res.OkWithData(dbPath, c)
}

// isSubPath 用来确认目标文件路径是否仍然位于允许的根目录之下。
// 这是防止路径穿越攻击的重要一步。
func isSubPath(basePath string, targetPath string) bool {
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}
	if absBase == absTarget {
		return true
	}
	return strings.HasPrefix(absTarget, absBase+string(os.PathSeparator))
}
