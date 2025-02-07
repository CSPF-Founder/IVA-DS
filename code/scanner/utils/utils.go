package utils

import "strings"

// SplitStrIntoSlice splits a string into a slice of strings based on a separator.
// It trims the whitespace from each element and ignores empty elements.
func SplitStrIntoSlice(input string, sep string) []string {
	var references []string
	if input != "" {
		splitRefs := strings.Split(input, sep)
		for _, ref := range splitRefs {
			ref = strings.TrimSpace(ref)
			if ref == "" {
				continue
			}
			references = append(references, ref)
		}
	}
	return references
}
