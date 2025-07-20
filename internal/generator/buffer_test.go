package generator_test

import (
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/generator"
	"github.com/stretchr/testify/assert"
)

func TestBuffer_WriteString(t *testing.T) {
	var b generator.Buffer

	n, err := b.WriteString("Hello, world!")
	assert.NoError(t, err)
	assert.Equal(t, len("Hello, world!"), n)

	expected := "Hello, world!\n---\n"
	assert.Equal(t, expected, string(b.Bytes()))
}

func TestBuffer_Write(t *testing.T) {
	var b generator.Buffer

	data := []byte("Some binary data")
	n, err := b.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	expected := "Some binary data\n---\n"
	assert.Equal(t, expected, string(b.Bytes()))
}

func TestBuffer_MultipleWrites(t *testing.T) {
	var b generator.Buffer

	_, _ = b.WriteString("First")
	_, _ = b.Write([]byte("Second"))
	_, _ = b.WriteString("Third")

	expected := "First\n---\nSecond\n---\nThird\n---\n"
	assert.Equal(t, expected, string(b.Bytes()))
}
