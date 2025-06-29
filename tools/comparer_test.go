package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratorOfTestdataResultsInKnownOutput(t *testing.T) {
	out := path.Join(t.TempDir(), "out.md")
	expected, err := os.ReadFile("testdata/comparer/expected.md")
	assert.Nil(t, err)

	g := Comparer{
		Current:  "testdata/comparer/current",
		Datreeio: "testdata/comparer/remote",
		Ignore:   "testdata/comparer/ignore-status.yaml",
		Out:      out,
	}
	err = g.Run()

	assert.Nil(t, err)
	actual, err := os.ReadFile(out)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

type scenario struct {
	known, remote []item
	ignores       []ignore
	expected      []markdownItem
}

func TestDifferenceCalculation(t *testing.T) {
	tests := []scenario{
		{
			known:    []item{},
			remote:   []item{},
			ignores:  []ignore{},
			expected: []markdownItem{},
		},
		{
			known:    []item{{group: "a", kind: "a", version: "1"}},
			remote:   []item{{group: "a", kind: "a", version: "1"}},
			ignores:  []ignore{},
			expected: []markdownItem{},
		},
		{
			known:    []item{{group: "a", kind: "a", version: "1"}, {group: "a", kind: "a", version: "2"}, {group: "a", kind: "a", version: "3"}},
			remote:   []item{{group: "a", kind: "a", version: "1"}, {group: "a", kind: "a", version: "4"}, {group: "a", kind: "a", version: "5"}},
			ignores:  []ignore{},
			expected: []markdownItem{{Group: "a", Items: []markdownItems{{Kind: "a", Versions: "4, 5"}}}},
		},
		{
			known:    []item{{group: "a", kind: "a", version: "1"}, {group: "a", kind: "b", version: "1"}},
			remote:   []item{{group: "a", kind: "a", version: "1"}, {group: "a", kind: "a", version: "2"}, {group: "a", kind: "b", version: "1"}, {group: "a", kind: "b", version: "2"}, {group: "a", kind: "c", version: "1"}},
			ignores:  []ignore{},
			expected: []markdownItem{{Group: "a", Items: []markdownItems{{Kind: "a", Versions: "2"}, {Kind: "b", Versions: "2"}, {Kind: "c", Versions: "1"}}}},
		},
		{
			known:    []item{},
			remote:   []item{{group: "a", kind: "a", version: "1"}},
			ignores:  []ignore{{Items: []ignoreItem{{Group: "a", Kind: "a", Version: "1"}}}},
			expected: []markdownItem{},
		},
		{
			known:    []item{},
			remote:   []item{{group: "a", kind: "a", version: "1"}},
			ignores:  []ignore{{Items: []ignoreItem{{Group: ".*", Kind: ".*", Version: ".*"}}}},
			expected: []markdownItem{},
		},
	}

	for i, test := range tests {
		data := markdownData{}
		data.update(test.known, test.remote, test.ignores)
		assert.Equal(t, test.expected, data.Data, "index %d failed", i)
	}
}
