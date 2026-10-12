// Package search builds the patterns repositories use for case-insensitive substring search.
package search

import "strings"

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// ContainsPattern escapes LIKE wildcards with a backslash, so pair it with ESCAPE '\' and a literal
// % or _ in the search matches literally. Blank text returns "", meaning no filter.
func ContainsPattern(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}

	return "%" + likeEscaper.Replace(trimmed) + "%"
}
