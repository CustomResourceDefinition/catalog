package generator

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/stretchr/testify/assert"
)

func TestGitGeneratorFactoryIsGitHubRepo(t *testing.T) {
	tests := []struct {
		name        string
		repoURL     string
		expectValid bool
		expectOwner string
		expectRepo  string
	}{
		{
			name:        "standard URL",
			repoURL:     "https://github.com/owner/repo",
			expectValid: true,
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:        "with trailing slash",
			repoURL:     "https://github.com/owner/repo/",
			expectValid: true,
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:        "with .git suffix",
			repoURL:     "https://github.com/owner/repo.git",
			expectValid: true,
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:        "non-GitHub URL",
			repoURL:     "https://gitlab.com/owner/repo",
			expectValid: false,
			expectOwner: "",
			expectRepo:  "",
		},
		{
			name:        "file URL",
			repoURL:     "file:///path/to/repo",
			expectValid: false,
			expectOwner: "",
			expectRepo:  "",
		},
		{
			name:        "with-protocol",
			repoURL:     "github.com/owner/repo",
			expectValid: false,
			expectOwner: "",
			expectRepo:  "",
		},
		{
			name:        "gibberish",
			repoURL:     "af/3/tl33lg/.3l/hih//alj;lls",
			expectValid: false,
			expectOwner: "",
			expectRepo:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isGitHub, owner, repo := isGitHubRepo(tt.repoURL)
			assert.Equal(t, tt.expectValid, isGitHub)
			assert.Equal(t, tt.expectOwner, owner)
			assert.Equal(t, tt.expectRepo, repo)
		})
	}
}

func TestGitGeneratorFactoryBuildGitHubSuccessWithVersions(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	cleanup := setupGitHubServer(t, []gitHubResponse{
		{
			prefix: "refs/tags/",
			tags: []githubRef{
				{name: "v1.0.0", committedDate: "2000-01-15T10:00:00Z"},
				{name: "v0.9.0", committedDate: "2000-01-10T10:00:00Z"},
			},
		},
		{
			prefix: "refs/heads/",
			branches: []githubRef{
				{name: "main", committedDate: "2000-01-20T10:00:00Z"},
				{name: "develop", committedDate: "2000-01-18T10:00:00Z"},
			},
		},
	})
	defer cleanup()

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := NewGitGeneratorFactory(config, nil, nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Contains(t, versions, "v1.0.0")
	assert.Contains(t, versions, "v0.9.0")
	assert.Contains(t, versions, "main")
	assert.Contains(t, versions, "develop")
}

func TestGitGeneratorFactoryBuildGitHubSuccessWithSortKey(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	cleanup := setupGitHubServer(t, []gitHubResponse{
		{
			prefix: "refs/tags/",
			tags: []githubRef{
				{name: "v1.0.0", committedDate: "2000-01-15T10:00:00Z"},
			},
		},
		{
			prefix: "refs/heads/",
			branches: []githubRef{
				{name: "main", committedDate: "2000-01-20T10:00:00Z"},
			},
		},
	})
	defer cleanup()

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := NewGitGeneratorFactory(config, nil, nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	key1, err := generator.VersionSortKey("v1.0.0")
	assert.Nil(t, err)
	assert.Greater(t, key1, int64(0))

	key2, err := generator.VersionSortKey("main")
	assert.Nil(t, err)
	assert.Greater(t, key2, int64(0))

	assert.Greater(t, key2, key1, "main (2000-01-20) should have later timestamp than v1.0.0 (2000-01-15)")

	_, err = generator.VersionSortKey("nonexistent")
	assert.NotNil(t, err)
}

func TestPreparedGitGeneratorVersionsSortedByDate(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	cleanup := setupGitHubServer(t, []gitHubResponse{
		{
			prefix: "refs/tags/",
			tags: []githubRef{
				{name: "v0.9.0", committedDate: "2000-01-10T10:00:00Z"},
				{name: "v1.0.0", committedDate: "2000-01-15T10:00:00Z"},
			},
		},
		{
			prefix: "refs/heads/",
			branches: []githubRef{
				{name: "develop", committedDate: "2000-01-18T10:00:00Z"},
				{name: "main", committedDate: "2000-01-20T10:00:00Z"},
			},
		},
	})
	defer cleanup()

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := NewGitGeneratorFactory(config, nil, nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, []string{"v0.9.0", "v1.0.0", "develop", "main"}, versions)
}

var logger = bytes.NewBuffer([]byte{})

func TestGitGeneratorFactoryBuildGitHubFailure(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	originalEndpoint := graphQLEndpoint
	graphQLEndpoint = server.URL
	defer func() {
		graphQLEndpoint = originalEndpoint
	}()

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := NewGitGeneratorFactory(config, nil, logger).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &GitGenerator{}, generator)
}

func TestGitGeneratorFactoryBuildGitHubSuccess(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	cleanup := setupGitHubServer(t, []gitHubResponse{
		{
			prefix: "refs/tags/",
			tags: []githubRef{
				{name: "v1.0.0", committedDate: "2000-01-15T10:00:00Z"},
			},
		},
		{
			prefix: "refs/heads/",
			branches: []githubRef{
				{name: "main", committedDate: "2000-01-20T10:00:00Z"},
			},
		},
	})
	defer cleanup()

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := NewGitGeneratorFactory(config, nil, nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)
}

func TestResolveGeneratorWithoutGitHubToken(t *testing.T) {
	os.Unsetenv("GITHUB_TOKEN")

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := resolveGenerator(config, nil, nil)
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &GitGenerator{}, generator)
}

func TestResolveGeneratorNonGitHubWithToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://gitlab.com/owner/repo",
	}

	generator, err := resolveGenerator(config, nil, nil)
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &GitGenerator{}, generator)
}

func TestGitGeneratorFactoryAnnotatedTags(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	cleanup := setupGitHubServer(t, []gitHubResponse{
		{
			prefix: "refs/tags/",
			tags: []githubRef{
				{name: "v1.0.0", committedDate: "2000-01-15T10:00:00Z"},
				{name: "v0.9.0", targetType: "tag", taggerDate: "2000-01-12T10:00:00Z"},
				{name: "v0.8.0", targetType: "tag", committedDate: "2000-01-10T10:00:00Z"},
			},
		},
		{
			prefix: "refs/heads/",
			branches: []githubRef{
				{name: "main", committedDate: "2000-01-20T10:00:00Z"},
			},
		},
	})
	defer cleanup()

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := NewGitGeneratorFactory(config, nil, nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Contains(t, versions, "v1.0.0")
	assert.Contains(t, versions, "v0.9.0")
	assert.Contains(t, versions, "v0.8.0")
	assert.Contains(t, versions, "main")

	key1, _ := generator.VersionSortKey("v1.0.0")
	key2, _ := generator.VersionSortKey("v0.9.0")
	key3, _ := generator.VersionSortKey("v0.8.0")
	keyMain, _ := generator.VersionSortKey("main")

	assert.Greater(t, keyMain, key1, "main > v1.0.0")
	assert.Greater(t, key1, key2, "v1.0.0 > v0.9.0 (via target.tagger.date)")
	assert.Greater(t, key2, key3, "v0.9.0 > v0.8.0 (via target.target.committedDate)")
}

func TestGitGeneratorFactoryNestedTags(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	cleanup := setupGitHubServer(t, []gitHubResponse{
		{
			prefix: "refs/tags/",
			tags: []githubRef{
				{name: "v1.0.0", committedDate: "2000-01-15T10:00:00Z"},
				{name: "v0.9.0", targetType: "tag", taggerDate: "2000-01-12T10:00:00Z"},
				{name: "nested-tag", targetType: "tag", committedDate: "", taggerDate: "2000-01-08T10:00:00Z", nested: &githubRef{
					targetType:    "tag",
					taggerDate:    "2000-01-05T10:00:00Z",
					committedDate: "2000-01-03T10:00:00Z",
				}},
			},
		},
		{
			prefix: "refs/heads/",
			branches: []githubRef{
				{name: "main", committedDate: "2000-01-20T10:00:00Z"},
			},
		},
	})
	defer cleanup()

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := NewGitGeneratorFactory(config, nil, nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Contains(t, versions, "v1.0.0")
	assert.Contains(t, versions, "v0.9.0")
	assert.Contains(t, versions, "nested-tag")
	assert.Contains(t, versions, "main")

	key1, _ := generator.VersionSortKey("v1.0.0")
	key2, _ := generator.VersionSortKey("v0.9.0")
	keyNested, _ := generator.VersionSortKey("nested-tag")
	keyMain, _ := generator.VersionSortKey("main")

	assert.Greater(t, keyMain, key1, "main > v1.0.0")
	assert.Greater(t, key1, key2, "v1.0.0 > v0.9.0")
	assert.Greater(t, key2, keyNested, "v0.9.0 > nested-tag (via nested target.target.committedDate)")
}

func TestGitGeneratorFactoryPagination(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	cleanup := setupGitHubServer(t, []gitHubResponse{
		{
			prefix:      "refs/tags/",
			refsPerPage: 2,
			tags: []githubRef{
				{name: "v1.0.0", committedDate: "2000-01-15T10:00:00Z"},
				{name: "v0.9.0", committedDate: "2000-01-14T10:00:00Z"},
				{name: "v0.8.0", committedDate: "2000-01-13T10:00:00Z"},
				{name: "v0.7.0", committedDate: "2000-01-12T10:00:00Z"},
			},
		},
		{
			prefix: "refs/heads/",
			branches: []githubRef{
				{name: "main", committedDate: "2000-01-20T10:00:00Z"},
			},
		},
	})
	defer cleanup()

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := NewGitGeneratorFactory(config, nil, nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, 5, len(versions), "should have all 4 tags + 1 branch")
	assert.Contains(t, versions, "v1.0.0")
	assert.Contains(t, versions, "v0.9.0")
	assert.Contains(t, versions, "v0.8.0")
	assert.Contains(t, versions, "v0.7.0")
	assert.Contains(t, versions, "main")
}

func TestGitGeneratorFactoryEmptyResponses(t *testing.T) {
	tests := []struct {
		name             string
		tags             []githubRef
		branches         []githubRef
		expectedVersions int
	}{
		{
			name:             "no tags",
			tags:             []githubRef{},
			branches:         []githubRef{{name: "main", committedDate: "2000-01-20T10:00:00Z"}},
			expectedVersions: 1,
		},
		{
			name:             "no branches",
			tags:             []githubRef{{name: "v1.0.0", committedDate: "2000-01-15T10:00:00Z"}},
			branches:         []githubRef{},
			expectedVersions: 1,
		},
		{
			name:             "empty repo",
			tags:             []githubRef{},
			branches:         []githubRef{},
			expectedVersions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_TOKEN", "test-token")

			cleanup := setupGitHubServer(t, []gitHubResponse{
				{prefix: "refs/tags/", tags: tt.tags},
				{prefix: "refs/heads/", branches: tt.branches},
			})
			defer cleanup()

			config := configuration.Configuration{
				Kind:       configuration.Git,
				Repository: "https://github.com/owner/repo",
			}

			generator, err := NewGitGeneratorFactory(config, nil, nil).Build()
			assert.Nil(t, err)
			defer generator.Close()
			assert.IsType(t, &PreparedGitGenerator{}, generator)

			versions, err := generator.Versions()
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedVersions, len(versions))
		})
	}
}

func TestParseGitHubDate(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{
			name:        "valid RFC3339Nano date",
			dateStr:     "2000-01-15T10:00:00.123456789Z",
			expectError: false,
		},
		{
			name:        "valid RFC3339 date without nano",
			dateStr:     "2000-01-15T10:00:00Z",
			expectError: false,
		},
		{
			name:        "valid date with timezone offset",
			dateStr:     "2000-01-15T10:00:00+05:30",
			expectError: false,
		},
		{
			name:        "empty date string",
			dateStr:     "",
			expectError: true,
		},
		{
			name:        "invalid date format",
			dateStr:     "2000-01-15",
			expectError: true,
		},
		{
			name:        "random string",
			dateStr:     "not-a-date",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseGitHubDate(tt.dateStr)
			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
