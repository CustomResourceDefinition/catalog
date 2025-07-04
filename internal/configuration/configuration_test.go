package configuration

import (
	"fmt"
	"os"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalCorrectConfigurations(t *testing.T) {
	f, err := os.Open("testdata/configuration/correct.yaml")
	assert.Nil(t, err)
	defer f.Close()

	conf, err := UnmarshalConfigurations(f)
	assert.Nil(t, err)
	assert.NotNil(t, conf)
	assert.Equal(t, 6, len(*conf))
}

func TestUnmarshalInvalidConfigurations(t *testing.T) {
	f, err := os.Open("testdata/configuration/invalid.yaml")
	assert.Nil(t, err)
	defer f.Close()

	conf, err := UnmarshalConfigurations(f)
	assert.NotNil(t, err)
	assert.Nil(t, conf)
}

func TestUnmarshalMalformedConfigurations(t *testing.T) {
	f, err := os.Open("testdata/configuration/malformed.yaml")
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
