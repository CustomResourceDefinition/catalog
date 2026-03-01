package generator

import (
	"fmt"

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

func NewPreparedGitGenerator(gitGenerator *GitGenerator, versions []versionInfo) *PreparedGitGenerator {
	return &PreparedGitGenerator{
		gitGenerator: gitGenerator,
		versions:     versions,
	}
}

func (g *PreparedGitGenerator) Versions() ([]string, error) {
	versions := make([]string, len(g.versions))
	for i, v := range g.versions {
		versions[i] = v.name
	}
	return versions, nil
}

func (g *PreparedGitGenerator) VersionSortKey(version string) (int64, error) {
	for _, v := range g.versions {
		if v.name == version {
			return v.timestamp, nil
		}
	}
	return 0, fmt.Errorf("version %q not found", version)
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
