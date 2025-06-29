package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalConfigurations(t *testing.T) {
	f, err := os.Open("../test/configuration.yaml")
	assert.Nil(t, err)
	defer f.Close()

	conf, err := UnmarshalConfigurations(f)
	assert.Nil(t, err)
	assert.NotNil(t, conf)
}
