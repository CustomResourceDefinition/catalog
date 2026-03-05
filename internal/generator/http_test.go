package generator

import (
	"regexp"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/stretchr/testify/assert"
)

func TestHttpGeneratorVersions(t *testing.T) {
	expectedVersions := []string{"1.0.0", "1.3.0", "2.0.0"}

	config := configuration.Configuration{
		Kind: configuration.Http,
		Downloads: []configuration.ConfigurationDownload{
			{Version: expectedVersions[0]},
			{Version: expectedVersions[1]},
			{Version: expectedVersions[2]},
		},
	}

	generator := NewHttpGenerator(config, nil)
	defer generator.Close()

	versions, err := generator.Versions(regexp.MustCompile(".*"))
	assert.Nil(t, err)
	assert.Equal(t, expectedVersions, versions)
}

func TestHttpGeneratorUnknownVersion(t *testing.T) {
	config := configuration.Configuration{
		Kind: configuration.Http,
		Downloads: []configuration.ConfigurationDownload{
			{Version: "1.0.0"},
			{Version: "2.0.0"},
			{Version: "3.0.0"},
		},
	}

	generator := NewHttpGenerator(config, nil)

	metadata, err := generator.MetaData("4.5.6")
	assert.Nil(t, metadata)
	assert.NotNil(t, err)
}

func TestHttpGeneratorSchemas(t *testing.T) {
	version := "1.0.0"
	urlPath := "anything.yaml"
	server, finish := setupServer(urlPath, "testdata/test-crd.yaml")
	defer finish()

	config := configuration.Configuration{
		Name:      "http",
		Kind:      configuration.Http,
		ApiGroups: []string{"crd.example.com"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: version,
				Paths:   []string{urlPath},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewHttpGenerator(config, reader)
	defer generator.Close()

	crds, err := generator.Crds(version)
	assert.Nil(t, err)

	schemas := make([]crd.CrdSchema, 0)
	for _, c := range crds {
		schema, err := c.Schema()
		if !assert.Nil(t, err) {
			return
		}
		schemas = append(schemas, schema...)
	}

	assert.Nil(t, err)
	assert.Equal(t, 1, len(schemas))
	assert.Equal(t, "crd.example.com", schemas[0].Group)
	assert.Equal(t, "test", schemas[0].Kind)
	assert.Equal(t, "v1", schemas[0].Version)
}

func TestHttpGeneratorMetadata(t *testing.T) {
	version := "1.0.0"
	urlPath := "anything.yaml"
	server, finish := setupServer(urlPath, "testdata/test-crd.yaml")
	defer finish()

	config := configuration.Configuration{
		Name:      "http",
		Kind:      configuration.Http,
		ApiGroups: []string{"crd.example.com"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: version,
				Paths:   []string{urlPath},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewHttpGenerator(config, reader)
	defer generator.Close()

	metadata, err := generator.MetaData(version)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "crd.example.com", metadata[0].Group)
	assert.Equal(t, "test", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestHttpGeneratorPartialSchemas(t *testing.T) {
	version := "1.0.0"
	urlPath := "version1.yaml"
	server, finish := setupServer(urlPath, "testdata/test-crd.yaml")
	defer finish()

	config := configuration.Configuration{
		Name:      "http",
		Kind:      configuration.Http,
		ApiGroups: []string{"crd.example.com"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: version,
				Paths: []string{
					"not-present.yaml",
					urlPath,
				},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewHttpGenerator(config, reader)
	defer generator.Close()

	crds, err := generator.Crds(version)
	assert.Nil(t, err)

	schemas := make([]crd.CrdSchema, 0)
	for _, c := range crds {
		schema, err := c.Schema()
		if !assert.Nil(t, err) {
			return
		}
		schemas = append(schemas, schema...)
	}

	if !assert.Equal(t, 1, len(schemas)) {
		return
	}
	assert.Equal(t, "crd.example.com", schemas[0].Group)
	assert.Equal(t, "test", schemas[0].Kind)
	assert.Equal(t, "v1", schemas[0].Version)
}

func TestHttpGeneratorNoSchemas(t *testing.T) {
	version := "1.0.0"
	server, finish := setupServer("file.yaml", "testdata/test-crd.yaml")
	defer finish()

	config := configuration.Configuration{
		Name:      "http",
		Kind:      configuration.Http,
		ApiGroups: []string{"crd.example.com"},
		Downloads: []configuration.ConfigurationDownload{
			{
				BaseUri: server.URL,
				Version: version,
				Paths: []string{
					"not-present.yaml",
				},
			},
		},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewHttpGenerator(config, reader)
	defer generator.Close()

	crds, err := generator.Crds(version)
	assert.Nil(t, err)

	schemas := make([]crd.CrdSchema, 0)
	for _, c := range crds {
		schema, err := c.Schema()
		if !assert.Nil(t, err) {
			return
		}
		schemas = append(schemas, schema...)
	}

	assert.Equal(t, 0, len(schemas))
}

func TestHttpGeneratorVersionsFiltering(t *testing.T) {
	tests := []struct {
		versions         []string
		expectedVersions []string
		pattern          string
	}{
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
			versions:         []string{"2.0.0", "v1.3v", "v1.0.0"},
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
		{
			versions:         []string{"main", "v1.0", "master"},
			expectedVersions: []string{"main", "master"},
			pattern:          `^(main|master)$`,
		},
		{
			versions:         []string{"main", "v1.0.0", "2.0.0", "master"},
			expectedVersions: []string{"2.0.0", "main", "master"},
			pattern:          `^([0-9]+\.[0-9]+\.[0-9]+)|(main|master)$`,
		},
	}

	for i, test := range tests {
		downloads := make([]configuration.ConfigurationDownload, 0)
		for _, v := range test.versions {
			downloads = append(downloads, configuration.ConfigurationDownload{Version: v})
		}
		config := configuration.Configuration{
			Kind:      configuration.Http,
			Downloads: downloads,
		}

		generator := NewHttpGenerator(config, nil)
		defer generator.Close()

		var filter *regexp.Regexp
		if test.pattern != "" {
			filter = regexp.MustCompile(test.pattern)
		} else {
			filter = regexp.MustCompile(".*")
		}

		versions, err := generator.Versions(filter)
		assert.Nil(t, err, "index %d failed", i)
		assert.Equal(t, test.expectedVersions, versions, "index %d failed", i)
	}
}
