package generator

import (
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

	versions, err := generator.Versions()
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

	schemas, err := generator.Schemas(version)
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

	schemas, err := generator.Schemas(version)
	assert.Nil(t, err)
	if !assert.Equal(t, 1, len(schemas)) {
		return
	}
	assert.Equal(t, "crd.example.com", schemas[0].Group)
	assert.Equal(t, "test", schemas[0].Kind)
	assert.Equal(t, "v1", schemas[0].Version)
}
