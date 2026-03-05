package generator

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/CustomResourceDefinition/catalog/internal/crd"
)

type PreparedGitGenerator struct {
	gitGenerator *GitGenerator
	versions     []versionInfo
}

type versionInfo struct {
	name      string
	timestamp int64
}

func NewPreparedGitGenerator(gitGenerator *GitGenerator, versions []versionInfo) Generator {
	return &PreparedGitGenerator{
		gitGenerator: gitGenerator,
		versions:     versions,
	}
}

func (g *PreparedGitGenerator) LatestVersion(filter *regexp.Regexp) (string, error) {
	filtered := make([]versionInfo, 0)
	for _, v := range g.versions {
		if filter.MatchString(v.name) {
			filtered = append(filtered, v)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].timestamp > filtered[j].timestamp
	})

	if len(filtered) == 0 {
		return "", fmt.Errorf("no versions are available")
	}

	return filtered[0].name, nil
}

func (g *PreparedGitGenerator) Versions(filter *regexp.Regexp) ([]string, error) {
	return g.gitGenerator.Versions(filter)
}

func (g *PreparedGitGenerator) MetaData(version string) ([]crd.CrdMetaSchema, error) {
	return g.gitGenerator.MetaData(version)
}

func (g *PreparedGitGenerator) Crds(version string) ([]crd.Crd, error) {
	return g.gitGenerator.Crds(version)
}

func (g *PreparedGitGenerator) Close() error {
	return g.gitGenerator.Close()
}
