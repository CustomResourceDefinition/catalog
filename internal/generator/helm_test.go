package generator

import (
	"os"
	"strings"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/stretchr/testify/assert"
)

func TestHelmGeneratorVersions(t *testing.T) {
	expectedVersions := []string{"2.0.0"}

	server, finish := setupServer("index.yaml", "testdata/helm-index.yaml")
	defer finish()

	config := configuration.Configuration{
		Kind:       configuration.Helm,
		Repository: server.URL,
	}

	generator := NewHelmGenerator("connect", config, nil)
	defer generator.Close()

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, expectedVersions, versions)
}

func TestHelmGeneratorUnknownTarget(t *testing.T) {
	server, finish := setupServer("index.yaml", "testdata/helm-index.yaml")
	defer finish()

	config := configuration.Configuration{
		Kind:       configuration.Helm,
		Repository: server.URL,
	}

	generator := NewHelmGenerator("unknown", config, nil)
	defer generator.Close()

	versions, err := generator.Versions()
	assert.Nil(t, versions)
	assert.NotNil(t, err)
}

func TestHelmGeneratorUnknownVersion(t *testing.T) {
	server, finish := setupServer("index.yaml", "testdata/helm-index.yaml")
	defer finish()

	config := configuration.Configuration{
		Name:       "helm",
		Kind:       configuration.Helm,
		Repository: server.URL,
	}

	generator := NewHelmGenerator("connect", config, nil)
	defer generator.Close()

	metadata, err := generator.MetaData("4.5.6")
	assert.Nil(t, metadata)
	assert.NotNil(t, err)
}

func TestHelmGeneratorMetadata(t *testing.T) {
	b, _ := os.ReadFile("testdata/helm-index.yaml")
	s := string(b)

	server, finish := setupServer("1Password/connect-helm-charts/releases/download/connect-2.0.0/connect-2.0.0.tgz", "testdata/connect-2.0.0.tgz")
	defer finish()
	updated := strings.ReplaceAll(s, "https://github.com", server.URL)
	server.addRequest(serverRequest{urlPath: "index.yaml", status: 200, response: strings.NewReader(updated)})

	config := configuration.Configuration{
		Name:       "helm",
		Kind:       configuration.Helm,
		Repository: server.URL,
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewHelmGenerator("connect", config, reader)
	defer generator.Close()

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "onepassword.com", metadata[0].Group)
	assert.Equal(t, "onepassworditem", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestHelmGeneratorHasInertSortingKeys(t *testing.T) {
	config := configuration.Configuration{
		Kind:       configuration.Helm,
		Repository: "http:localhost",
	}

	generator := NewHelmGenerator("connect", config, nil)
	defer generator.Close()

	versions := []string{"0.0.0", "1.0.0", "3.2.1", "999.999.999"}

	for _, version := range versions {
		key, err := generator.VersionSortKey(version)
		assert.Nil(t, err)
		assert.Equal(t, key, int64(0))
	}
}
