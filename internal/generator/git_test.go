package generator

import (
	"fmt"
	"os"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/stretchr/testify/assert"
)

func TestGitGeneratorVersionsCombinesTagsAndBranches(t *testing.T) {
	expectedVersion := "1.0.0"
	expectedDevelop := "develop"
	expectedFeature := "feature"

	bundles := []gitBundle{
		{
			tag:    expectedVersion,
			branch: expectedDevelop,
			paths: []gitPath{
				{
					path: "regular/crd.yaml",
					file: "testdata/test-crd.yaml",
				},
			},
		},
		{
			branch: expectedFeature,
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
	defer generator.Close()

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Contains(t, versions, expectedVersion)
	assert.Contains(t, versions, expectedDevelop)
	assert.Contains(t, versions, expectedFeature)
	assert.Contains(t, versions, "main")
	assert.True(t, len(versions) == 4, "unexpected values found in versions")
}

func TestGitGeneratorUnknownVersion(t *testing.T) {
	bundles := []gitBundle{
		{
			tag: "1.0.0",
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
	defer generator.Close()

	metadata, err := generator.MetaData("4.5.6")
	assert.Nil(t, metadata)
	assert.NotNil(t, err)
}

func TestGitGeneratorMetadataForRegularFile(t *testing.T) {
	bundles := []gitBundle{
		{
			tag: "1.0.0",
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
	defer generator.Close()

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "crd.example.com", metadata[0].Group)
	assert.Equal(t, "test", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestGitGeneratorMetadataForKustomizeFile(t *testing.T) {
	bundles := []gitBundle{
		{
			tag: "1.0.0",
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
	defer generator.Close()

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "crd.example.com", metadata[0].Group)
	assert.Equal(t, "test", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestGitGeneratorMetadataForSourceFiles(t *testing.T) {
	bundles := []gitBundle{
		{
			tag: "1.0.0",
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
		Kind:       configuration.Git,
		Repository: fmt.Sprintf("file://%s", *p),
		GenPaths:   []string{"./api/..."},
	}

	reader, err := crd.NewCrdReader(setupLogger())
	assert.Nil(t, err)

	generator := NewGitGenerator(config, reader)
	defer generator.Close()

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "source.example.com", metadata[0].Group)
	assert.Equal(t, "foo", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestGitGeneratorVersionSortKeyForBranch(t *testing.T) {
	bundles := []gitBundle{
		{
			tag:    "1.0.0",
			branch: "develop",
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
	defer generator.Close()

	key, err := generator.VersionSortKey("main")
	assert.Nil(t, err)
	assert.Greater(t, key, int64(0))

	key, err = generator.VersionSortKey("develop")
	assert.Nil(t, err)
	assert.Greater(t, key, int64(0))
}

func TestGitGeneratorCloneFailure(t *testing.T) {
	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "file:///nonexistent/path/repo.git",
	}

	generator := NewGitGenerator(config, nil)
	defer generator.Close()

	_, err := generator.Versions()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unable to clone")
}

func TestGitGeneratorCheckoutFailure(t *testing.T) {
	bundles := []gitBundle{
		{
			tag: "1.0.0",
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
	defer generator.Close()

	_, err = generator.Crds("nonexistent-branch-tag")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unable to checkout")
}

func TestGitGeneratorClose(t *testing.T) {
	bundles := []gitBundle{
		{
			tag: "1.0.0",
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
	defer generator.Close()

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.NotEmpty(t, versions)

	tmpDir := versions[0]

	err = generator.Close()
	assert.Nil(t, err)

	_, err = os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(err), "temp dir should be removed after Close()")
}
