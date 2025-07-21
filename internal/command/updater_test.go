package command

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	template := `
- apiGroups:
    - chart.uri
  crds:
    - baseUri: {{ server }}
      paths:
        - chart-1.0.0.yaml
      version: 1.0.0
  kind: http
  name: http
`
	b, err := os.ReadFile("testdata/updater/multiple.yaml")
	assert.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	config := path.Join(tmpDir, "config.yaml")
	os.WriteFile(config, []byte(strings.ReplaceAll(template, "{{ server }}", server.URL)), 0664)

	updater := NewUpdater(config, tmpDir, bytes.NewBuffer([]byte{}), nil)

	err = updater.Run()
	assert.Nil(t, err)

	err = os.Remove(config)
	assert.Nil(t, err)

	assertDirectories(t, "testdata/updater/output", tmpDir)
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

func TestReadConfiguration(t *testing.T) {
	_, err := readConfiguration("testdata/does-not.exist")
	assert.NotNil(t, err)

	_, err = readConfiguration("testdata/updater/multiple.yaml")
	assert.NotNil(t, err)
}

func TestInitialize(t *testing.T) {
	updater := Updater{}
	err := updater.initialize()
	assert.Nil(t, err)
	assert.Equal(t, os.Stderr, updater.Logger)
}

func TestValidateConfiguration(t *testing.T) {
	updater := Updater{
		Output:        "testdata",
		Configuration: "testdata",
	}
	err := updater.validate()
	assert.NotNil(t, err)
}

func TestValidateOutput(t *testing.T) {
	updater := Updater{
		Output:        "testdata/does-not-exist",
		Configuration: "testdata/updater/multiple.yaml",
	}
	err := updater.validate()
	assert.NotNil(t, err)
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

// Test is skipped during unit testing and meant for step-debugging a local check
func TestCheckLocal(t *testing.T) {
	output := "../../build/ephemeral/schema"
	config := "../../test/configuration.yaml"
	updater := NewUpdater(config, output, nil, nil)

	err := updater.Run()
	assert.Nil(t, err)
	assert.True(t, false, "this test should always be skipped and/or ignored")
}
