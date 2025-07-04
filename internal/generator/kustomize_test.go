package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FIXME: more tests
func TestKustomize(t *testing.T) {
	bytes, err := renderKustomize("testdata/kustomize")
	assert.Nil(t, err)
	assert.NotNil(t, bytes)
}
