package generator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/stretchr/testify/assert"
)

func TestIsGitHubRepo(t *testing.T) {
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
			name:        "file path",
			repoURL:     "file:///path/to/repo",
			expectValid: false,
			expectOwner: "",
			expectRepo:  "",
		},
		{
			name:        "non-https URL",
			repoURL:     "github.com/owner/repo",
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

func TestGitHubGeneratorVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		_ = body["query"].(string)
		variables := body["variables"].(map[string]any)

		var response map[string]any

		prefix := variables["prefix"].(string)
		if prefix == "refs/tags/" {
			response = map[string]any{
				"data": map[string]any{
					"repository": map[string]any{
						"refs": map[string]any{
							"nodes": []any{
								map[string]any{
									"name": "v1.0.0",
									"target": map[string]any{
										"committedDate": "2024-01-15T10:00:00Z",
									},
								},
								map[string]any{
									"name": "v0.9.0",
									"target": map[string]any{
										"committedDate": "2024-01-10T10:00:00Z",
									},
								},
							},
							"pageInfo": map[string]any{
								"hasNextPage": false,
								"endCursor":   "",
							},
						},
					},
				},
			}
		} else {
			response = map[string]any{
				"data": map[string]any{
					"repository": map[string]any{
						"refs": map[string]any{
							"nodes": []any{
								map[string]any{
									"name": "main",
									"target": map[string]any{
										"committedDate": "2024-01-20T10:00:00Z",
									},
								},
								map[string]any{
									"name": "develop",
									"target": map[string]any{
										"committedDate": "2024-01-18T10:00:00Z",
									},
								},
							},
							"pageInfo": map[string]any{
								"hasNextPage": false,
								"endCursor":   "",
							},
						},
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalEndpoint := graphQLEndpoint
	graphQLEndpoint = server.URL
	t.Cleanup(func() {
		graphQLEndpoint = originalEndpoint
	})

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator := NewGitHubGenerator(config, nil, "owner", "repo", "test-token")
	defer generator.Close()

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Contains(t, versions, "v1.0.0")
	assert.Contains(t, versions, "v0.9.0")
	assert.Contains(t, versions, "main")
	assert.Contains(t, versions, "develop")
}

func TestGitHubGeneratorVersionSortKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		variables := body["variables"].(map[string]any)
		prefix := variables["prefix"].(string)

		var nodes []any
		if prefix == "refs/tags/" {
			nodes = []any{
				map[string]any{
					"name": "v1.0.0",
					"target": map[string]any{
						"committedDate": "2024-01-15T10:00:00Z",
					},
				},
			}
		} else {
			nodes = []any{
				map[string]any{
					"name": "main",
					"target": map[string]any{
						"committedDate": "2024-01-20T10:00:00Z",
					},
				},
			}
		}

		response := map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"refs": map[string]any{
						"nodes": nodes,
						"pageInfo": map[string]any{
							"hasNextPage": false,
							"endCursor":   "",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalEndpoint := graphQLEndpoint
	graphQLEndpoint = server.URL
	t.Cleanup(func() {
		graphQLEndpoint = originalEndpoint
	})

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator := NewGitHubGenerator(config, nil, "owner", "repo", "test-token")
	defer generator.Close()

	key1, err := generator.VersionSortKey("v1.0.0")
	assert.Nil(t, err)
	assert.Greater(t, key1, int64(0))

	key2, err := generator.VersionSortKey("main")
	assert.Nil(t, err)
	assert.Greater(t, key2, int64(0))

	assert.Greater(t, key2, key1, "main (2024-01-20) should have later timestamp than v1.0.0 (2024-01-15)")

	_, err = generator.VersionSortKey("nonexistent")
	assert.NotNil(t, err)
}

func TestGitHubGeneratorVersionsSortedByDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		variables := body["variables"].(map[string]any)
		prefix := variables["prefix"].(string)

		var nodes []any
		if prefix == "refs/tags/" {
			nodes = []any{
				map[string]any{
					"name": "v0.9.0",
					"target": map[string]any{
						"committedDate": "2024-01-10T10:00:00Z",
					},
				},
				map[string]any{
					"name": "v1.0.0",
					"target": map[string]any{
						"committedDate": "2024-01-15T10:00:00Z",
					},
				},
			}
		} else {
			nodes = []any{
				map[string]any{
					"name": "develop",
					"target": map[string]any{
						"committedDate": "2024-01-18T10:00:00Z",
					},
				},
				map[string]any{
					"name": "main",
					"target": map[string]any{
						"committedDate": "2024-01-20T10:00:00Z",
					},
				},
			}
		}

		response := map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"refs": map[string]any{
						"nodes": nodes,
						"pageInfo": map[string]any{
							"hasNextPage": false,
							"endCursor":   "",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalEndpoint := graphQLEndpoint
	graphQLEndpoint = server.URL
	t.Cleanup(func() {
		graphQLEndpoint = originalEndpoint
	})

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator := NewGitHubGenerator(config, nil, "owner", "repo", "test-token")
	defer generator.Close()

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, []string{"v0.9.0", "v1.0.0", "develop", "main"}, versions)
}

func TestGitHubGeneratorGraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	originalEndpoint := graphQLEndpoint
	graphQLEndpoint = server.URL
	t.Cleanup(func() {
		graphQLEndpoint = originalEndpoint
	})

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator := NewGitHubGenerator(config, nil, "owner", "repo", "test-token")
	defer generator.Close()

	_, err := generator.Versions()
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "GraphQL API returned status"))
}

func TestResolveGeneratorWithGitHubToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := resolveGenerator(config, nil)
	assert.Nil(t, err)
	assert.IsType(t, &GitHubGenerator{}, generator)

	t.Cleanup(func() {
		if g, ok := generator.(*GitHubGenerator); ok {
			g.Close()
		}
	})
}

func TestResolveGeneratorWithoutGitHubToken(t *testing.T) {
	os.Unsetenv("GITHUB_TOKEN")

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://github.com/owner/repo",
	}

	generator, err := resolveGenerator(config, nil)
	assert.Nil(t, err)
	assert.IsType(t, &GitGenerator{}, generator)

	t.Cleanup(func() {
		if g, ok := generator.(*GitGenerator); ok {
			g.Close()
		}
	})
}

func TestResolveGeneratorNonGitHubWithToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "https://gitlab.com/owner/repo",
	}

	generator, err := resolveGenerator(config, nil)
	assert.Nil(t, err)
	assert.IsType(t, &GitGenerator{}, generator)

	t.Cleanup(func() {
		if g, ok := generator.(*GitGenerator); ok {
			g.Close()
		}
	})
}

func TestResolveGeneratorFileURL(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	config := configuration.Configuration{
		Kind:       configuration.Git,
		Repository: "file:///path/to/repo",
	}

	generator, err := resolveGenerator(config, nil)
	assert.Nil(t, err)
	assert.IsType(t, &GitGenerator{}, generator)

	t.Cleanup(func() {
		if g, ok := generator.(*GitGenerator); ok {
			g.Close()
		}
	})
}
