package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoArgumentsRunning(t *testing.T) {
	args := []string{"bin"}
	err := run(args)
	assert.ErrorIs(t, err, errNoArguments)
}

func TestNoArgumentsParsing(t *testing.T) {
	args := []string{"bin"}
	_, err := parse(args)
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
		_, err := parse(args)
		assert.ErrorIs(t, err, errUnknownCommand)
	}
}

func TestGenerateCommandParsingIncorrectFlags(t *testing.T) {
	args := []string{
		"bin",
		commandGenerate,
		"--not-a-flag",
	}
	cmd, err := parse(args)
	assert.NotNil(t, err)
	assert.Nil(t, cmd)
}

func TestGenerateCommandParsing(t *testing.T) {
	args := []string{
		"bin",
		commandGenerate,
	}
	cmd, err := parse(args)
	assert.Nil(t, err)
	assert.NotNil(t, cmd)
}

func TestGenerateCommandRunningInvalidConfiguration(t *testing.T) {
	tests := [][]string{
		{"bin", commandGenerate},
		{"bin", commandGenerate, "--current", "."},
		{"bin", commandGenerate, "--target", "."},
		{"bin", commandGenerate, "--current", ".", "--target", "."},
	}

	for _, test := range tests {
		err := run(test)
		assert.ErrorIs(t, err, errInvalidConfiguration)
	}
}
