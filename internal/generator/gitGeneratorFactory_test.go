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
