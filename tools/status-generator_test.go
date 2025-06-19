package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAThing(t *testing.T) {
	out := path.Join(t.TempDir(), "out.md")
	g := StatusGenerator{
		Current: "testdata/current",
		Remote:  "testdata/remote",
		Out:     out,
	}
	err := g.Run()
	assert.Nil(t, err, err.Error()) // FIXME: test way more
}
