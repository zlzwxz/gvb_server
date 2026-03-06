package images_api

import (
	"fmt"
	"io"
	"os"
	"path"
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
	file, err := c.FormFile("image")
	if err != nil {
		global.Log.Error(err)
		res.FailWithCode(res.ArgumentError, c)
		return
	}
	_claims, _ := c.Get("claims")
	claims := _claims.(*jwts.CustomClaims)
	if claims.Role == 3 {
		res.FailWithMessage("游客用户不可上传图片", c)
		return
	}

	fileName := file.Filename
	basePath := global.Config.Upload.Path
	nameList := strings.Split(fileName, ".")
	suffix := strings.ToLower(nameList[len(nameList)-1])
	if !utils.InList(suffix, image_ser.WhiteImageList) {
		res.FailWithMessage("非法文件", c)
		return
	}

	// 判断文件大小，避免超大图片直接把存储和带宽打满。
	size := float64(file.Size) / float64(1024*1024)
	if size >= float64(global.Config.Upload.Max_Size) {
		msg := fmt.Sprintf("图片大小超过设定大小，当前大小为:%.2fMB， 设定大小为：%dMB ", size, global.Config.Upload.Max_Size)
		res.FailWithMessage(msg, c)
		return
	}

	fileObj, err := file.Open()
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("文件打开失败", c)
		return
	}
	byteData, err := io.ReadAll(fileObj)
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("文件读取失败", c)
		return
	}
	imageHash := utils.Md5(byteData)

	var bannerModel models.BannerModel
	err = global.DB.Take(&bannerModel, "hash = ?", imageHash).Error
	if err == nil {
		if canOperateImage(claims, bannerModel) {
			res.OkWithData(bannerModel.Path, c)
			return
		}
	}

	dirList, err := os.ReadDir(basePath)
	if err != nil {
		res.FailWithMessage("文件目录不存在", c)
		return
	}
	if !isInDirEntry(dirList, claims.NickName) {
		err := os.Mkdir(path.Join(basePath, claims.NickName), 0755)
		if err != nil {
			global.Log.Error(err)
			res.FailWithMessage("创建目录失败", c)
			return
		}
	}

	now := time.Now().Format("20060102150405")
	fileName = nameList[0] + "_" + now + "." + suffix
	filePath := path.Join(basePath, claims.NickName, fileName)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}

	err = global.DB.Create(&models.BannerModel{
		Path:      "/" + filePath,
		Hash:      imageHash,
		Name:      fileName,
		ImageType: ctype.Local,
	}).Error
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("图片入库失败", c)
		return
	}

	res.OkWithData("/"+filePath, c)
}

// isInDirEntry 判断目标目录是否已经存在。
func isInDirEntry(dirList []os.DirEntry, name string) bool {
	for _, entry := range dirList {
		if entry.Name() == name && entry.IsDir() {
			return true
		}
	}
	return false
}
