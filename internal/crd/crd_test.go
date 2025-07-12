package crd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var logger = bytes.NewBuffer([]byte{})

func TestCrdMetadata(t *testing.T) {
	reader, err := NewCrdReader(logger)
	assert.Nil(t, err)

	f, err := os.Open("testdata/argo-application-crd.yaml")
	assert.Nil(t, err)
	defer f.Close()

	crds, err := reader.Read(f, "argo-application-crd.yaml")
	assert.Nil(t, err)

	assert.NotNil(t, crds)
	assert.Equal(t, 1, len(crds))

	crd := crds[0]

	metadata, err := crd.MetaSchema()
	assert.Nil(t, err)

	assert.NotNil(t, metadata)
	assert.Equal(t, 1, len(metadata))

	item := metadata[0]
	assert.Equal(t, "argoproj.io", item.Group)
	assert.Equal(t, "application", item.Kind)
	assert.Equal(t, "v1alpha1", item.Version)
}

func TestCrdSchema(t *testing.T) {
	reader, err := NewCrdReader(logger)
	assert.Nil(t, err)

	f, err := os.Open("testdata/argo-application-crd.yaml")
	assert.Nil(t, err)
	defer f.Close()

	crds, err := reader.Read(f, "argo-application-crd.yaml")
	assert.Nil(t, err)

	assert.NotNil(t, crds)
	assert.Equal(t, 1, len(crds))

	crd := crds[0]

	schemas, err := crd.Schema()
	assert.Nil(t, err)

	assert.NotNil(t, schemas)
	assert.Equal(t, 1, len(schemas))

	schema := schemas[0]
	assert.Equal(t, "argoproj.io", schema.Group)
	assert.Equal(t, "application", schema.Kind)
	assert.Equal(t, "v1alpha1", schema.Version)
}

func TestCrdReaderRead(t *testing.T) {
	logger := bytes.NewBuffer([]byte{})

	reader, err := NewCrdReader(logger)
	assert.Nil(t, err)

	f, err := os.Open("testdata/documents.yaml")
	assert.Nil(t, err)
	defer f.Close()

	crds, err := reader.Read(f, "documents.yaml")
	assert.Nil(t, err)

	assert.NotNil(t, crds)
	assert.Equal(t, 1, len(crds))
}

type testScenario struct {
	schema   CrdMetaSchema
	expected string
}

func TestCrdMetaSchemaFilePath(t *testing.T) {
	tests := []testScenario{
		{
			schema: CrdMetaSchema{
				Group:   "",
				Kind:    "",
				Version: "",
			},
			expected: "/_.json",
		},
		{
			schema: CrdMetaSchema{
				Group:   "argoproj.io",
				Kind:    "application",
				Version: "v1alpha1",
			},
			expected: "argoproj.io/application_v1alpha1.json",
		},
		{
			schema: CrdMetaSchema{
				Group:   "a",
				Kind:    "b",
				Version: "c",
			},
			expected: "a/b_c.json",
		},
	}

	for i, test := range tests {
		assert.Equal(t, test.expected, test.schema.Filepath(), "index %d failed for %s", i, test.expected)
	}
}
