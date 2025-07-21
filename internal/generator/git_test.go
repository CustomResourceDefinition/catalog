package generator

import (
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

	p, err := setupGit(t, bundles)
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

	p, err := setupGit(t, bundles)
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

	p, err := setupGit(t, bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:        configuration.Git,
		Repository:  fmt.Sprintf("file://%s", *p),
		SearchPaths: []string{"extra/", "regular/"},
	}

	reader, err := crd.NewCrdReader(setupLogger())
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
					path: "kustomize/kustomization.yaml",
					file: "../kustomize/testdata/kustomization.yaml",
				},
				{
					path: "kustomize/password.txt",
					file: "../kustomize/testdata/password.txt",
				},
				{
					path: "kustomize/crd.yaml",
					file: "../kustomize/testdata/crd.yaml",
				},
			},
		},
	}

	p, err := setupGit(t, bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:           configuration.Git,
		Repository:     fmt.Sprintf("file://%s", *p),
		KustomizePaths: []string{"kustomize/"},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewGitGenerator(config, reader)

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "crd.example.com", metadata[0].Group)
	assert.Equal(t, "test", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestGitGeneratorMetadataForSourceFiles(t *testing.T) {
	expectedVersion := "1.0.0"

	bundles := []gitBundle{
		{
			tag: expectedVersion,
			paths: []gitPath{
				{
					path: "api/go.mod",
					file: "../genall/testdata/source/go.mod",
				},
				{
					path: "api/go.sum",
					file: "../genall/testdata/source/go.sum",
				},
				{
					path: "api/source.go",
					file: "../genall/testdata/source/source.go",
				},
			},
		},
	}

	p, err := setupGit(t, bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:        configuration.Git,
		Repository:  fmt.Sprintf("file://%s", *p),
		GenPaths:    []string{"./api/..."},
		IncludeHead: true,
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewGitGenerator(config, reader)

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "source.example.com", metadata[0].Group)
	assert.Equal(t, "foo", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}
