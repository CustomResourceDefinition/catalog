package generator

import (
	"fmt"
	"io"
	"regexp"
	"sort"

	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/semver"
)

type Generator interface {
	AllVersions() ([]string, error)
	Versions(filter *regexp.Regexp) ([]string, error)
	LatestVersion(filter *regexp.Regexp) (string, error)
	MetaData(version string) ([]crd.CrdMetaSchema, error)
	Crds(version string) ([]crd.Crd, error)
	io.Closer
}

type baseGenerator struct{}

func (b *baseGenerator) LatestVersion(filter *regexp.Regexp) (string, error) {
	versions, err := b.Versions(filter)
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions are available")
	}
	return versions[0], nil
}

func (b baseGenerator) Versions(filter *regexp.Regexp) ([]string, error) {
	versions, err := b.AllVersions()
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0)
	for _, v := range versions {
		if filter.MatchString(v) {
			filtered = append(filtered, v)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		a := normalizeVersion(filter.FindAllStringSubmatch(filtered[i], -1))
		b := normalizeVersion(filter.FindAllStringSubmatch(filtered[j], -1))
		return semver.Compare(a, b) > 0
	})
	return filtered, nil
}

func (b baseGenerator) AllVersions() ([]string, error) {
	panic("base implementation should not be called")
}
