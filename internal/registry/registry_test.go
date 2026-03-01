package registry

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadFileNotFound(t *testing.T) {
	reg, err := Load("/nonexistent/path/registry.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, reg)
	assert.NotNil(t, reg.Sources)
	assert.Equal(t, 0, len(reg.Sources))
}

func TestLoadInvalidYaml(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := path.Join(tmpDir, "registry.yaml")

	err := os.WriteFile(filePath, []byte("invalid: yaml: content:"), 0644)
	assert.Nil(t, err)

	_, err = Load(filePath)
	assert.NotNil(t, err)
}

func TestLoadYamlWithNoSources(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := path.Join(tmpDir, "registry.yaml")

	yamlContent := `sources: null`
	err := os.WriteFile(filePath, []byte(yamlContent), 0644)
	assert.Nil(t, err)

	reg, err := Load(filePath)
	assert.Nil(t, err)
	assert.NotNil(t, reg)
	assert.NotNil(t, reg.Sources)
	assert.Equal(t, 0, len(reg.Sources))
}

func TestLoadValidYaml(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := path.Join(tmpDir, "registry.yaml")

	yamlContent := `sources:
  test-source:
    kind: git
    version: "1.0.0"
    lastUpdated: "2000-01-01T00:00:00Z"
`
	err := os.WriteFile(filePath, []byte(yamlContent), 0644)
	assert.Nil(t, err)

	reg, err := Load(filePath)
	assert.Nil(t, err)
	assert.NotNil(t, reg)
	assert.Equal(t, 1, len(reg.Sources))

	entry, ok := reg.Get("test-source")
	assert.True(t, ok)
	assert.Equal(t, "git", entry.Kind)
	assert.Equal(t, "1.0.0", entry.Version)
	assert.Equal(t, "2000-01-01T00:00:00Z", entry.LastUpdated)
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := path.Join(tmpDir, "registry.yaml")

	original := &SourceRegistry{
		Sources: map[string]SourceEntry{
			"source1": {Kind: "git", Version: "1.0.0", LastUpdated: "2000-01-01T00:00:00Z"},
			"source2": {Kind: "helm", Version: "2.0.0", LastUpdated: "2000-01-02T00:00:00Z"},
		},
	}

	err := original.Save(filePath)
	assert.Nil(t, err)

	loaded, err := Load(filePath)
	assert.Nil(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, 2, len(loaded.Sources))

	entry1, ok := loaded.Get("source1")
	assert.True(t, ok)
	assert.Equal(t, "git", entry1.Kind)
	assert.Equal(t, "1.0.0", entry1.Version)

	entry2, ok := loaded.Get("source2")
	assert.True(t, ok)
	assert.Equal(t, "helm", entry2.Kind)
	assert.Equal(t, "2.0.0", entry2.Version)
}

func TestGet(t *testing.T) {
	reg := &SourceRegistry{
		Sources: map[string]SourceEntry{
			"existing": {Kind: "git", Version: "1.0.0"},
		},
	}

	entry, ok := reg.Get("existing")
	assert.True(t, ok)
	assert.Equal(t, "git", entry.Kind)
	assert.Equal(t, "1.0.0", entry.Version)

	_, ok = reg.Get("nonexistent")
	assert.False(t, ok)
}

func TestSet(t *testing.T) {
	reg := &SourceRegistry{
		Sources: make(map[string]SourceEntry),
	}

	reg.Set("new-source", "http", "3.0.0")

	entry, ok := reg.Get("new-source")
	assert.True(t, ok)
	assert.Equal(t, "http", entry.Kind)
	assert.Equal(t, "3.0.0", entry.Version)
	assert.NotEmpty(t, entry.LastUpdated)
}

func TestSetUpdate(t *testing.T) {
	reg := &SourceRegistry{
		Sources: map[string]SourceEntry{
			"source1": {Kind: "git", Version: "1.0.0", LastUpdated: "2000-01-01T00:00:00Z"},
		},
	}

	reg.Set("source1", "git", "2.0.0")

	entry, ok := reg.Get("source1")
	assert.True(t, ok)
	assert.Equal(t, "2.0.0", entry.Version)
	assert.NotEqual(t, "2000-01-01T00:00:00Z", entry.LastUpdated)
}

func TestLoadEmptyYaml(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := path.Join(tmpDir, "registry.yaml")

	err := os.WriteFile(filePath, []byte("sources: {}"), 0644)
	assert.Nil(t, err)

	reg, err := Load(filePath)
	assert.Nil(t, err)
	assert.NotNil(t, reg.Sources)
	assert.Equal(t, 0, len(reg.Sources))
}

func TestSourceRegistrySetMultiple(t *testing.T) {
	reg := &SourceRegistry{
		Sources: make(map[string]SourceEntry),
	}

	reg.Set("source1", "git", "1.0.0")
	reg.Set("source2", "helm", "2.0.0")
	reg.Set("source3", "oci", "3.0.0")

	assert.Equal(t, 3, len(reg.Sources))

	v1, _ := reg.Get("source1")
	v2, _ := reg.Get("source2")
	v3, _ := reg.Get("source3")

	assert.Equal(t, "1.0.0", v1.Version)
	assert.Equal(t, "git", v1.Kind)
	assert.Equal(t, "2.0.0", v2.Version)
	assert.Equal(t, "helm", v2.Kind)
	assert.Equal(t, "3.0.0", v3.Version)
	assert.Equal(t, "oci", v3.Kind)
}

func TestSourceRegistryLastUpdatedFormat(t *testing.T) {
	reg := &SourceRegistry{
		Sources: make(map[string]SourceEntry),
	}

	before := time.Now().UTC().Format(time.RFC3339)
	reg.Set("source", "git", "1.0.0")
	after := time.Now().UTC().Format(time.RFC3339)

	entry, _ := reg.Get("source")

	_, err := time.Parse(time.RFC3339, entry.LastUpdated)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, entry.LastUpdated, before)
	assert.LessOrEqual(t, entry.LastUpdated, after)
}
