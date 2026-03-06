package file_api

import (
	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/res"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type fileDownloadURI struct {
	ID uint `uri:"id" binding:"required"`
}

// FileDownloadView 下载文章附件
func (FileApi) FileDownloadView(c *gin.Context) {
	var cr fileDownloadURI
	if err := c.ShouldBindUri(&cr); err != nil {
		res.FailWithCode(res.ArgumentError, c)
		return
	}

	var record models.ArticleFileModel
	if err := global.DB.Take(&record, cr.ID).Error; err != nil {
		res.FailWithMessage("附件不存在", c)
		return
	}

	localPath := strings.TrimPrefix(record.Path, "/")
	localPath = filepath.Clean(localPath)
	if strings.HasPrefix(localPath, "..") {
		res.FailWithMessage("非法附件路径", c)
		return
	}
	if _, err := os.Stat(localPath); err != nil {
		res.FailWithMessage("附件文件不存在", c)
		return
	}

	c.FileAttachment(localPath, record.Name)
}
