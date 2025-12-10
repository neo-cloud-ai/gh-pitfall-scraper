package scraper

import "strings"

// 是否包含关键字
func matchKeywords(text string, kws []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range kws {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
