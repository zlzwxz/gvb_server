package images_api

import (
	"fmt"
	"path/filepath"
	"regexp"
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
	for _, prefix := range imageOwnerPathPrefixes(claims) {
		if strings.HasPrefix(image.Path, prefix) {
			return true
		}
	}
	return false
}

func imageOwnerPathPrefixes(claims *jwts.CustomClaims) []string {
	basePath := strings.Trim(global.Config.Upload.Path, "/")
	if basePath == "" {
		return nil
	}

	prefixes := []string{
		fmt.Sprintf("/%s/u_%d/", basePath, claims.UserID),
	}

	if old := strings.TrimSuffix(imageOwnerPathLike(claims.NickName), "%"); old != "" {
		prefixes = append(prefixes, old)
	}

	safeNick := sanitizePathSegment(claims.NickName, "")
	if safeNick != "" && safeNick != strings.TrimSpace(claims.NickName) {
		prefixes = append(prefixes, fmt.Sprintf("/%s/%s/", basePath, safeNick))
	}
	return prefixes
}

var invalidPathSegmentRegex = regexp.MustCompile(`[^\p{Han}\p{L}\p{N}_-]+`)

func sanitizePathSegment(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.ReplaceAll(trimmed, "/", "_")
	trimmed = strings.ReplaceAll(trimmed, "\\", "_")
	trimmed = invalidPathSegmentRegex.ReplaceAllString(trimmed, "_")
	trimmed = strings.Trim(trimmed, "._-")
	if trimmed == "" {
		return fallback
	}
	if len(trimmed) > 48 {
		trimmed = trimmed[:48]
	}
	return filepath.ToSlash(trimmed)
}
