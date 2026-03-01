package generator

import (
	"fmt"
	"os"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/stretchr/testify/assert"
)

func TestPreparedGitGeneratorVersions(t *testing.T) {
	bundles := []gitBundle{
		{
			tag:    "1.0.0",
			branch: "main",
		},
	}

	p, err := setupGit(t, bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: fmt.Sprintf("file://%s", *p),
	}

	gitGen := NewGitGenerator(config, nil).(*GitGenerator)
	defer gitGen.Close()

	versions := []versionInfo{
		{name: "v1.0.0", timestamp: 1705317600},
		{name: "main", timestamp: 1705749600},
		{name: "develop", timestamp: 1705569600},
	}

	generator := NewPreparedGitGenerator(gitGen, versions)

	result, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, []string{"v1.0.0", "main", "develop"}, result)
}

func TestPreparedGitGeneratorVersionSortKey(t *testing.T) {
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

	gitGen := NewGitGenerator(config, nil).(*GitGenerator)
	defer gitGen.Close()

	versions := []versionInfo{
		{name: "v1.0.0", timestamp: 1705317600},
		{name: "main", timestamp: 1705749600},
	}

	generator := NewPreparedGitGenerator(gitGen, versions)

	key, err := generator.VersionSortKey("v1.0.0")
	assert.Nil(t, err)
	assert.Equal(t, int64(1705317600), key)

	key, err = generator.VersionSortKey("main")
	assert.Nil(t, err)
	assert.Equal(t, int64(1705749600), key)

	_, err = generator.VersionSortKey("nonexistent")
	assert.NotNil(t, err)
}

func TestPreparedGitGeneratorVersionSortKeyEmpty(t *testing.T) {
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

	gitGen := NewGitGenerator(config, nil).(*GitGenerator)
	defer gitGen.Close()

	versions := []versionInfo{}

	generator := NewPreparedGitGenerator(gitGen, versions)

	_, err = generator.VersionSortKey("v1.0.0")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "not found")
}

func TestPreparedGitGeneratorClose(t *testing.T) {
	bundles := []gitBundle{
		{
			tag: "1.0.0",
			paths: []gitPath{
				{
					path: "crd.yaml",
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

	gitGen := NewGitGenerator(config, nil).(*GitGenerator)

	versions := []versionInfo{
		{name: "v1.0.0", timestamp: 1705317600},
	}

	generator := NewPreparedGitGenerator(gitGen, versions)

	versionsResult, err := generator.Versions()
	assert.Nil(t, err)
	assert.NotEmpty(t, versionsResult)

	tmpDir := versionsResult[0]

	err = generator.Close()
	assert.Nil(t, err)

	_, err = os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(err), "temp dir should be removed after Close()")
}
