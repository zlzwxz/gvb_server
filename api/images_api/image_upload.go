package images_api

import (
	"fmt"
	"gvb-server/global"
	"gvb-server/models/res"
	"gvb-server/service/image_ser"
	"gvb-server/utils"
	"gvb-server/utils/jwts"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type FileUploadResponse struct {
	FileName  string `json:"file_name"`  // 文件名
	IsSuccess bool   `json:"is_success"` // 是否上传成功
	Msg       string `json:"msg"`        // 消息
}

func (ImagesApi) ImageUploadView(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		global.Log.Error(err)
		res.FailWithMessage("参数校验失败", c)
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
	// 判断文件大小
	size := float64(file.Size) / float64(1024*1024)
	if size >= float64(global.Config.Upload.Max_Size) {
		msg := fmt.Sprintf("图片大小超过设定大小，当前大小为:%.2fMB， 设定大小为：%dMB ", size, global.Config.Upload.Max_Size)
		res.FailWithMessage(msg, c)
		return
	}

	// 创建这个文件夹 /uploads/file/nick_name
	dirList, err := os.ReadDir(basePath)
	if err != nil {
		res.FailWithMessage("文件目录不存在", c)
		return
	}
	if !isInDirEntry(dirList, claims.NickName) {
		// 创建这个目录
		err := os.Mkdir(path.Join(basePath, claims.NickName), 077)
		if err != nil {
			global.Log.Error(err)
		}
	}
	// 1.如果存在重名，就加随机字符串 时间戳
	// 2.直接+时间戳
	now := time.Now().Format("20060102150405")
	fileName = nameList[0] + "_" + now + "." + suffix
	filePath := path.Join(basePath, claims.NickName, fileName)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		res.FailWithMessage(err.Error(), c)
		return
	}

	res.OkWithData("/"+filePath, c)
	return

}

func isInDirEntry(dirList []os.DirEntry, name string) bool {
	for _, entry := range dirList {
		if entry.Name() == name && entry.IsDir() {
			return true
		}
	}
	return false
}
