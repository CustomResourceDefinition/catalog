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

	result = Compare("2.10.0", "v2.0.0")
	assert.Equal(t, 1, result)

	result = Compare("v2.10.0", "2.0.0")
	assert.Equal(t, 1, result)
}

func TestICanonical(t *testing.T) {
	assert.True(t, IsCanonical("1.0.0"))

	assert.True(t, IsCanonical("2.3.0"))

	assert.True(t, IsCanonical("100.0.0"))

	assert.True(t, IsCanonical("999.999.999"))

	assert.False(t, IsCanonical(""))

	assert.False(t, IsCanonical("1"))

	assert.False(t, IsCanonical("1.0"))

	assert.False(t, IsCanonical("1.0.0-preview1"))

	assert.False(t, IsCanonical("1.0.0+hash"))

	assert.False(t, IsCanonical("1.0.0-preview1+hash"))

	assert.False(t, IsCanonical("1.0.0+hash-preview1"))
}
