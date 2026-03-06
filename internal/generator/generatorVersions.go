package generator

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

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

var versionFormat = regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)$`)

func normalizeVersion(matches [][]string) string {
	for _, match := range matches {
		if len(match) == 0 {
			continue
		}

		var version string
		if len(match) == 1 {
			version = match[0] // no capture groups - use entire match
		} else {
			version = match[1] // use matched capture group
		}

		var parts = versionFormat.FindAllStringSubmatch(version, -1)
		if len(parts) == 0 {
			continue
		}
		ints := make([]int, 3)
		for i := range 3 {
			n, _ := strconv.Atoi(parts[0][i+1])
			ints[i] = n
		}
		return fmt.Sprintf("v%d.%d.%d", ints[0], ints[1], ints[2])
	}

	return "v0.0.0"
}
