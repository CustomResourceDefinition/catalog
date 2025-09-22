package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/CustomResourceDefinition/catalog/internal/command"
)

const commandCompare = "compare"
const commandUpdate = "update"
const commandVerify = "verify"

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

func parse(args []string, logger io.Writer) (command.Command, error) {
	if len(args) < 2 {
		return nil, command.ErrNoArguments
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
		return command.NewComparer(*datreeio, *current, *output, *ignore, cmd), nil
	case commandUpdate:
		cmd := flag.NewFlagSet(commandUpdate, flag.ContinueOnError)
		configuration := cmd.String("configuration", "", "Path to configuration file for CRD sources")
		schema := cmd.String("output", "", "Path of directory for openapi schema output files")
		definitions := cmd.String("definitions", "", "Path of directory for definition output files")
		cmd.SetOutput(logger)
		err := cmd.Parse(args[2:])
		if err != nil {
			return nil, err
		}
		return command.NewUpdater(*configuration, *schema, *definitions, logger, cmd), nil
	case commandVerify:
		cmd := flag.NewFlagSet(commandVerify, flag.ContinueOnError)
		schema := cmd.String("schema", "", "Path of jsonschema file to use")
		file := cmd.String("file", "", "Path of file to check")
		cmd.SetOutput(logger)
		err := cmd.Parse(args[2:])
		if err != nil {
			return nil, err
		}
		return command.NewVerifier(*schema, *file, cmd), nil
	default:
		return nil, errors.Join(command.ErrUnknownCommand, fmt.Errorf("unknown arguments: %s", arg))
	}
}
