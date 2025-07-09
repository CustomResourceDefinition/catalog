package command

import (
	"os"
	"path"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/stretchr/testify/assert"
)

// func TestMultipleCRD(t *testing.T) {
// 	out := t.TempDir()

// 	u := Updater{
// 		Configuration: "testdata/updater/input",
// 		Output:        out,
// 	}
// 	err := u.Run()

// 	assert.Nil(t, err)

// 	assertDirectories(t, out, "testdata/updater/output")
// }

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

type testScenario struct {
	configs        []configuration.Configuration
	resolvedAmount int
}

func TestSplittingConfiguration(t *testing.T) {
	tests := []testScenario{
		{
			configs:        []configuration.Configuration{},
			resolvedAmount: 0,
		},
		{
			configs: []configuration.Configuration{
				{},
			},
			resolvedAmount: 1,
		},
		{
			configs: []configuration.Configuration{
				{},
				{},
			},
			resolvedAmount: 2,
		},
		{
			configs: []configuration.Configuration{
				{
					Entries: []string{},
				},
				{},
			},
			resolvedAmount: 2,
		},
		{
			configs: []configuration.Configuration{
				{
					Entries: []string{""},
				},
				{},
			},
			resolvedAmount: 2,
		},
		{
			configs: []configuration.Configuration{
				{
					Entries: []string{"one", "two", "three"},
				},
				{
					Entries: []string{"four"},
				},
			},
			resolvedAmount: 4,
		},
		{
			configs: []configuration.Configuration{
				{
					Entries: []string{"", "", ""},
				},
				{
					Entries: []string{"four"},
				},
			},
			resolvedAmount: 4,
		},
	}

	for i, test := range tests {
		result := splitConfigurations(test.configs)
		assert.Equal(t, test.resolvedAmount, len(result), "index %d failed", i)
	}
}
