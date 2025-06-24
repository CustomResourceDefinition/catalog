package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleCRD(t *testing.T) {
	out := t.TempDir()
	expected, err := os.ReadFile("testdata/test_v1.json")
	assert.Nil(t, err)

	g := Converter{
		Input:  "testdata/crd.yaml",
		Output: out,
	}
	err = g.Run()

	assert.Nil(t, err)
	actual, err := os.ReadFile(path.Join(out, "test_v1.json"))
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestArgoCRD(t *testing.T) {
	out := t.TempDir()
	expected, err := os.ReadFile("testdata/application_v1alpha1.json")
	assert.Nil(t, err)

	g := Converter{
		Input:  "testdata/argo-application-crd.yaml",
		Output: out,
	}
	err = g.Run()

	assert.Nil(t, err)
	actual, err := os.ReadFile(path.Join(out, "application_v1alpha1.json"))
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}
