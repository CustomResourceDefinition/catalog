package generator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/stretchr/testify/assert"
)

func TestBuilderVersionSorting(t *testing.T) {
	seedVersions := []string{
		"2.10.0", "1.0.0", "2.2.0", "2.1.0", "1.01.01",
	}
	bundles := make([]gitBundle, 0)
	for _, v := range seedVersions {
		bundles = append(bundles, gitBundle{tag: v, paths: []gitPath{}})
	}
	expectedVersions := []string{seedVersions[0], seedVersions[2], seedVersions[3], seedVersions[4], seedVersions[1]}

	p, err := setupGit(t, bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:           configuration.Git,
		Repository:     fmt.Sprintf("file://%s", *p),
		VersionPattern: `^([0-9]+\.[0-9]+\.[0-9]+)$`,
	}

	b, err := NewBuilder(config, nil, "-", "-", "-", nil)
	assert.Nil(t, err)

	versions, err := b.versions()
	assert.Nil(t, err)
	assert.NotNil(t, versions)
	assert.Equal(t, expectedVersions, versions)
}

type testScenario struct {
	versions         []string
	expectedVersions []string
	pattern          string
}

func TestBuilderVersionFiltering(t *testing.T) {
	tests := []testScenario{
		{
			versions:         []string{"2.0.0", "1.3.0", "1.0.0"},
			expectedVersions: []string{"2.0.0", "1.3.0", "1.0.0"},
		},
		{
			versions:         []string{"v2.0.0", "v1.3.0", "v1.0.0"},
			expectedVersions: []string{},
		},
		{
			versions:         []string{"v2.0.0", "v1.3.0", "v1.0.0"},
			expectedVersions: []string{"v2.0.0", "v1.3.0", "v1.0.0"},
			pattern:          `^v([0-9]+\.[0-9]+\.[0-9]+)$`,
		},
		{
			versions:         []string{"2.0.0", "v1.3.0", "v1.0.0"},
			expectedVersions: []string{"2.0.0", "v1.3.0", "v1.0.0"},
			pattern:          `^v?([0-9]+\.[0-9]+\.[0-9]+)$`,
		},
		{
			versions:         []string{"2.0.0", "v1.3.0v", "v1.0.0"},
			expectedVersions: []string{"2.0.0", "v1.0.0"},
			pattern:          `^v?([0-9]+\.[0-9]+\.[0-9]+)$`,
		},
		{
			versions:         []string{"2.0.0-2", "1.3.0-1892", "1.0.0-01"},
			expectedVersions: []string{"2.0.0-2", "1.3.0-1892", "1.0.0-01"},
			pattern:          `^([0-9]+\.[0-9]+\.[0-9]+-\d+)$`,
		},
		{
			versions:         []string{"v1.33.2+k0s.0"},
			expectedVersions: []string{"v1.33.2+k0s.0"},
			pattern:          `^v([0-9]+\.[0-9]+\.[0-9]+\+k0s\.0)$`,
		},
	}

	for i, test := range tests {
		downloads := make([]configuration.ConfigurationDownload, 0)
		for _, v := range test.versions {
			downloads = append(downloads, configuration.ConfigurationDownload{Version: v})
		}
		config := configuration.Configuration{
			Kind:           configuration.Http,
			Downloads:      downloads,
			VersionPattern: test.pattern,
		}

		b, err := NewBuilder(config, nil, "-", "-", "-", nil)
		assert.Nil(t, err)

		versions, err := b.versions()
		assert.Nil(t, err, "index %d failed", i)
		assert.NotNil(t, versions, "index %d failed", i)
		assert.Equal(t, test.expectedVersions, versions, "index %d failed", i)
	}
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
		g, err := resolveGenerator(c, nil)
		assert.Nil(t, err)
		assert.NotNil(t, g)
	}

	for _, c := range invalidConfigs {
		g, err := resolveGenerator(c, nil)
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

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger())
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

	builder, err := NewBuilder(config, reader, tmpDir, tmpDir, tmpDir, setupLogger())
	assert.Nil(t, err)

	err = builder.Build()
	assert.Nil(t, err)

	fs, err := os.Stat(filename)
	assert.Nil(t, err)
	assert.True(t, !fs.IsDir())
}
