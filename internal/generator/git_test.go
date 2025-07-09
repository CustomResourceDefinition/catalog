package generator

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/stretchr/testify/assert"
)

func TestGitGeneratorVersions(t *testing.T) {
	expectedVersion := "1.0.0"

	bundles := []gitBundle{
		{
			tag: expectedVersion,
			paths: []gitPath{
				{
					path: "regular/crd.yaml",
					file: "testdata/test-crd.yaml",
				},
			},
		},
	}

	p, err := setupGit(bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: fmt.Sprintf("file://%s", *p),
	}

	generator := NewGitGenerator(config, nil)

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, []string{expectedVersion}, versions)
}

func TestGitGeneratorUnknownVersion(t *testing.T) {
	expectedVersion := "1.0.0"

	bundles := []gitBundle{
		{
			tag: expectedVersion,
			paths: []gitPath{
				{
					path: "regular/crd.yaml",
					file: "testdata/test-crd.yaml",
				},
			},
		},
	}

	p, err := setupGit(bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: fmt.Sprintf("file://%s", *p),
	}

	generator := NewGitGenerator(config, nil)

	metadata, err := generator.MetaData("4.5.6")
	assert.Nil(t, metadata)
	assert.NotNil(t, err)
}

func TestGitGeneratorMetadataForRegularFile(t *testing.T) {
	expectedVersion := "1.0.0"

	bundles := []gitBundle{
		{
			tag: expectedVersion,
			paths: []gitPath{
				{
					path: "regular/crd.yaml",
					file: "testdata/test-crd.yaml",
				},
			},
		},
	}

	p, err := setupGit(bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:        configuration.Git,
		Repository:  fmt.Sprintf("file://%s", *p),
		SearchPaths: []string{"extra/", "regular/"},
	}

	reader, err := crd.NewCrdReader(bytes.NewBuffer([]byte{}))
	assert.Nil(t, err)

	generator := NewGitGenerator(config, reader)

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "crd.example.com", metadata[0].Group)
	assert.Equal(t, "test", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestGitGeneratorMetadataForKustomizeFile(t *testing.T) {
	expectedVersion := "1.0.0"

	bundles := []gitBundle{
		{
			tag: expectedVersion,
			paths: []gitPath{
				{
					path: "regular/crd.yaml",
					file: "testdata/test-crd.yaml",
				},
			},
		},
	}

	p, err := setupGit(bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:           configuration.Git,
		Repository:     fmt.Sprintf("file://%s", *p),
		KustomizePaths: []string{"kustomize/"},
	}

	reader, err := crd.NewCrdReader(bytes.NewBuffer([]byte{}))
	assert.Nil(t, err)

	generator := NewGitGenerator(config, reader)

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "onepassword.com", metadata[0].Group)
	assert.Equal(t, "onepassworditem", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestGitGeneratorMetadataForSourceFiles(t *testing.T) {
	expectedVersion := "1.0.0"

	bundles := []gitBundle{
		{
			tag: expectedVersion,
			paths: []gitPath{
				{
					path: "regular/crd.yaml",
					file: "testdata/test-crd.yaml",
				},
			},
		},
	}

	p, err := setupGit(bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: fmt.Sprintf("file://%s", *p),
		GenPaths:   []string{"./api/..."},
	}

	reader, err := crd.NewCrdReader(bytes.NewBuffer([]byte{}))
	assert.Nil(t, err)

	generator := NewGitGenerator(config, reader)

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "onepassword.com", metadata[0].Group)
	assert.Equal(t, "onepassworditem", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}
