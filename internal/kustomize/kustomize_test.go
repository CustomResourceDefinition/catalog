package kustomize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKustomize(t *testing.T) {
	bytes, err := Render("testdata")
	assert.Nil(t, err)
	assert.NotNil(t, bytes)
}
