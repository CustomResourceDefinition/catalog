package genall

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenallRender(t *testing.T) {
	expected, err := os.ReadFile("testdata/expected.yaml")
	assert.Nil(t, err)

	bytes, _ := Render("testdata/source")

	assert.Equal(t, expected, bytes)
}

func TestGenallRenderWithNonExistingInput(t *testing.T) {
	bytes, _ := Render("testdata/does-not-exist")
	assert.Nil(t, bytes)
}

func TestGenallRenderWithEmptyInput(t *testing.T) {
	bytes, _ := Render("testdata/empty")
	assert.Nil(t, bytes)
}
