package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FIXME: more tests
func TestHelm(t *testing.T) {
	s, err := pullHelmChart("https://1password.github.io/connect-helm-charts", "", "1password", "connect")
	assert.Nil(t, err)
	assert.NotNil(t, s)

	o, err := renderChart(s, "release", "namespace")
	assert.Nil(t, err)
	assert.NotNil(t, o)
}

func TestHelmOci(t *testing.T) {
	out, err := pullOCIChart("oci://registry.developers.crunchydata.com/crunchydata/pgo", "") // 5.8.2
	assert.Nil(t, err)

	o, err := renderChart(out, "release", "namespace")
	assert.Nil(t, err)
	assert.NotNil(t, o)
}
