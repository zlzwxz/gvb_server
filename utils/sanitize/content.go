package sanitize

import (
	"net/url"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var strictHTMLPolicy = bluemonday.StrictPolicy()

func CleanMarkdownInput(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	return strictHTMLPolicy.Sanitize(trimmed)
}

func CleanURL(value string, allowRelative bool) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return ""
	}

	if allowRelative && strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return ""
	}
	if parsed.Host == "" {
		return ""
	}
	return parsed.String()
}
