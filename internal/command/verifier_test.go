package command

import (
	"fmt"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

type verifierScenario struct {
	file, schema string
}

func TestVerifierValidatesKnownSamples(t *testing.T) {
	tests := []verifierScenario{
		{
			file:   "testdata/verifier/object-data.json",
			schema: "testdata/verifier/object-schema.json",
		},
		{
			file:   "testdata/verifier/array-data.json",
			schema: "testdata/verifier/array-schema.json",
		},
	}

	for i, test := range tests {
		cmd := NewVerifier(test.schema, test.file, nil)

		err := cmd.Run()

		assert.Nil(t, err, "index %d failed", i)
	}
}

func TestVerifierRejectsKnownSamples(t *testing.T) {
	tests := []verifierScenario{
		{
			file:   "testdata/verifier/object-data.json",
			schema: "testdata/verifier/array-schema.json",
		},
		{
			file:   "testdata/verifier/array-data.json",
			schema: "testdata/verifier/object-schema.json",
		},
		{
			file:   "testdata/verifier/conf.ini",
			schema: "testdata/verifier/object-schema.json",
		},
		{
			file:   "testdata/verifier/conf.ini",
			schema: "testdata/verifier/array-schema.json",
		},
	}

	for i, test := range tests {
		cmd := NewVerifier(test.schema, test.file, nil)

		err := cmd.Run()

		assert.NotNil(t, err, "index %d failed", i)
	}
}

func TestVerifierValidateWithWrongPaths(t *testing.T) {
	cmd := NewVerifier("testdata/verifier/does-not-exist.json", "testdata/verifier/object-data.json", nil)
	err := cmd.validate()
	assert.NotNil(t, err)
}

func TestReadFileWithWrongPaths(t *testing.T) {
	bytes, err := readFile("testdata/verifier/does-not-exist.json")

	assert.Nil(t, bytes)
	assert.NotNil(t, err)
}

func TestReadBytesWithReadIssue(t *testing.T) {
	specificErr := fmt.Errorf("specific error")
	reader := iotest.ErrReader(specificErr)

	bytes, err := readBytes(reader)

	assert.Nil(t, bytes)
	assert.ErrorIs(t, err, specificErr)
}

func TestUnpackWithWrongFormat(t *testing.T) {
	bytes, err := readFile("testdata/verifier/conf.ini")
	assert.Nil(t, err)

	o, err := unpack("conf.ini", bytes)

	assert.Nil(t, o)
	assert.NotNil(t, err)
}
