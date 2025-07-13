package semver

import (
	"fmt"
	"strings"

	semantic "golang.org/x/mod/semver"
)

// Compare returns an integer comparing two versions according to
// semantic version precedence.
func Compare(a, b string) int {
	a = normalize(a)
	b = normalize(b)
	return semantic.Compare(a, b)
}

func IsCanonical(s string) bool {
	s = normalize(s)
	return len(semantic.Build(s)) == 0 && len(semantic.Prerelease(s)) == 0 && strings.Count(s, ".") == 2
}

func normalize(s string) string {
	if len(s) > 0 && s[0] != 'v' {
		return fmt.Sprintf("v%s", s)
	}
	return s
}
