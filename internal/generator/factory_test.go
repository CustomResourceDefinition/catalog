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
	"github.com/CustomResourceDefinition/catalog/internal/timing"
	"github.com/stretchr/testify/assert"
)

func TestBuildWithVersionPatternFiltering(t *testing.T) {
	config := configuration.Configuration{
		Kind:           configuration.Http,
		Name:           "test",
		ApiGroups:      []string{"chart.uri"},
		VersionPattern: `^v([0-9]+\.[0-9]+\.[0-9]+)$`,
		Downloads: []configuration.ConfigurationDownload{
			{
				Version: "v1.0.0",
			},
			{
				Version: "v2.0.0",
			},
			{
				Version: "3.0.0",
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	tmpDir := t.TempDir()

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), nil)
	assert.Nil(t, err)

	version, result, _, err := builder.registryStatus()
	assert.Nil(t, err)
	assert.False(t, result)
	assert.Equal(t, "v2.0.0", version)
}

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

	version, result, _, err := builder.registryStatus()
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
	reg.Set("test", "http", "1.0.0", nil, nil)

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), reg)
	assert.Nil(t, err)

	version, result, _, err := builder.registryStatus()
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
	reg.Set("test", "http", "1.0.0", nil, nil)

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), reg)
	assert.Nil(t, err)

	version, result, _, err := builder.registryStatus()
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
	reg.Set("test", "git", "1.0.0", nil, nil)

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger(), reg)
	assert.Nil(t, err)

	version, result, _, err := builder.registryStatus()
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
	reg.Set("test", "http", "1.0.0", nil, nil)

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

func TestBuildRecordsStats(t *testing.T) {
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

	assert.NotNil(t, builder.Stats())

	err = builder.Build()
	assert.Nil(t, err)

	stats := builder.Stats()
	assert.Greater(t, stats.TotalOperations(), 0)
	assert.Greater(t, int(stats.TotalTime()), 0)

	httpOps := stats.GetCategoryStats(timing.CategoryHTTP)
	assert.NotNil(t, httpOps)
	assert.NotEmpty(t, httpOps)

	genOps := stats.GetCategoryStats(timing.CategoryGeneration)
	assert.NotNil(t, genOps)
	assert.NotEmpty(t, genOps)
}

func TestBuildTracksCrds(t *testing.T) {
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

	entry, ok := reg.Get("test")
	assert.True(t, ok)
	assert.Len(t, entry.Refs, 1)
	assert.Equal(t, "crd.example.com", entry.Refs[0].Group)
	assert.Equal(t, "test", entry.Refs[0].Kind)
	assert.Equal(t, "v1", entry.Refs[0].Version)
	assert.Empty(t, entry.PastRefs)
}

func TestComputeRemoved(t *testing.T) {
	currentCrds := []registry.CrdRef{
		{Group: "a.com", Kind: "Test", Version: "v1"},
		{Group: "a.com", Kind: "Test", Version: "v2"},
	}

	pastRefs := []registry.CrdRef{
		{Group: "a.com", Kind: "Old", Version: "v1"},
		{Group: "a.com", Kind: "Test", Version: "v1"},
	}

	removed := computeRemoved(pastRefs, pastRefs, currentCrds)

	assert.Len(t, removed, 1)
	found := make(map[string]bool)
	for _, r := range removed {
		found[r.Kind] = true
	}
	assert.False(t, found["Test"])
	assert.True(t, found["Old"])
}

func TestComputeRemovedEmpty(t *testing.T) {
	currentCrds := []registry.CrdRef{
		{Group: "a.com", Kind: "Test", Version: "v1"},
	}

	empty := make([]registry.CrdRef, 0)
	removed := computeRemoved(empty, empty, currentCrds)
	assert.Len(t, removed, 0)
}

func TestComputeRemovedAllPresent(t *testing.T) {
	reg := &registry.SourceRegistry{Sources: make(map[string]registry.SourceEntry)}
	reg.Set("test", "git", "1.0.0",
		[]registry.CrdRef{
			{Group: "a.com", Kind: "Test", Version: "v1"},
		},
		[]registry.CrdRef{
			{Group: "b.com", Kind: "Old", Version: "v1"},
		},
	)

	currentCrds := []registry.CrdRef{
		{Group: "a.com", Kind: "Test", Version: "v1"},
		{Group: "b.com", Kind: "Old", Version: "v1"},
	}

	empty := make([]registry.CrdRef, 0)
	removed := computeRemoved(empty, empty, currentCrds)
	assert.Len(t, removed, 0)
}

func TestExtractCrdRefs(t *testing.T) {
	// Test with empty slice
	builder := Builder{}
	refs := builder.extractCrdRefs(nil)
	assert.Empty(t, refs)

	refs = builder.extractCrdRefs([]crd.Crd{})
	assert.Empty(t, refs)
}

func TestOperationCategory(t *testing.T) {
	tests := []struct {
		kind     configuration.Kind
		expected timing.Category
	}{
		{kind: configuration.Git, expected: timing.CategoryGit},
		{kind: configuration.Http, expected: timing.CategoryHTTP},
		{kind: configuration.Helm, expected: timing.CategoryHelm},
		{kind: configuration.HelmOci, expected: timing.CategoryOCI},
		{kind: configuration.Kind("unknown"), expected: timing.CategoryMisc},
	}

	for _, tt := range tests {
		config := configuration.Configuration{
			Kind:      tt.kind,
			Name:      "test",
			Downloads: []configuration.ConfigurationDownload{{Version: "1.0.0"}},
		}

		builder := &Builder{config: config}
		result := builder.operationCategory()
		assert.Equal(t, tt.expected, result, "kind: %s", tt.kind)
	}
}

func TestStatsInitialized(t *testing.T) {
	config := configuration.Configuration{
		Kind:      configuration.Http,
		Name:      "test",
		Downloads: []configuration.ConfigurationDownload{{Version: "1.0.0"}},
	}

	builder, err := NewBuilder(config, nil, "-", "-", "-", nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, builder.stats)
}
