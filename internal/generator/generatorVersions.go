package generator

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/CustomResourceDefinition/catalog/internal/semver"
)

type GeneratorVersions struct{}

func (v *GeneratorVersions) latest(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions are available")
	}
	return versions[0], nil
}

func (v *GeneratorVersions) semverSort(versions []string, filter *regexp.Regexp) ([]string, error) {
	if filter == nil {
		return nil, fmt.Errorf("filter is required")
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
