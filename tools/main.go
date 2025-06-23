package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

type Command interface {
	Run() error
}

const commandConvert = "convert"
const commandGenerate = "generate-status"

var (
	errNoArguments          = errors.New("expected a subcommand, but got no arguments")
	errUnknownCommand       = errors.New("expected a known subcommand")
	errInvalidConfiguration = errors.New("invalid configuration")
)

func main() {
	err := run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	cmd, err := parse(args)
	if err != nil {
		return err
	}
	return cmd.Run()
}

func parse(args []string) (Command, error) {
	if len(args) < 2 {
		return nil, errNoArguments
	}

	arg := args[1]
	switch arg {
	case commandGenerate:
		cmd := flag.NewFlagSet(commandGenerate, flag.ContinueOnError)
		datreeio := cmd.String("datreeio", "", "Path of checked out remote datreeio directory")
		current := cmd.String("current", "", "Path of local schema directory")
		ignore := cmd.String("ignore", "", "Path of ignore configuration")
		out := cmd.String("out", "", "Path of output markdown file")
		err := cmd.Parse(args[2:])
		if err != nil {
			return nil, err
		}
		return StatusGenerator{
			flags:    cmd,
			Current:  *current,
			Datreeio: *datreeio,
			Ignore:   *ignore,
			Out:      *out,
		}, nil
	case commandConvert:
		cmd := flag.NewFlagSet(commandConvert, flag.ContinueOnError)
		input := cmd.String("input", "", "Path for CRD input file")
		output := cmd.String("output", "", "Directory for openapi schema output files")
		err := cmd.Parse(args[2:])
		if err != nil {
			return nil, err
		}
		return Converter{
			flags:  cmd,
			Output: *output,
			Input:  *input,
		}, nil
	default:
		return nil, errors.Join(errUnknownCommand, fmt.Errorf("unknown arguments: %s", arg))
	}
}
