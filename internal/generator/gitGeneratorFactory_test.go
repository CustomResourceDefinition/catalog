package generator

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
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

	generator, err := NewGitGeneratorFactory(config, nil, regexp.MustCompile(".*"), nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	version, err := generator.LatestVersion()
	assert.Nil(t, err)
	assert.Equal(t, version, "main")
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

	generator, err := NewGitGeneratorFactory(config, nil, regexp.MustCompile(".*"), logger).Build()
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

	generator, err := NewGitGeneratorFactory(config, nil, regexp.MustCompile(".*"), nil).Build()
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

	generator, err := NewGitGeneratorFactory(config, nil, regexp.MustCompile(".*"), nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	info := generator.(*PreparedGitGenerator).versions
	versions := make([]string, 0)
	for _, i := range info {
		versions = append(versions, i.name)
	}

	assert.Equal(t, len(versions), 4)
	assert.Contains(t, versions, "v1.0.0")
	assert.Contains(t, versions, "v0.9.0")
	assert.Contains(t, versions, "v0.8.0")
	assert.Contains(t, versions, "main")
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

	generator, err := NewGitGeneratorFactory(config, nil, regexp.MustCompile(".*"), nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	info := generator.(*PreparedGitGenerator).versions
	versions := make([]string, 0)
	for _, i := range info {
		versions = append(versions, i.name)
	}

	assert.Equal(t, len(versions), 4)
	assert.Contains(t, versions, "v1.0.0")
	assert.Contains(t, versions, "v0.9.0")
	assert.Contains(t, versions, "nested-tag")
	assert.Contains(t, versions, "main")
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

	generator, err := NewGitGeneratorFactory(config, nil, regexp.MustCompile(".*"), nil).Build()
	assert.Nil(t, err)
	defer generator.Close()
	assert.IsType(t, &PreparedGitGenerator{}, generator)

	info := generator.(*PreparedGitGenerator).versions
	versions := make([]string, 0)
	for _, i := range info {
		versions = append(versions, i.name)
	}

	assert.Equal(t, len(versions), 5)
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

			generator, err := NewGitGeneratorFactory(config, nil, regexp.MustCompile(".*"), nil).Build()
			assert.Nil(t, err)
			defer generator.Close()
			assert.IsType(t, &PreparedGitGenerator{}, generator)

			versions := generator.(*PreparedGitGenerator).versions
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
