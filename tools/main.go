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
	datreeio := status.String("datreeio", "", "Path of checked out remote datreeio directory")
	current := status.String("current", "", "Path of local schema directory")
	ignore := status.String("ignore", "", "Path of ignore configuration")
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
			flags:    status,
			Current:  *current,
			Datreeio: *datreeio,
			Ignore:   *ignore,
			Out:      *out,
		}, nil
	default:
		return nil, errors.Join(errUnknownCommand, fmt.Errorf("unknown arguments: %s", arg))
	}
}
