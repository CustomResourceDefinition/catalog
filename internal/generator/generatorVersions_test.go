package generator

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratorVersionsSemverSortNoFilter(t *testing.T) {
	gv := GeneratorVersions{}

	_, err := gv.semverSort([]string{"1.0.0", "2.0.0", "1.5.0"}, nil)
	assert.NotNil(t, err)
}

func TestGeneratorVersionsSemverSortWithFilter(t *testing.T) {
	gv := GeneratorVersions{}

	filter := regexp.MustCompile(`^v([0-9]+\.[0-9]+\.[0-9]+)$`)
	versions, err := gv.semverSort([]string{"v1.0.0", "v2.0.0", "main"}, filter)
	assert.Nil(t, err)
	assert.Equal(t, []string{"v2.0.0", "v1.0.0"}, versions)
}

func TestGeneratorVersionsSemverSortFiltersCorrectly(t *testing.T) {
	gv := GeneratorVersions{}

	filter := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	versions, err := gv.semverSort([]string{"1.0.0", "main", "2.0.0", "master"}, filter)
	assert.Nil(t, err)
	assert.Equal(t, []string{"2.0.0", "1.0.0"}, versions)
}

func TestGeneratorVersionsSemverSortSortsDescending(t *testing.T) {
	gv := GeneratorVersions{}

	filter := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	versions, err := gv.semverSort([]string{"1.0.0", "2.10.0", "2.2.0", "1.5.0", "1.01.01"}, filter)
	assert.Nil(t, err)
	assert.Equal(t, []string{"2.10.0", "2.2.0", "1.5.0", "1.01.01", "1.0.0"}, versions)
}

func TestGeneratorVersionsFiltering(t *testing.T) {
	gv := GeneratorVersions{}

	tests := []struct {
		name             string
		versions         []string
		expectedVersions []string
		pattern          string
	}{
		{
			name:             "no filter matches all",
			versions:         []string{"2.0.0", "1.3.0", "1.0.0"},
			expectedVersions: []string{"2.0.0", "1.3.0", "1.0.0"},
			pattern:          ".*",
		},
		{
			name:             "v prefix versions no pattern",
			versions:         []string{"v2.0.0", "v1.3.0", "v1.0.0"},
			expectedVersions: []string{},
			pattern:          "",
		},
		{
			name:             "v prefix with pattern",
			versions:         []string{"v2.0.0", "v1.3.0", "v1.0.0"},
			expectedVersions: []string{"v2.0.0", "v1.3.0", "v1.0.0"},
			pattern:          `^v([0-9]+\.[0-9]+\.[0-9]+)$`,
		},
		{
			name:             "optional v prefix",
			versions:         []string{"2.0.0", "v1.3.0", "v1.0.0"},
			expectedVersions: []string{"2.0.0", "v1.3.0", "v1.0.0"},
			pattern:          `^v?([0-9]+\.[0-9]+\.[0-9]+)$`,
		},
		{
			name:             "optional v prefix filters non-matching",
			versions:         []string{"2.0.0", "v1.3v", "v1.0.0"},
			expectedVersions: []string{"2.0.0", "v1.0.0"},
			pattern:          `^v?([0-9]+\.[0-9]+\.[0-9]+)$`,
		},
		{
			name:             "prerelease versions",
			versions:         []string{"2.0.0-2", "1.3.0-1892", "1.0.0-01"},
			expectedVersions: []string{"2.0.0-2", "1.3.0-1892", "1.0.0-01"},
			pattern:          `^([0-9]+\.[0-9]+\.[0-9]+-\d+)$`,
		},
		{
			name:             "version with build metadata",
			versions:         []string{"v1.33.2+k0s.0"},
			expectedVersions: []string{"v1.33.2+k0s.0"},
			pattern:          `^v([0-9]+\.[0-9]+\.[0-9]+\+k0s\.0)$`,
		},
		{
			name:             "branch names only",
			versions:         []string{"main", "v1.0", "master"},
			expectedVersions: []string{"main", "master"},
			pattern:          `^(main|master)$`,
		},
		{
			name:             "mixed versions and branches",
			versions:         []string{"main", "v1.0.0", "2.0.0", "master"},
			expectedVersions: []string{"2.0.0", "main", "master"},
			pattern:          `^([0-9]+\.[0-9]+\.[0-9]+)|(main|master)$`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var filter *regexp.Regexp
			if test.pattern != "" {
				filter = regexp.MustCompile(test.pattern)
			} else {
				filter = regexp.MustCompile(".*")
			}

			versions, err := gv.semverSort(test.versions, filter)
			assert.Nil(t, err)
			assert.Equal(t, test.expectedVersions, versions)
		})
	}
}

func TestGeneratorVersionsLatestEmpty(t *testing.T) {
	gv := GeneratorVersions{}

	version, err := gv.latest([]string{})
	assert.NotNil(t, err)
	assert.Equal(t, "", version)
}

func TestGeneratorVersionsLatestReturnsFirst(t *testing.T) {
	gv := GeneratorVersions{}

	version, err := gv.latest([]string{"1.0.0", "2.0.0", "1.5.0"})
	assert.Nil(t, err)
	assert.Equal(t, "1.0.0", version)
}
