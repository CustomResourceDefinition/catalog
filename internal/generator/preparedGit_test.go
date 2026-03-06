package generator

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreparedGitGeneratorLatestVersionByTimestamp(t *testing.T) {
	versions := []versionInfo{
		{name: "v1.0.0", timestamp: 1705317600},
		{name: "v2.0.0", timestamp: 1706000000},
		{name: "main", timestamp: 1705749600},
	}

	generator := NewPreparedGitGenerator(nil, versions, regexp.MustCompile(".*"))

	result, err := generator.LatestVersion()
	assert.Nil(t, err)
	assert.Equal(t, "v2.0.0", result)
}
