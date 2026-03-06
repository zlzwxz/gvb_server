package images_api

import (
	"fmt"
	"strings"

	"gvb-server/global"
	"gvb-server/models"
	"gvb-server/models/ctype"
	"gvb-server/utils/jwts"
)

func isImageAdmin(claims *jwts.CustomClaims) bool {
	return claims.Role == int(ctype.PermissionAdmin)
}

func imageOwnerPathLike(nickName string) string {
	nickName = strings.TrimSpace(nickName)
	if nickName == "" {
		return ""
	}
	basePath := strings.Trim(global.Config.Upload.Path, "/")
	if basePath == "" {
		return ""
	}
	return fmt.Sprintf("/%s/%s/%%", basePath, nickName)
}

func canOperateImage(claims *jwts.CustomClaims, image models.BannerModel) bool {
	if isImageAdmin(claims) {
		return true
	}
	like := imageOwnerPathLike(claims.NickName)
	if like == "" {
		return false
	}
	prefix := strings.TrimSuffix(like, "%")
	return strings.HasPrefix(image.Path, prefix)
}
