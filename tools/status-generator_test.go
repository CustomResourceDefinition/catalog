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

type scenario struct {
	known, remote []item
	expected      []markdownItem
}

func TestDifferenceCalculation(t *testing.T) {
	tests := []scenario{
		{
			known:    []item{},
			remote:   []item{},
			expected: []markdownItem{},
		},
		{
			known:    []item{{group: "a", kind: "a", version: "1"}},
			remote:   []item{{group: "a", kind: "a", version: "1"}},
			expected: []markdownItem{},
		},
		{
			known:    []item{{group: "a", kind: "a", version: "1"}, {group: "a", kind: "a", version: "2"}, {group: "a", kind: "a", version: "3"}},
			remote:   []item{{group: "a", kind: "a", version: "1"}, {group: "a", kind: "a", version: "4"}, {group: "a", kind: "a", version: "5"}},
			expected: []markdownItem{{Group: "a", Items: []markdownItems{{Kind: "a", Versions: "4, 5"}}}},
		},
		{
			known:    []item{{group: "a", kind: "a", version: "1"}, {group: "a", kind: "b", version: "1"}},
			remote:   []item{{group: "a", kind: "a", version: "1"}, {group: "a", kind: "a", version: "2"}, {group: "a", kind: "b", version: "1"}, {group: "a", kind: "b", version: "2"}, {group: "a", kind: "c", version: "1"}},
			expected: []markdownItem{{Group: "a", Items: []markdownItems{{Kind: "a", Versions: "2"}, {Kind: "b", Versions: "2"}, {Kind: "c", Versions: "1"}}}},
		},
	}

	for i, test := range tests {
		data := markdownData{}
		data.update(test.known, test.remote)
		assert.Equal(t, test.expected, data.Data, "index %d failed", i)
	}
}
