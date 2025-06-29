package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoArgumentsRunning(t *testing.T) {
	args := []string{"bin"}
	err := run(args, bytes.NewBuffer([]byte{}))
	assert.ErrorIs(t, err, errNoArguments)
}

func TestNoArgumentsParsing(t *testing.T) {
	args := []string{"bin"}
	_, err := parse(args, bytes.NewBuffer([]byte{}))
	assert.ErrorIs(t, err, errNoArguments)
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
		assert.ErrorIs(t, err, errUnknownCommand)
	}
}

func TestCompareCommandParsingIncorrectFlags(t *testing.T) {
	args := []string{
		"bin",
		commandCompare,
		"--not-a-flag",
	}
	cmd, err := parse(args, bytes.NewBuffer([]byte{}))
	assert.NotNil(t, err)
	assert.Nil(t, cmd)
}

func TestCompareCommandParsing(t *testing.T) {
	args := []string{
		"bin",
		commandCompare,
	}
	cmd, err := parse(args, bytes.NewBuffer([]byte{}))
	assert.Nil(t, err)
	assert.NotNil(t, cmd)
}

func TestCommandRunningInvalidConfiguration(t *testing.T) {
	tests := [][]string{
		{"bin", commandCompare},
		{"bin", commandCompare, "--current", "."},
		{"bin", commandCompare, "--datreeio", "."},
		{"bin", commandCompare, "--current", ".", "--datreeio", "."},
		{"bin", commandConvert},
		{"bin", commandConvert, "--input", "testdata/crd.yaml"},
		{"bin", commandConvert, "--output", "testdata/"},
		{"bin", commandUpdate},
		{"bin", commandUpdate, "--input", "testdata"},
		{"bin", commandUpdate, "--output", "testdata"},
		{"bin", commandUpdate, "--input", "testdata/does-not-exist", "--output", "testdata"},
		{"bin", commandUpdate, "--input", "testdata", "--output", "testdata/does-not-exist"},
	}

	for i, test := range tests {
		err := run(test, bytes.NewBuffer([]byte{}))
		assert.ErrorIs(t, err, errInvalidConfiguration, "index %d failed", i)
	}
}
