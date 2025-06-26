package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultipleCRD(t *testing.T) {
	out := t.TempDir()

	u := Updater{
		Input:  "testdata/updater/input",
		Output: out,
	}
	err := u.Run()

	assert.Nil(t, err)

	assertDirectories(t, out, "testdata/updater/output")
}

func assertDirectories(t *testing.T, a, b string) {
	entries, err := os.ReadDir(a)
	assert.Nil(t, err)

	for _, e := range entries {
		aEntry := path.Join(a, e.Name())
		bEntry := path.Join(b, e.Name())

		aStat, err := os.Stat(aEntry)
		assert.Nil(t, err)
		bStat, err := os.Stat(bEntry)
		assert.Nil(t, err)
		assert.Equal(t, aStat.IsDir(), bStat.IsDir())

		if aStat.IsDir() {
			assertDirectories(t, aEntry, bEntry)
		} else {
			aBytes, err := os.ReadFile(aEntry)
			assert.Nil(t, err)
			bBytes, err := os.ReadFile(bEntry)
			assert.Nil(t, err)
			assert.Equal(t, aBytes, bBytes)
		}
	}
}
