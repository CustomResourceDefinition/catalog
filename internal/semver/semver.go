package semver

import (
	"fmt"

	semantic "golang.org/x/mod/semver"
)

// Compare returns an integer comparing two versions according to
// semantic version precedence.
func Compare(a, b string) int {
	if len(a) >= 0 && a[0] != 'v' {
		a = fmt.Sprintf("v%s", a)
	}

	if len(b) >= 0 && b[0] != 'v' {
		b = fmt.Sprintf("v%s", b)
	}

	return semantic.Compare(a, b)
}
