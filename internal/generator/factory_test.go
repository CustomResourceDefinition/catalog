package generator

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/registry"
	"github.com/stretchr/testify/assert"
)

func TestResolveGenerator(t *testing.T) {
	invalidConfigs := []configuration.Configuration{
		{
			Kind: "",
		},
	}
	validConfigs := []configuration.Configuration{
		{
			Kind: configuration.Git,
		},
		{
			Kind: configuration.Http,
		},
		{
			Kind:    configuration.Helm,
			Entries: []string{""},
		},
		{
			Kind: configuration.HelmOci,
		},
	}

	for _, c := range validConfigs {
		g, err := resolveGenerator(c, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, g)
	}

	for _, c := range invalidConfigs {
		g, err := resolveGenerator(c, nil, nil)
		assert.NotNil(t, err)
		assert.Nil(t, g)
	}
}

func TestBuild(t *testing.T) {
	b, err := os.ReadFile("testdata/test-crd.yaml")
	assert.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer server.Close()

	config := configuration.Configuration{
		Kind:      configuration.Http,
		Name:      "test",
		ApiGroups: []string{"chart.uri"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: "1.0.0",
				Paths:   []string{"any"},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	tmpDir := t.TempDir()

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), nil)
	assert.Nil(t, err)

	err = builder.Build()
	assert.Nil(t, err)

	fs, err := os.Stat(path.Join(tmpDir, "crd.example.com", "test_v1.json"))
	assert.Nil(t, err)
	assert.True(t, !fs.IsDir())
}

func TestBuildWithLatestVersion(t *testing.T) {
	b, err := os.ReadFile("testdata/test-crd.yaml")
	assert.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer server.Close()

	config := configuration.Configuration{
		Kind:      configuration.Http,
		Name:      "test",
		ApiGroups: []string{"chart.uri"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: "1.0.0",
				Paths:   []string{"any"},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	tmpDir := t.TempDir()
	directory := path.Join(tmpDir, "crd.example.com")
	filename := path.Join(directory, "test_v1.json")
	os.MkdirAll(directory, 0775)
	os.WriteFile(filename, []byte{}, 0664)

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), nil)
	assert.Nil(t, err)

	err = builder.Build()
	assert.Nil(t, err)

	fs, err := os.Stat(filename)
	assert.Nil(t, err)
	assert.True(t, !fs.IsDir())
}

func TestRegistryStatusNoRegistry(t *testing.T) {
	config := configuration.Configuration{
		Kind:      configuration.Http,
		Name:      "test",
		Downloads: []configuration.ConfigurationDownload{{Version: "1.0.0"}},
	}

	builder, err := NewBuilder(config, nil, "-", "-", "-", nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, builder)

	version, result, err := builder.registryStatus()
	assert.Nil(t, err)
	assert.False(t, result)
	assert.Equal(t, "1.0.0", version)
}

func TestRegistryStatusSameVersion(t *testing.T) {
	b, err := os.ReadFile("testdata/test-crd.yaml")
	assert.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer server.Close()

	config := configuration.Configuration{
		Kind:      configuration.Http,
		Name:      "test",
		ApiGroups: []string{"chart.uri"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: "1.0.0",
				Paths:   []string{"any"},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	tmpDir := t.TempDir()
	reg := &registry.SourceRegistry{Sources: make(map[string]registry.SourceEntry)}
	reg.Set("test", "http", "1.0.0")

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), reg)
	assert.Nil(t, err)

	version, result, err := builder.registryStatus()
	assert.Nil(t, err)
	assert.True(t, result)
	assert.Equal(t, "1.0.0", version)
}

func TestRegistryStatusDifferentVersion(t *testing.T) {
	b, err := os.ReadFile("testdata/test-crd.yaml")
	assert.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer server.Close()

	config := configuration.Configuration{
		Kind:      configuration.Http,
		Name:      "test",
		ApiGroups: []string{"chart.uri"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: "2.0.0",
				Paths:   []string{"any"},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	tmpDir := t.TempDir()
	reg := &registry.SourceRegistry{Sources: make(map[string]registry.SourceEntry)}
	reg.Set("test", "http", "1.0.0")

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), reg)
	assert.Nil(t, err)

	version, result, err := builder.registryStatus()
	assert.Nil(t, err)
	assert.False(t, result)
	assert.Equal(t, "2.0.0", version)
}

func TestRegistryStatusDifferentKind(t *testing.T) {
	b, err := os.ReadFile("testdata/test-crd.yaml")
	assert.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer server.Close()

	config := configuration.Configuration{
		Kind:      configuration.Http,
		Name:      "test",
		ApiGroups: []string{"chart.uri"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: "1.0.0",
				Paths:   []string{"any"},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	tmpDir := t.TempDir()
	reg := &registry.SourceRegistry{Sources: make(map[string]registry.SourceEntry)}
	reg.Set("test", "git", "1.0.0")

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), reg)
	assert.Nil(t, err)

	version, result, err := builder.registryStatus()
	assert.Nil(t, err)
	assert.False(t, result)
	assert.Equal(t, "1.0.0", version)
}

func TestBuildSkipsWhenUnchanged(t *testing.T) {
	b, err := os.ReadFile("testdata/test-crd.yaml")
	assert.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer server.Close()

	config := configuration.Configuration{
		Kind:      configuration.Http,
		Name:      "test",
		ApiGroups: []string{"chart.uri"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: "1.0.0",
				Paths:   []string{"any"},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	tmpDir := t.TempDir()
	reg := &registry.SourceRegistry{Sources: make(map[string]registry.SourceEntry)}
	reg.Set("test", "http", "1.0.0")

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), reg)
	assert.Nil(t, err)

	err = builder.Build()
	assert.Nil(t, err)

	_, err = os.Stat(path.Join(tmpDir, "crd.example.com", "test_v1.json"))
	assert.True(t, os.IsNotExist(err))
}

func TestBuildRunsWhenChanged(t *testing.T) {
	b, err := os.ReadFile("testdata/test-crd.yaml")
	assert.Nil(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer server.Close()

	config := configuration.Configuration{
		Kind:      configuration.Http,
		Name:      "test",
		ApiGroups: []string{"chart.uri"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: "1.0.0",
				Paths:   []string{"any"},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	tmpDir := t.TempDir()
	reg := &registry.SourceRegistry{Sources: make(map[string]registry.SourceEntry)}

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), reg)
	assert.Nil(t, err)

	err = builder.Build()
	assert.Nil(t, err)

	fs, err := os.Stat(path.Join(tmpDir, "crd.example.com", "test_v1.json"))
	assert.Nil(t, err)
	assert.False(t, fs.IsDir())

	entry, ok := reg.Get("test")
	assert.True(t, ok)
	assert.Equal(t, "http", entry.Kind)
	assert.Equal(t, "1.0.0", entry.Version)
}
