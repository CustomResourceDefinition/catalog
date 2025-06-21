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

const commandGenerate = "generate-status"

var (
	errNoArguments    = errors.New("expected a subcommand, but got no arguments")
	errUnknownCommand = errors.New("expected a known subcommand")
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
	status := flag.NewFlagSet("generate-status", flag.ContinueOnError)
	remote := status.String("remote", "", "Path of checked out remote target directory")
	current := status.String("current", "", "Path of local schema directory")
	out := status.String("out", "", "Path of output markdown file")

	if len(args) < 2 {
		return nil, errNoArguments
	}

	arg := args[1]
	switch arg {
	case commandGenerate:
		err := status.Parse(args[2:])
		if err != nil {
			return nil, err
		}
		return StatusGenerator{
			flags:   status,
			Out:     *out,
			Remote:  *remote,
			Current: *current,
		}, nil
	default:
		return nil, errors.Join(errUnknownCommand, fmt.Errorf("unknown arguments: %s", arg))
	}
}
