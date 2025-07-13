package command

import "errors"

var (
	ErrNoArguments          = errors.New("expected a subcommand, but got no arguments")
	ErrUnknownCommand       = errors.New("expected a known subcommand")
	ErrInvalidConfiguration = errors.New("invalid configuration")
)

type Command interface {
	Run() error
}
