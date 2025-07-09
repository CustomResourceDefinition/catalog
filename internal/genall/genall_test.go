package genall

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenAllCrd(t *testing.T) {
	expected, err := os.ReadFile("testdata/genall/expected.yaml")
	assert.Nil(t, err)

	bytes, _ := Render("testdata/genall/source")

	assert.Equal(t, expected, bytes)
}

func TestGenAllCrdWithNonExistingInput(t *testing.T) {
	bytes, _ := Render("testdata/genall/does-not-exist")
	assert.Nil(t, bytes)
}

func TestGenAllCrdWithEmptyInput(t *testing.T) {
	bytes, _ := Render("testdata/genall/empty")
	assert.Nil(t, bytes)
}
