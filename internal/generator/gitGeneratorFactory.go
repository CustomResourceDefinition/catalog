package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
)

type gitGeneratorFactory struct {
	config configuration.Configuration
	reader crd.CrdReader
	filter *regexp.Regexp
	logger io.Writer
}

func NewGitGeneratorFactory(config configuration.Configuration, reader crd.CrdReader, filter *regexp.Regexp, logger io.Writer) *gitGeneratorFactory {
	return &gitGeneratorFactory{
		config: config,
		reader: reader,
		filter: filter,
		logger: logger,
	}
}

func (f *gitGeneratorFactory) Build() (Generator, error) {
	generator := NewGitGenerator(f.config, f.reader, f.filter).(*GitGenerator)

	if isGitHub, owner, repo := isGitHubRepo(f.config.Repository); isGitHub {
		token := os.Getenv("GITHUB_TOKEN")
		if token != "" {
			versions, err := f.fetchGitHubVersions(owner, repo, token)
			if err == nil {
				return NewPreparedGitGenerator(generator, versions, f.filter), nil
			}
			fmt.Fprintf(f.logger, "Use git fallback instead of GitHub APIs for %s: %s\n", f.config.Name, err.Error())
		}
	}
	return generator, nil
}

func (f *gitGeneratorFactory) fetchGitHubVersions(owner, repo, token string) ([]versionInfo, error) {
	tags, err := f.fetchRefs(owner, repo, token, "refs/tags/")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	branches, err := f.fetchRefs(owner, repo, token, "refs/heads/")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch branches: %w", err)
	}

	versions := make([]versionInfo, 0, len(tags)+len(branches))
	versions = append(versions, tags...)
	versions = append(versions, branches...)

	return versions, nil
}

func (f *gitGeneratorFactory) fetchRefs(owner, repo, token, prefix string) ([]versionInfo, error) {
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
							... on Tag {
								target {
									... on Commit {
										committedDate
									}
									... on Tag {
										target {
											... on Commit {
												committedDate
											}
											... on Tag {
												target {
													... on Commit {
														committedDate
													}
												}
											}
											... on Tag {
												tagger {
													date
												}
											}
										}
										... on Tag {
											tagger {
												date
											}
										}
									}
								}
								... on Tag {
									tagger {
										date
									}
								}
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
			"owner":  owner,
			"repo":   repo,
		}
		if cursor != "" {
			variables["cursor"] = cursor
		}

		resp, err := f.graphQLRequest(query, token, variables)
		if err != nil {
			return nil, err
		}

		var result githubResponse

		if err := json.Unmarshal(resp, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response for prefix %s: %w", prefix, err)
		}

		for _, node := range result.Data.Repository.Refs.Nodes {
			dateStr := findCommitDate(node)
			ts, err := parseGitHubDate(dateStr)
			if err != nil {
				return nil, err
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

var graphQLEndpoint = "https://api.github.com/graphql"

func (f *gitGeneratorFactory) graphQLRequest(query, token string, variables map[string]any) ([]byte, error) {
	reqBody := map[string]any{
		"query":     query,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", graphQLEndpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
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
	if len(dateStr) == 0 {
		return 0, fmt.Errorf("unable to parse date from zero-length string")
	}
	ts, err := time.Parse(time.RFC3339Nano, dateStr)
	if err != nil {
		return 0, err
	}
	return ts.Unix(), nil
}

func findCommitDate(node githubNode) string {
	target := node.Target
	if len(target.CommittedDate) != 0 {
		return target.CommittedDate
	}
	if len(target.Tagger.Date) != 0 {
		return target.Tagger.Date
	}
	if len(target.Target.CommittedDate) != 0 {
		return target.Target.CommittedDate
	}
	if len(target.Target.Tagger.Date) != 0 {
		return target.Target.Tagger.Date
	}
	if len(target.Target.Target.CommittedDate) != 0 {
		return target.Target.Target.CommittedDate
	}
	if len(target.Target.Target.Tagger.Date) != 0 {
		return target.Target.Target.Tagger.Date
	}
	if len(target.Target.Target.Target.CommittedDate) != 0 {
		return target.Target.Target.Target.CommittedDate
	}
	if len(target.Target.Target.Target.Tagger.Date) != 0 {
		return target.Target.Target.Target.Tagger.Date
	}
	return ""
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
		Target        struct {
			CommittedDate string `json:"committedDate"`
			Target        struct {
				CommittedDate string `json:"committedDate"`
				Target        struct {
					CommittedDate string `json:"committedDate"`
					Tagger        struct {
						Date string `json:"date"`
					} `json:"tagger"`
				} `json:"target"`
				Tagger struct {
					Date string `json:"date"`
				} `json:"tagger"`
			} `json:"target"`
			Tagger struct {
				Date string `json:"date"`
			} `json:"tagger"`
		} `json:"target"`
		Tagger struct {
			Date string `json:"date"`
		} `json:"tagger"`
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
