package main

import (
	"bytes"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/command"
	"github.com/stretchr/testify/assert"
)

func TestNoArgumentsRunning(t *testing.T) {
	args := []string{"bin"}
	err := run(args, bytes.NewBuffer([]byte{}))
	assert.ErrorIs(t, err, command.ErrNoArguments)
}

func TestNoArgumentsParsing(t *testing.T) {
	args := []string{"bin"}
	_, err := parse(args, bytes.NewBuffer([]byte{}))
	assert.ErrorIs(t, err, command.ErrNoArguments)
}

func TestUnknownCommandParsing(t *testing.T) {
	tests := [][]string{
		{"horse"},
		{"monkey"},
		{"monkey", "horse"},
	}

	for _, test := range tests {
		args := []string{"bin"}
		args = append(args, test...)
		_, err := parse(args, bytes.NewBuffer([]byte{}))
		assert.ErrorIs(t, err, command.ErrUnknownCommand)
	}
}

func TestCompareCommandParsingIncorrectFlags(t *testing.T) {
	tests := [][]string{
		{"bin", commandCompare, "--not-a-flag"},
		{"bin", commandUpdate, "--not-a-flag"},
		{"bin", commandVerify, "--not-a-flag"},
	}
	for i, test := range tests {
		cmd, err := parse(test, bytes.NewBuffer([]byte{}))
		assert.NotNil(t, err, "index %d failed", i)
		assert.Nil(t, cmd, "index %d failed", i)
	}
}

func TestCompareCommandParsing(t *testing.T) {
	tests := [][]string{
		{"bin", commandCompare},
		{"bin", commandUpdate},
		{"bin", commandVerify},
	}
	for i, test := range tests {
		cmd, err := parse(test, bytes.NewBuffer([]byte{}))
		assert.Nil(t, err, "index %d failed", i)
		assert.NotNil(t, cmd, "index %d failed", i)
	}
}

func TestCommandRunningInvalidConfiguration(t *testing.T) {
	tests := [][]string{
		{"bin", commandCompare},
		{"bin", commandCompare, "--current", "."},
		{"bin", commandCompare, "--datreeio", "."},
		{"bin", commandCompare, "--current", ".", "--datreeio", "."},
		{"bin", commandUpdate},
		{"bin", commandUpdate, "--input", "testdata"},
		{"bin", commandUpdate, "--output", "testdata"},
		{"bin", commandUpdate, "--input", "testdata/does-not-exist", "--output", "testdata"},
		{"bin", commandUpdate, "--input", "testdata", "--output", "testdata/does-not-exist"},
		{"bin", commandVerify},
		{"bin", commandVerify, "--schema", "testdata/schema.json"},
		{"bin", commandVerify, "--file", "testdata/schema.json"},
		{"bin", commandVerify, "--schema", "testdata/does-not-exist.json", "--file", "testdata/schema.json"},
		{"bin", commandVerify, "--schema", "testdata/schema.json", "--file", "testdata/does-not-exist.json"},
	}

	for i, test := range tests {
		err := run(test, bytes.NewBuffer([]byte{}))
		assert.ErrorIs(t, err, command.ErrInvalidConfiguration, "index %d failed", i)
	}
}
