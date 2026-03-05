package generator

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreparedGitGeneratorVersions(t *testing.T) {
	versions := []versionInfo{
		{name: "v1.0.0", timestamp: 1705317600},
		{name: "main", timestamp: 1705749600},
		{name: "develop", timestamp: 1705569600},
	}

	generator := NewPreparedGitGenerator(nil, versions)

	result, err := generator.Versions(regexp.MustCompile(".*"))
	assert.Nil(t, err)
	assert.Equal(t, []string{"v1.0.0", "main", "develop"}, result)
}

func TestPreparedGitGeneratorLatestVersionByTimestamp(t *testing.T) {
	versions := []versionInfo{
		{name: "v1.0.0", timestamp: 1705317600},
		{name: "v2.0.0", timestamp: 1706000000},
		{name: "main", timestamp: 1705749600},
	}

	generator := NewPreparedGitGenerator(nil, versions)

	result, err := generator.LatestVersion(regexp.MustCompile(".*"))
	assert.Nil(t, err)
	assert.Equal(t, "v2.0.0", result)
}
