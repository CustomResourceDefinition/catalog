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
	server, config, tmpDir := setup(t)
	defer server.Close()

	updater := NewUpdater(config, tmpDir, tmpDir, "", bytes.NewBuffer([]byte{}), nil)

	err := updater.Run()
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

func TestValidSourceKeys(t *testing.T) {
	assert.Empty(t, validSourceKeys(nil))
	assert.Empty(t, validSourceKeys([]configuration.Configuration{}))

	configs := []configuration.Configuration{{Name: "foo"}}
	keys := validSourceKeys(configs)
	assert.True(t, keys["foo"])
	assert.Len(t, keys, 1)

	configs = []configuration.Configuration{{Name: "bar", Entries: []string{"a", "b"}}}
	keys = validSourceKeys(configs)
	assert.True(t, keys["bar.a"])
	assert.True(t, keys["bar.b"])
	assert.Len(t, keys, 2)

	configs = []configuration.Configuration{
		{Name: "single"},
		{Name: "multi", Entries: []string{"x", "y", "z"}},
	}
	keys = validSourceKeys(configs)
	assert.True(t, keys["single"])
	assert.True(t, keys["multi.x"])
	assert.True(t, keys["multi.y"])
	assert.True(t, keys["multi.z"])
	assert.Len(t, keys, 4)
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
		Schema:        "testdata",
		Configuration: "testdata",
	}
	err := updater.validate()
	assert.NotNil(t, err)
}

func TestValidateOutput(t *testing.T) {
	updater := Updater{
		Schema:        "testdata/does-not-exist",
		Configuration: "testdata/updater/multiple.yaml",
	}
	err := updater.validate()
	assert.NotNil(t, err)
}

func setup(t *testing.T) (*httptest.Server, string, string) {
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

	tmpDir := t.TempDir()

	config := path.Join(tmpDir, "config.yaml")
	os.WriteFile(config, []byte(strings.ReplaceAll(template, "{{ server }}", server.URL)), 0664)

	return server, config, tmpDir
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

func TestRunWithRegistryLoadError(t *testing.T) {
	unused := "testdata/updater/multiple.yaml"
	tmpDir := t.TempDir()

	registryPath := path.Join(tmpDir, "registry.yaml")
	os.WriteFile(registryPath, []byte("invalid: yaml: content:"), 0644)

	updater := NewUpdater(unused, tmpDir, tmpDir, registryPath, bytes.NewBuffer([]byte{}), nil)

	err := updater.Run()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to load registry")
}

func TestRunWithRegistrySavesUpdates(t *testing.T) {
	server, config, tmpDir := setup(t)
	defer server.Close()

	registryPath := path.Join(tmpDir, "registry.yaml")
	initialContent := "sources: {}\n"
	os.WriteFile(registryPath, []byte(initialContent), 0664)

	updater := NewUpdater(config, tmpDir, tmpDir, registryPath, bytes.NewBuffer([]byte{}), nil)

	err := updater.Run()
	assert.Nil(t, err)

	content, err := os.ReadFile(registryPath)
	assert.Nil(t, err)
	assert.NotEqual(t, initialContent, string(content))
	assert.Contains(t, string(content), "http")
	assert.Contains(t, string(content), "1.0.0")
}

func TestRunWithRegistryRemovesStaleEntries(t *testing.T) {
	server, config, tmpDir := setup(t)
	defer server.Close()

	registryPath := path.Join(tmpDir, "registry.yaml")
	initialContent := `sources:
  http:
    kind: http
    version: 1.0.0
    lastUpdated: "2026-01-01T00:00:00Z"
  stale1:
    kind: helm
    version: v1.0.0
    lastUpdated: "2026-01-01T00:00:00Z"
  stale2:
    kind: git
    version: v1.0.0
    lastUpdated: "2026-01-01T00:00:00Z"
`
	os.WriteFile(registryPath, []byte(initialContent), 0664)

	updater := NewUpdater(config, tmpDir, tmpDir, registryPath, bytes.NewBuffer([]byte{}), nil)

	err := updater.Run()
	assert.Nil(t, err)

	content, err := os.ReadFile(registryPath)
	assert.Nil(t, err)

	assert.Contains(t, string(content), "http")
	assert.Contains(t, string(content), "1.0.0")
	assert.NotContains(t, string(content), "stale1")
	assert.NotContains(t, string(content), "stale2")
}

// Test is skipped during unit testing and meant for step-debugging a local check
func TestCheckLocal(t *testing.T) {
	output := "../../build/ephemeral/schema"
	config := "../../build/configuration.yaml"
	updater := NewUpdater(config, output, output, "", nil, nil)

	err := updater.Run()
	assert.Nil(t, err)
	assert.True(t, false, "this test should always be skipped and/or ignored")
}
