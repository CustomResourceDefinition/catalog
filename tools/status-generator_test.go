package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratorOfTestdataResultsInKnownOutput(t *testing.T) {
	out := path.Join(t.TempDir(), "out.md")
	expected, err := os.ReadFile("testdata/expected.md")
	assert.Nil(t, err)

	g := StatusGenerator{
		Current: "testdata/current",
		Remote:  "testdata/remote",
		Out:     out,
	}
	err = g.Run()

	assert.Nil(t, err)
	actual, err := os.ReadFile(out)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}
