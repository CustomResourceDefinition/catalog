package generator

import (
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

	result, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, []string{"v1.0.0", "main", "develop"}, result)
}

func TestPreparedGitGeneratorVersionSortKey(t *testing.T) {
	versions := []versionInfo{
		{name: "v1.0.0", timestamp: 1705317600},
		{name: "main", timestamp: 1705749600},
	}

	generator := NewPreparedGitGenerator(nil, versions)

	key, err := generator.VersionSortKey("v1.0.0")
	assert.Nil(t, err)
	assert.Equal(t, int64(1705317600), key)

	key, err = generator.VersionSortKey("main")
	assert.Nil(t, err)
	assert.Equal(t, int64(1705749600), key)

	_, err = generator.VersionSortKey("nonexistent")
	assert.NotNil(t, err)
}

func TestPreparedGitGeneratorVersionSortKeyEmpty(t *testing.T) {
	versions := []versionInfo{}

	generator := NewPreparedGitGenerator(nil, versions)

	_, err := generator.VersionSortKey("v1.0.0")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "not found")
}
