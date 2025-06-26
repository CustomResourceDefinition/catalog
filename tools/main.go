package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type Command interface {
	Run() error
}

const commandCompare = "compare"
const commandConvert = "convert"
const commandUpdate = "update"

var (
	errNoArguments          = errors.New("expected a subcommand, but got no arguments")
	errUnknownCommand       = errors.New("expected a known subcommand")
	errInvalidConfiguration = errors.New("invalid configuration")
)

func main() {
	err := run(os.Args, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func run(args []string, logger io.Writer) error {
	cmd, err := parse(args, logger)
	if err != nil {
		return err
	}
	return cmd.Run()
}

func parse(args []string, logger io.Writer) (Command, error) {
	if len(args) < 2 {
		return nil, errNoArguments
	}

	arg := args[1]
	switch arg {
	case commandCompare:
		cmd := flag.NewFlagSet(commandCompare, flag.ContinueOnError)
		datreeio := cmd.String("datreeio", "", "Path of checked out remote datreeio directory")
		current := cmd.String("current", "", "Path of local schema directory")
		ignore := cmd.String("ignore", "", "Path of ignore configuration")
		output := cmd.String("out", "", "Path of output markdown file")
		cmd.SetOutput(logger)
		err := cmd.Parse(args[2:])
		if err != nil {
			return nil, err
		}
		return Comparer{
			flags:    cmd,
			Current:  *current,
			Datreeio: *datreeio,
			Ignore:   *ignore,
			Out:      *output,
		}, nil
	case commandConvert:
		cmd := flag.NewFlagSet(commandConvert, flag.ContinueOnError)
		input := cmd.String("input", "", "Path for CRD input file")
		output := cmd.String("output", "", "Directory for openapi schema output files")
		cmd.SetOutput(logger)
		err := cmd.Parse(args[2:])
		if err != nil {
			return nil, err
		}
		return Converter{
			flags:  cmd,
			Output: *output,
			Input:  *input,
			Logger: logger,
		}, nil
	case commandUpdate:
		cmd := flag.NewFlagSet(commandUpdate, flag.ContinueOnError)
		input := cmd.String("input", "", "Directory for prepared templated CRDs directory")
		output := cmd.String("output", "", "Directory for openapi schema output files")
		cmd.SetOutput(logger)
		err := cmd.Parse(args[2:])
		if err != nil {
			return nil, err
		}
		return Updater{
			flags:  cmd,
			Input:  *input,
			Output: *output,
			Logger: logger,
		}, nil
	default:
		return nil, errors.Join(errUnknownCommand, fmt.Errorf("unknown arguments: %s", arg))
	}
}
