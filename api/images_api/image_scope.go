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

	"gorm.io/gorm"
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
	if image.MatchesOwner(claims.UserID, claims.NickName) {
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

func normalizeImageCategory(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	if value == "" {
		return "未分类"
	}
	runes := []rune(value)
	if len(runes) > 24 {
		value = string(runes[:24])
	}
	return value
}

func normalizeImageSort(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "created_at asc":
		return "created_at asc"
	case "name asc":
		return "name asc"
	case "name desc":
		return "name desc"
	case "id asc":
		return "id asc"
	default:
		return "created_at desc"
	}
}

func buildOwnImageWhere(claims *jwts.CustomClaims) (string, []any) {
	conditions := []string{"user_id = ?"}
	args := []any{claims.UserID}
	for _, prefix := range imageOwnerPathPrefixes(claims) {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			continue
		}
		conditions = append(conditions, "path LIKE ?")
		args = append(args, prefix+"%")
	}
	return "(" + strings.Join(conditions, " OR ") + ")", args
}

func publicImageWhere() string {
	return "(visibility = 'public' OR visibility = '' OR visibility IS NULL)"
}

func normalizeImageRows(list []models.BannerModel) []models.BannerModel {
	for index := range list {
		list[index].Visibility = models.NormalizeImageVisibility(list[index].Visibility)
		list[index].ImageCategory = normalizeImageCategory(list[index].ImageCategory)
	}
	return list
}

func applyImageAccessQuery(query *gorm.DB, claims *jwts.CustomClaims, visibility string, mineOnly bool) *gorm.DB {
	visibility = strings.ToLower(strings.TrimSpace(visibility))
	if isImageAdmin(claims) {
		if mineOnly {
			return query.Where("user_id = ?", claims.UserID)
		}
		if visibility == models.ImageVisibilityPrivate {
			return query.Where("visibility = ?", models.ImageVisibilityPrivate)
		}
		if visibility == models.ImageVisibilityPublic {
			return query.Where(publicImageWhere())
		}
		return query
	}

	ownWhere, ownArgs := buildOwnImageWhere(claims)
	if mineOnly {
		return query.Where(ownWhere, ownArgs...)
	}
	if visibility == models.ImageVisibilityPrivate {
		args := append([]any{}, ownArgs...)
		return query.Where(ownWhere, args...).Where("visibility = ?", models.ImageVisibilityPrivate)
	}
	if visibility == models.ImageVisibilityPublic {
		return query.Where(publicImageWhere())
	}

	args := append([]any{}, ownArgs...)
	combined := "(" + publicImageWhere() + " OR " + ownWhere + ")"
	return query.Where(combined, args...)
}
