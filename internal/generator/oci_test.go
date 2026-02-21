package generator

import (
	"fmt"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/stretchr/testify/assert"
)

func TestOciGeneratorVersions(t *testing.T) {
	t.Setenv(HELM_OCI_PLAIN_HTTP, "true")

	server, finish := setupOciServer(t, ociChart{
		repoName: "helm",
		name:     "connect",
		tag:      "2.0.0",
		path:     "testdata/connect-2.0.0.tgz",
	})
	defer finish()

	config := configuration.Configuration{
		Kind:       configuration.HelmOci,
		Repository: fmt.Sprintf("%s%s", server.URL, "/helm/connect"),
	}

	generator := NewOciGenerator(config, nil)
	defer generator.Close()

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, []string{"2.0.0"}, versions)
}

func TestOciGeneratorUnknownVersion(t *testing.T) {
	t.Setenv(HELM_OCI_PLAIN_HTTP, "true")

	server, finish := setupOciServer(t, ociChart{
		repoName: "helm",
		name:     "connect",
		tag:      "2.0.0",
		path:     "testdata/connect-2.0.0.tgz",
	})
	defer finish()

	config := configuration.Configuration{
		Name:       "oci",
		Kind:       configuration.HelmOci,
		Repository: fmt.Sprintf("%s%s", server.URL, "/helm/connect"),
	}

	generator := NewOciGenerator(config, nil)
	defer generator.Close()

	metadata, err := generator.MetaData("4.5.6")
	assert.Nil(t, metadata)
	assert.NotNil(t, err)
}

func TestOciGeneratorMetadata(t *testing.T) {
	t.Setenv(HELM_OCI_PLAIN_HTTP, "true")

	server, finish := setupOciServer(t, ociChart{
		repoName: "helm",
		name:     "connect",
		tag:      "2.0.0",
		path:     "testdata/connect-2.0.0.tgz",
	})
	defer finish()

	config := configuration.Configuration{
		Name:       "oci",
		Kind:       configuration.HelmOci,
		Repository: fmt.Sprintf("%s%s", server.URL, "/helm/connect"),
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewOciGenerator(config, reader)
	defer generator.Close()

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "onepassword.com", metadata[0].Group)
	assert.Equal(t, "onepassworditem", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestOciGeneratorHasInertSortingKeys(t *testing.T) {
	config := configuration.Configuration{
		Name:       "oci",
		Kind:       configuration.HelmOci,
		Repository: fmt.Sprintf("%s%s", "http://localhost", "/helm/connect"),
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewOciGenerator(config, reader)
	defer generator.Close()

	versions := []string{"0.0.0", "1.0.0", "3.2.1", "999.999.999"}

	for _, version := range versions {
		key, err := generator.VersionSortKey(version)
		assert.Nil(t, err)
		assert.Equal(t, key, int64(0))
	}
}
