package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionSorting(t *testing.T) {
	var result int

	result = Compare("1.0.0", "1.0.0")
	assert.Equal(t, 0, result)

	result = Compare("2.0.0", "1.0.0")
	assert.Equal(t, 1, result)

	result = Compare("2.1.0", "2.2.0")
	assert.Equal(t, -1, result)

	result = Compare("2.10.0", "2.2.0")
	assert.Equal(t, 1, result)

	result = Compare("2.10.0", "2.1.0")
	assert.Equal(t, 1, result)

	result = Compare("2.10.0", "2.0.0")
	assert.Equal(t, 1, result)
}
