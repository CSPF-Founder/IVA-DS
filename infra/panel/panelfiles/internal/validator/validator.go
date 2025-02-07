package validator

import (
	"html"
	"regexp"
	"strings"
)

func IsValidUsername(username string) bool {
	if username != "" {
		match, _ := regexp.MatchString("^[A-Za-z][a-zA-Z0-9_]{1,30}$", username)
		return match
	}
	return false
}

// SanitizeXss removes any HTML tags and attributes from the input string
func SanitizeXss(input string) string {
	if input != "" {
		filteredInput := html.EscapeString(input)

		replacements := map[string]string{
			"(": "&#40;",
			")": "&#41;",
			"+": "&#43;",
			"{": "&#123;",
			"}": "&#125;",
			"[": "&#91;",
			"]": "&#93;",
		}

		for char, replacement := range replacements {
			filteredInput = strings.ReplaceAll(filteredInput, char, replacement)
		}
		return filteredInput
	}
	return ""
}
