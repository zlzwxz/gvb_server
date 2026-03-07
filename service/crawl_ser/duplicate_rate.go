package crawl_ser

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"unicode"

	"github.com/olivere/elastic/v7"

	"gvb-server/global"
	"gvb-server/models"
)

func CalculateArticleDuplicateRate(articleID string, title string, content string) (float64, string, string, error) {
	if global.ESClient == nil {
		return 0, "", "", errors.New("ES未初始化")
	}
	seed := strings.TrimSpace(strings.Join([]string{title, content}, "\n"))
	if seed == "" {
		return 0, "", "", nil
	}

	mltQuery := elastic.NewMoreLikeThisQuery().
		Field("title").
		Field("content").
		LikeText(seed).
		MinDocFreq(1).
		MinTermFreq(1).
		Analyzer("standard")

	boolQuery := elastic.NewBoolQuery().Must(mltQuery)
	if strings.TrimSpace(articleID) != "" {
		boolQuery.MustNot(elastic.NewIdsQuery().Ids(articleID))
	}

	result, err := global.ESClient.
		Search(models.ArticleModel{}.Index()).
		Query(boolQuery).
		Size(8).
		Do(context.Background())
	if err != nil {
		return 0, "", "", err
	}
	if result.Hits == nil || len(result.Hits.Hits) == 0 {
		return 0, "", "", nil
	}

	contentSet := buildTextShingleSet(content, 2)
	titleSet := buildTextShingleSet(title, 2)
	bestRate := 0.0
	bestID := ""
	bestTitle := ""

	for _, hit := range result.Hits.Hits {
		if strings.TrimSpace(hit.Id) == "" {
			continue
		}
		var article models.ArticleModel
		if unmarshalErr := json.Unmarshal(hit.Source, &article); unmarshalErr != nil {
			continue
		}

		contentRate := jaccardRate(contentSet, buildTextShingleSet(article.Content, 2))
		titleRate := jaccardRate(titleSet, buildTextShingleSet(article.Title, 2))
		score := contentRate*0.85 + titleRate*0.15
		if score > bestRate {
			bestRate = score
			bestID = hit.Id
			bestTitle = strings.TrimSpace(article.Title)
		}
	}

	return roundFloat(bestRate*100, 2), bestID, bestTitle, nil
}

func calculateArticleDuplicateRate(articleID string, title string, content string) (float64, string, string, error) {
	return CalculateArticleDuplicateRate(articleID, title, content)
}

func buildTextShingleSet(raw string, n int) map[string]struct{} {
	set := map[string]struct{}{}
	normalized := normalizeDuplicateText(raw)
	if normalized == "" {
		return set
	}
	runes := []rune(normalized)
	if len(runes) <= n {
		set[normalized] = struct{}{}
		return set
	}
	if len(runes) > 5000 {
		runes = runes[:5000]
	}
	for idx := 0; idx <= len(runes)-n; idx++ {
		part := string(runes[idx : idx+n])
		if strings.TrimSpace(part) == "" {
			continue
		}
		set[part] = struct{}{}
	}
	return set
}

func normalizeDuplicateText(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		switch {
		case unicode.Is(unicode.Han, r):
			builder.WriteRune(r)
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(r)
		case unicode.IsSpace(r):
			builder.WriteRune(' ')
		default:
			// ignore punctuation
		}
	}
	normalized := strings.Join(strings.Fields(builder.String()), " ")
	return strings.TrimSpace(normalized)
}

func jaccardRate(a map[string]struct{}, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for key := range a {
		if _, ok := b[key]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union <= 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func roundFloat(value float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(value*pow) / pow
}
