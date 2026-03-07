package routers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gvb-server/global"
	"gvb-server/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

// RouterGroup 把 *gin.RouterGroup 包一层，方便在方法上按业务模块拆分路由注册逻辑。
type RouterGroup struct {
	*gin.RouterGroup
}

// InitRouter 初始化整个 Gin 引擎。
// 这里集中做三件事：设置运行模式、挂公共中间能力、注册所有业务路由。
func InitRouter() *gin.Engine {
	gin.SetMode(global.Config.System.Env)

	router := gin.Default()
	// 统一追加安全响应头，例如禁止某些危险嵌入和 MIME 猜测。
	router.Use(middleware.SecurityHeaders())
	// 限制 multipart 表单最大内存占用，避免上传时把内存直接撑爆。
	// 这里在配置上额外多给 2MB，留一点协议头和边界缓冲空间。
	router.MaxMultipartMemory = int64(global.Config.Upload.Max_Size+2) * 1024 * 1024
	router.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
	// 图片等静态资源通过 `/uploads/*` 暴露；附件目录则额外受保护，不允许直接裸链访问。
	registerUploadStaticRoute(router)

	apiRouterGroup := router.Group("/api")
	// 对业务 API 统一挂操作审计中间件，便于后台记录关键操作日志。
	apiRouterGroup.Use(middleware.OperationAudit())
	routerGroupApp := RouterGroup{RouterGroup: apiRouterGroup}
	routerGroupApp.SettinsRouter()
	routerGroupApp.ImageRouter()
	routerGroupApp.FileRouter()
	routerGroupApp.AdvertRouter()
	routerGroupApp.AnnouncementRouter()
	routerGroupApp.MenuRouter()
	routerGroupApp.BoardRouter()
	routerGroupApp.UserRouter()
	routerGroupApp.SocialRouter()
	routerGroupApp.TagRouter()
	routerGroupApp.MessageRouter()
	routerGroupApp.ArticleRouter()
	routerGroupApp.DiggRouter()
	routerGroupApp.CommentRouter()
	routerGroupApp.NewRouter()
	routerGroupApp.ChatRouter()
	routerGroupApp.LogRouter()
	routerGroupApp.DataRouter()
	return router
}

// registerUploadStaticRoute 负责公开访问上传目录中的“可公开资源”。
// 注意：这里不是把整个 uploads 目录无脑暴露出去，而是会额外拦截受保护子目录。
func registerUploadStaticRoute(router *gin.Engine) {
	router.GET("/uploads/*filepath", func(c *gin.Context) {
		cleanPath, ok := sanitizeUploadRelativePath(c.Param("filepath"))
		if !ok {
			c.Status(http.StatusForbidden)
			return
		}

		normalized := filepath.ToSlash(cleanPath)
		if isProtectedUploadPath(normalized) {
			c.Status(http.StatusForbidden)
			return
		}

		// 最终仍然从固定的公开根目录 `uploads` 下解析，避免访问任意磁盘路径。
		baseDir := uploadStaticBaseDir()
		targetPath := filepath.Join(baseDir, cleanPath)
		absBase, err := filepath.Abs(baseDir)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		absTarget, err := filepath.Abs(targetPath)
		if err != nil || (absTarget != absBase && !strings.HasPrefix(absTarget, absBase+string(os.PathSeparator))) {
			c.Status(http.StatusForbidden)
			return
		}
		if _, err = os.Stat(absTarget); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.File(absTarget)
	})
}

// sanitizeUploadRelativePath 负责把用户传入的 URL 路径转成安全的相对路径。
// 如果路径为空、是当前目录、或者试图以 `..` 回退上级目录，就直接判失败。
func sanitizeUploadRelativePath(rawPath string) (string, bool) {
	cleanPath := filepath.Clean(strings.TrimPrefix(strings.TrimSpace(rawPath), "/"))
	if cleanPath == "." || cleanPath == "" || strings.HasPrefix(cleanPath, "..") {
		return "", false
	}
	return cleanPath, true
}

// isProtectedUploadPath 判断当前访问目标是否属于“虽然在 uploads 下，但不允许直接公开访问”的路径。
func isProtectedUploadPath(normalizedPath string) bool {
	cleanPath := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(normalizedPath)), "./")
	if cleanPath == "." || cleanPath == "" {
		return true
	}

	for _, prefix := range uploadProtectedPrefixes() {
		if cleanPath == prefix || strings.HasPrefix(cleanPath, prefix+"/") {
			return true
		}
	}
	return false
}

// uploadProtectedPrefixes 返回所有需要禁止静态直链的目录前缀。
// 当前最重要的是附件目录：附件下载必须走带权限判断的 API，而不是直接访问磁盘文件。
func uploadProtectedPrefixes() []string {
	protectedLeaf := "attachments"
	relativeUploadPath := uploadPathRelativeToPublicRoot()
	if relativeUploadPath == "" {
		return []string{protectedLeaf}
	}
	return []string{filepath.ToSlash(filepath.Join(relativeUploadPath, protectedLeaf))}
}

// uploadPathRelativeToPublicRoot 用来计算“真正的上传根目录”相对于公开根目录 `uploads` 的相对位置。
// 例如：
// - 公开根：`uploads`
// - 配置上传根：`uploads/file`
// 那么返回值就是 `file`。
func uploadPathRelativeToPublicRoot() string {
	if global.Config == nil {
		return "file"
	}

	publicRoot := filepath.Clean(uploadStaticBaseDir())
	uploadPath := filepath.Clean(strings.TrimSpace(global.Config.Upload.Path))
	if uploadPath == "." || uploadPath == "" {
		return ""
	}

	relativePath, err := filepath.Rel(publicRoot, uploadPath)
	if err != nil {
		return ""
	}
	relativePath = filepath.ToSlash(filepath.Clean(relativePath))
	if relativePath == "." || relativePath == "" || relativePath == ".." || strings.HasPrefix(relativePath, "../") {
		return ""
	}
	return strings.TrimPrefix(relativePath, "./")
}

// uploadStaticBaseDir 是所有静态上传资源的公开根目录。
// 单独抽函数而不是写死字符串，是为了后续如果改公开根目录时更集中。
func uploadStaticBaseDir() string {
	return "uploads"
}
