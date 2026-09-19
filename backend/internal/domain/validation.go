package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// htmlTagPattern matches HTML tags like <script>, <div>, <b>, <img ...>, etc.
var htmlTagPattern = regexp.MustCompile(`(?i)<\s*/?\s*[a-z][a-z0-9]*(\s+[^>]*)?>`)

// EnsureNoHTML verifies that the input does not contain HTML tags.
// Normal punctuation such as apostrophes (e.g. O'Connor), hyphens and quotes are preserved.
func EnsureNoHTML(field, value string) error {
	if htmlTagPattern.MatchString(value) || strings.Contains(value, "<script") || strings.Contains(value, "</") {
		return fmt.Errorf("%w: %s cannot contain HTML tags", ErrInvalidInput, field)
	}
	return nil
}
