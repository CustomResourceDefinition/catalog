package generator

import (
	"fmt"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/stretchr/testify/assert"
)

// FIXME: more tests
func TestEngineWithNoConfiguration(t *testing.T) {
	// config := configuration.Configuration{}

	// schemas, err := NewDriver().Read(config, "")

	// assert.Nil(t, schemas)
	// assert.NotNil(t, err)
}

func TestEngineWithConfiguration(t *testing.T) {
	// config := configuration.Configuration{
	// 	Kind: configuration.Http,
	// }

	// schemas, err := NewDriver().Read(config, "")

	// assert.Nil(t, err)
	// assert.NotNil(t, schemas)
}

func TestVersionSorting(t *testing.T) {
	var result int

	result = compareVersion("1.0.0", "1.0.0")
	assert.Equal(t, 0, result)

	result = compareVersion("2.0.0", "1.0.0")
	assert.Equal(t, 1, result)

	result = compareVersion("2.1.0", "2.2.0")
	assert.Equal(t, -1, result)

	result = compareVersion("2.10.0", "2.2.0")
	assert.Equal(t, 1, result)

	result = compareVersion("2.10.0", "2.1.0")
	assert.Equal(t, 1, result)

	result = compareVersion("2.10.0", "2.0.0")
	assert.Equal(t, 1, result)
}

func TestBuilderVersionSorting(t *testing.T) {
	seedVersions := []string{
		"2.10.0", "1.0.0", "2.2.0", "2.1.0",
	}
	bundles := make([]gitBundle, 0)
	for _, v := range seedVersions {
		bundles = append(bundles, gitBundle{tag: v, paths: []gitPath{}})
	}
	expectedVersions := []string{seedVersions[0], seedVersions[2], seedVersions[3], seedVersions[1], "head"}

	p, err := setupGit(bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:        configuration.Git,
		Repository:  fmt.Sprintf("file://%s", *p),
		IncludeHead: true,
	}

	b, err := NewBuilder(config, nil, "-", nil)
	assert.Nil(t, err)

	versions, err := b.versions()
	assert.Nil(t, err)
	assert.NotNil(t, versions)
	assert.Equal(t, expectedVersions, versions)
}

type testScenario struct {
	versions         []string
	expectedVersions []string
	prefix, suffix   string
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
			prefix:           "v",
		},
		{
			versions:         []string{"2.0.0", "v1.3.0", "v1.0.0"},
			expectedVersions: []string{"2.0.0", "v1.3.0", "v1.0.0"},
			prefix:           "v?",
		},
		{
			versions:         []string{"2.0.0", "v1.3.0v", "v1.0.0"},
			expectedVersions: []string{"2.0.0", "v1.0.0"},
			prefix:           "v?",
			suffix:           "",
		},
		{
			versions:         []string{"2.0.0-2", "1.3.0-1892", "1.0.0-01"},
			expectedVersions: []string{"2.0.0-2", "1.3.0-1892", "1.0.0-01"},
			suffix:           `\-\d*`,
		},
		{
			versions:         []string{"v1.33.2+k0s.0"},
			expectedVersions: []string{"v1.33.2+k0s.0"},
			prefix:           "v",
			suffix:           `\+k0s\.`,
		},
	}

	for i, test := range tests {
		downloads := make([]configuration.ConfigurationDownload, 0)
		for _, v := range test.versions {
			downloads = append(downloads, configuration.ConfigurationDownload{Version: v})
		}
		config := configuration.Configuration{
			Kind:          configuration.Http,
			Downloads:     downloads,
			VersionPrefix: test.prefix,
			VersionSuffix: test.suffix,
		}

		b, err := NewBuilder(config, nil, "-", nil)
		assert.Nil(t, err)

		versions, err := b.versions()
		assert.Nil(t, err, "index %d failed", i)
		assert.NotNil(t, versions, "index %d failed", i)
		assert.Equal(t, test.expectedVersions, versions, "index %d failed", i)
	}
}
