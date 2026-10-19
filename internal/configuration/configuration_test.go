package configuration

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

func TestConfigurationIsSortedAndValid(t *testing.T) {
	f, err := os.Open("../../configuration.yaml")
	assert.Nil(t, err)
	defer f.Close()

	conf, err := UnmarshalConfigurations(f)
	assert.Nil(t, err)
	assert.NotNil(t, conf)

	unsorted := make([]string, 0)
	for _, c := range conf {
		unsorted = append(unsorted, c.Name)
		hasFilePrefix := strings.HasPrefix(strings.ToLower(c.Repository), "file://")
		assert.False(t, hasFilePrefix, "invalid registry scheme for '%s'", c.Name)
	}
	sorted := make([]string, len(unsorted))
	copy(sorted, unsorted)
	slices.Sort(sorted)

	assert.Equal(t, sorted, unsorted)
}

func TestConfigurationAllPatternsAreValidRegex(t *testing.T) {
	f, err := os.Open("../../configuration.yaml")
	assert.Nil(t, err)
	defer f.Close()

	conf, err := UnmarshalConfigurations(f)
	assert.Nil(t, err)
	assert.NotNil(t, conf)

	for _, c := range conf {
		if c.VersionPattern != "" {
			_, err := regexp.Compile(c.VersionPattern)
			assert.Nil(t, err, "invalid regex for '%s': %s", c.Name, c.VersionPattern)
		}
	}
}

func TestUnmarshalCorrectConfigurations(t *testing.T) {
	f, err := os.Open("testdata/correct.yaml")
	assert.Nil(t, err)
	defer f.Close()

	conf, err := UnmarshalConfigurations(f)
	assert.Nil(t, err)
	assert.NotNil(t, conf)
	assert.Equal(t, 6, len(conf))
}

func TestUnmarshalSkippedConfigurations(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		skipped bool
		wantErr bool
	}{
		{name: "omitted", value: ""},
		{name: "false", value: "skipped: false"},
		{name: "true", value: "skipped: true", skipped: true},
		{name: "string", value: `skipped: "true"`, wantErr: true},
		{name: "null", value: "skipped: null", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := fmt.Sprintf("- kind: http\n  name: test\n  apiGroups: [example.com]\n  crds:\n    - version: 1.0.0\n      baseUri: https://example.com\n      paths: [crd.yaml]\n  %s\n", test.value)
			configs, err := UnmarshalConfigurations(strings.NewReader(input))
			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, configs)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, configs, 1)
			assert.Equal(t, test.skipped, configs[0].Skipped)
		})
	}
}

func TestUnmarshalInvalidConfigurations(t *testing.T) {
	f, err := os.Open("testdata/invalid.yaml")
	assert.Nil(t, err)
	defer f.Close()

	conf, err := UnmarshalConfigurations(f)
	assert.NotNil(t, err)
	assert.Nil(t, conf)
}

func TestUnmarshalMalformedConfigurations(t *testing.T) {
	f, err := os.Open("testdata/malformed.yaml")
	assert.Nil(t, err)
	defer f.Close()

	conf, err := UnmarshalConfigurations(f)
	assert.NotNil(t, err)
	assert.Nil(t, conf)
}

func TestUnmarshalBrokenFileConfigurations(t *testing.T) {
	reader := iotest.ErrReader(fmt.Errorf("expected error"))

	conf, err := UnmarshalConfigurations(reader)
	assert.NotNil(t, err)
	assert.Nil(t, conf)
}

func TestUnmarshalEmptyConfigurations(t *testing.T) {
	conf, err := UnmarshalConfigurations(nil)
	assert.NotNil(t, err)
	assert.Nil(t, conf)
}

func TestResolvingConfigurationValues(t *testing.T) {
	cv := []ConfigurationValues{
		{
			Version:    "0.0.0",
			ValuesFile: "",
		},
		{
			Version:    "2.0.0",
			ValuesFile: "output: 2",
		},
		{
			Version:    "1.0.0",
			ValuesFile: "output: 1",
		},
	}

	c := Configuration{
		Values: cv,
	}
	v, err := c.ValuesFile("0.0.0")
	assert.Nil(t, err)
	assert.Nil(t, v)

	v, err = c.ValuesFile("0.9.1")
	assert.Nil(t, err)
	assert.Nil(t, v)

	v, err = c.ValuesFile("1.9.1")
	assert.Nil(t, err)
	assert.Equal(t, v, map[string]any{"output": 1})

	v, err = c.ValuesFile("2.0.0")
	assert.Nil(t, err)
	assert.Equal(t, v, map[string]any{"output": 2})

	v, err = c.ValuesFile("2.0.1")
	assert.Nil(t, err)
	assert.Equal(t, v, map[string]any{"output": 2})
}
