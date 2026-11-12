package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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

	generator := NewGitGenerator(config, nil, regexp.MustCompile(".*"))
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

	generator := NewGitGenerator(config, nil, regexp.MustCompile(".*"))
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

	generator := NewGitGenerator(config, reader, regexp.MustCompile(".*"))
	defer generator.Close()

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "crd.example.com", metadata[0].Group)
	assert.Equal(t, "test", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestGitGeneratorExcludePaths(t *testing.T) {
	tests := []struct {
		name         string
		searchPaths  []string
		excludePaths []string
		want         []string
	}{
		{
			name:        "no exclusions",
			searchPaths: []string{"deploy/crds"},
			want:        []string{"current", "deprecated", "nested", "sibling", "keep"},
		},
		{
			name:         "excludes directory and descendants",
			searchPaths:  []string{"deploy/crds"},
			excludePaths: []string{"deploy/crds/deprecated"},
			want:         []string{"current", "sibling", "keep"},
		},
		{
			name:         "excludes individual file and retains later sibling",
			searchPaths:  []string{"deploy/crds"},
			excludePaths: []string{"deploy/crds/current.yaml"},
			want:         []string{"deprecated", "nested", "sibling", "keep"},
		},
		{
			name:         "does not exclude similarly named directory",
			searchPaths:  []string{"deploy/crds"},
			excludePaths: []string{"deploy/crds/deprecated"},
			want:         []string{"current", "sibling", "keep"},
		},
		{
			name:         "search path nested in exclusion",
			searchPaths:  []string{"deploy/crds/deprecated"},
			excludePaths: []string{"deploy/crds/deprecated"},
			want:         []string{},
		},
		{
			name:         "exclusions apply to multiple search paths",
			searchPaths:  []string{"deploy/crds", "other"},
			excludePaths: []string{"deploy/crds/deprecated"},
			want:         []string{"current", "sibling", "keep", "other"},
		},
		{
			name:         "missing exclusion",
			searchPaths:  []string{"deploy/crds"},
			excludePaths: []string{"deploy/crds/missing"},
			want:         []string{"current", "deprecated", "nested", "sibling", "keep"},
		},
		{
			name:         "path matching is case sensitive",
			searchPaths:  []string{"deploy/crds"},
			excludePaths: []string{"Deploy/crds/deprecated"},
			want:         []string{"current", "deprecated", "nested", "sibling", "keep"},
		},
		{
			name:         "trailing slash",
			searchPaths:  []string{"deploy/crds"},
			excludePaths: []string{"deploy/crds/deprecated/"},
			want:         []string{"current", "sibling", "keep"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture, err := os.ReadFile("testdata/test-crd.yaml")
			if err != nil {
				t.Fatal(err)
			}

			files := []struct {
				path   string
				marker string
			}{
				{path: "deploy/crds/current.yaml", marker: "current"},
				{path: "deploy/crds/deprecated/old.yaml", marker: "deprecated"},
				{path: "deploy/crds/deprecated/nested/older.yaml", marker: "nested"},
				{path: "deploy/crds/deprecated-other/other.yaml", marker: "sibling"},
				{path: "deploy/crds/keep.yaml", marker: "keep"},
				{path: "other/crd.yaml", marker: "other"},
			}
			bundlePaths := make([]gitPath, 0, len(files))
			for _, file := range files {
				content := strings.Replace(string(fixture), "This is a description", file.marker, 1)
				fixturePath := filepath.Join(t.TempDir(), file.marker+".yaml")
				if err := os.WriteFile(fixturePath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				bundlePaths = append(bundlePaths, gitPath{path: file.path, file: fixturePath})
			}

			repository, err := setupGit(t, []gitBundle{{tag: "1.0.0", paths: bundlePaths}})
			if err != nil {
				t.Fatal(err)
			}

			config := configuration.Configuration{
				Kind:         configuration.Git,
				Repository:   fmt.Sprintf("file://%s", *repository),
				SearchPaths:  test.searchPaths,
				ExcludePaths: test.excludePaths,
			}
			reader, err := crd.NewCrdReader(setupLogger())
			if err != nil {
				t.Fatal(err)
			}
			generator := NewGitGenerator(config, reader, regexp.MustCompile(".*"))
			defer generator.Close()

			crds, err := generator.Crds("1.0.0")
			if err != nil {
				t.Fatal(err)
			}

			got := make([]string, 0, len(crds))
			for _, item := range crds {
				for _, file := range files {
					if strings.Contains(string(item.Bytes), file.marker) {
						got = append(got, file.marker)
						break
					}
				}
			}
			sort.Strings(got)
			want := append([]string{}, test.want...)
			sort.Strings(want)
			assert.Len(t, crds, len(want))
			assert.Equal(t, want, got)
		})
	}
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

	generator := NewGitGenerator(config, reader, regexp.MustCompile(".*"))
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

	generator := NewGitGenerator(config, reader, regexp.MustCompile(".*"))
	defer generator.Close()

	metadata, err := generator.MetaData("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metadata))
	assert.Equal(t, "source.example.com", metadata[0].Group)
	assert.Equal(t, "foo", metadata[0].Kind)
	assert.Equal(t, "v1", metadata[0].Version)
}

func TestGitGeneratorCloneFailure(t *testing.T) {
	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "file:///nonexistent/path/repo.git",
	}

	generator := NewGitGenerator(config, nil, regexp.MustCompile(".*"))
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

	generator := NewGitGenerator(config, nil, regexp.MustCompile(".*"))
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

	generator := NewGitGenerator(config, nil, regexp.MustCompile(".*"))
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

func TestGitGeneratorVersionsCustomSort(t *testing.T) {
	seedVersions := []string{
		"2.10.0", "1.0.0", "2.2.0", "2.1.0", "1.01.01",
	}
	bundles := make([]gitBundle, 0)
	for _, v := range seedVersions {
		bundles = append(bundles, gitBundle{tag: v, paths: []gitPath{}})
	}

	p, err := setupGit(t, bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: fmt.Sprintf("file://%s", *p),
	}

	filter := regexp.MustCompile(`^([0-9]+\.[0-9]+\.[0-9]+)$`)
	generator := NewGitGenerator(config, nil, filter)
	defer generator.Close()

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, []string{"2.10.0", "2.2.0", "2.1.0", "1.01.01", "1.0.0"}, versions)
}

func TestGitGeneratorVersionsSortByTimestamp(t *testing.T) {
	baseTime := int64(1704000000)
	newerTime := baseTime + 86400

	bundles := []gitBundle{
		{tag: "v1.0.0", timestamp: &baseTime, paths: []gitPath{{path: "crds/v1.yaml", file: "testdata/test-crd.yaml"}}},
		{tag: "v2.0.0", timestamp: &baseTime, paths: []gitPath{{path: "crds/v2.yaml", file: "testdata/test-crd.yaml"}}},
		{tag: "v3.0.0", timestamp: &newerTime, paths: []gitPath{{path: "crds/v3.yaml", file: "testdata/test-crd.yaml"}}},
	}

	p, err := setupGit(t, bundles)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: fmt.Sprintf("file://%s", *p),
	}

	filter := regexp.MustCompile(`^v([0-9]+\.[0-9]+\.[0-9]+)$`)
	generator := NewGitGenerator(config, nil, filter)
	defer generator.Close()

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, []string{"v3.0.0", "v2.0.0", "v1.0.0"}, versions)
}
