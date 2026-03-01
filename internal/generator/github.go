package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
)

type GitHubGenerator struct {
	config       configuration.Configuration
	reader       crd.CrdReader
	owner        string
	repo         string
	token        string
	gitGenerator *GitGenerator
	versions     []versionInfo
}

type versionInfo struct {
	name      string
	timestamp int64
}

func NewGitHubGenerator(config configuration.Configuration, reader crd.CrdReader, owner, repo, token string) Generator {
	return &GitHubGenerator{
		config:       config,
		reader:       reader,
		owner:        owner,
		repo:         repo,
		token:        token,
		gitGenerator: NewGitGenerator(config, reader).(*GitGenerator),
	}
}

func (g *GitHubGenerator) Close() error {
	return g.gitGenerator.Close()
}

func (g *GitHubGenerator) Versions() ([]string, error) {
	if err := g.loadVersions(); err != nil {
		return nil, err
	}

	versions := make([]string, len(g.versions))
	for i, v := range g.versions {
		versions[i] = v.name
	}
	return versions, nil
}

func (g *GitHubGenerator) VersionSortKey(version string) (int64, error) {
	if err := g.loadVersions(); err != nil {
		return 0, err
	}

	for _, v := range g.versions {
		if v.name == version {
			return v.timestamp, nil
		}
	}
	return 0, fmt.Errorf("version %q not found", version)
}

func (g *GitHubGenerator) MetaData(version string) ([]crd.CrdMetaSchema, error) {
	return g.gitGenerator.MetaData(version)
}

func (g *GitHubGenerator) Crds(version string) ([]crd.Crd, error) {
	return g.gitGenerator.Crds(version)
}

func (g *GitHubGenerator) loadVersions() error {
	if len(g.versions) > 0 {
		return nil
	}

	tags, err := g.fetchTags()
	if err != nil {
		return fmt.Errorf("failed to fetch tags: %w", err)
	}

	branches, err := g.fetchBranches()
	if err != nil {
		return fmt.Errorf("failed to fetch branches: %w", err)
	}

	g.versions = make([]versionInfo, 0, len(tags)+len(branches))
	g.versions = append(g.versions, tags...)
	g.versions = append(g.versions, branches...)

	return nil
}

func (g *GitHubGenerator) fetchTags() ([]versionInfo, error) {
	return g.fetchRef("refs/tags/")
}

func (g *GitHubGenerator) fetchBranches() ([]versionInfo, error) {
	return g.fetchRef("refs/heads/")
}

func (g *GitHubGenerator) fetchRef(prefix string) ([]versionInfo, error) {
	query := `
		query($owner: String!, $repo: String!, $prefix: String!, $cursor: String) {
			repository(owner: $owner, name: $repo) {
				refs(refPrefix: $prefix, first: 100, after: $cursor) {
					nodes {
						name
						target {
							... on Commit {
								committedDate
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	`

	var versions []versionInfo
	cursor := ""
	hasNextPage := true

	for hasNextPage {
		variables := map[string]any{
			"prefix": prefix,
			"owner":  g.owner,
			"repo":   g.repo,
		}
		if cursor != "" {
			variables["cursor"] = cursor
		}

		resp, err := g.graphQLRequest(query, variables)
		if err != nil {
			return nil, err
		}

		var result githubResponse

		if err := json.Unmarshal(resp, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response for prefix %s: %w", prefix, err)
		}

		for _, node := range result.Data.Repository.Refs.Nodes {
			ts, err := parseGitHubDate(node.Target.CommittedDate)
			if err != nil {
				continue
			}
			versions = append(versions, versionInfo{
				name:      node.Name,
				timestamp: ts,
			})
		}

		hasNextPage = result.Data.Repository.Refs.PageInfo.HasNextPage
		cursor = result.Data.Repository.Refs.PageInfo.EndCursor

		if cursor == "" {
			hasNextPage = false
		}
	}

	return versions, nil
}

func (g *GitHubGenerator) graphQLRequest(query string, variables map[string]any) ([]byte, error) {
	reqBody := map[string]any{
		"query":     query,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", "https://api.github.com/graphql", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL API returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func parseGitHubDate(dateStr string) (int64, error) {
	ts, err := time.Parse(time.RFC3339Nano, dateStr)
	if err != nil {
		return 0, err
	}
	return ts.Unix(), nil
}

func isGitHubRepo(repoURL string) (bool, string, string) {
	repoURL = strings.TrimSuffix(repoURL, ".git")

	re := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+)/?`)
	matches := re.FindStringSubmatch(repoURL)
	if len(matches) == 3 {
		return true, matches[1], matches[2]
	}

	return false, "", ""
}

type githubNode struct {
	Name   string `json:"name"`
	Target struct {
		CommittedDate string `json:"committedDate"`
	} `json:"target"`
}

type githubPageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type githubRefs struct {
	Nodes    []githubNode   `json:"nodes"`
	PageInfo githubPageInfo `json:"pageInfo"`
}

type githubRepository struct {
	Refs githubRefs `json:"refs"`
}

type githubData struct {
	Repository githubRepository `json:"repository"`
}

type githubResponse struct {
	Data githubData `json:"data"`
}
