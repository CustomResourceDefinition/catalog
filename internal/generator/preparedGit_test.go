package generator

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreparedGitGeneratorLatestVersionByTimestamp(t *testing.T) {
	versions := []versionInfo{
		{name: "a", timestamp: 1705317600},
		{name: "b", timestamp: 1706000000},
		{name: "c", timestamp: 1705749600},
	}

	generator := NewPreparedGitGenerator(nil, versions, regexp.MustCompile(".*"))

	result, err := generator.LatestVersion()
	assert.Nil(t, err)
	assert.Equal(t, "b", result)
}

func TestPreparedGitGeneratorLatestVersionByTimestampNotSemver(t *testing.T) {
	versions := []versionInfo{
		{name: "1.0.0", timestamp: 1705000000},
		{name: "2.0.0", timestamp: 1706000000},
		{name: "10.0.0", timestamp: 1704000000},
	}

	generator := NewPreparedGitGenerator(nil, versions, regexp.MustCompile(".*"))

	result, err := generator.LatestVersion()
	assert.Nil(t, err)
	assert.Equal(t, "2.0.0", result)
}

func TestPreparedGitGeneratorLatestVersionNoMatchingFilter(t *testing.T) {
	versions := []versionInfo{
		{name: "v1.0.0", timestamp: 1705317600},
		{name: "v2.0.0", timestamp: 1706000000},
	}

	filter := regexp.MustCompile(`^main$`)
	generator := NewPreparedGitGenerator(nil, versions, filter)

	_, err := generator.LatestVersion()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no versions are available")
}
