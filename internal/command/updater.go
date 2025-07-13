package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/generator"
)

type Updater struct {
	Configuration, Output string
	Logger                io.Writer
	flags                 *flag.FlagSet
	reader                crd.CrdReader
}

func NewUpdater(configuration, output string, logger io.Writer, flags *flag.FlagSet) Updater {
	return Updater{
		flags:         flags,
		Configuration: configuration,
		Output:        output,
		Logger:        logger,
	}
}

func (cmd Updater) Run() error {
	if err := cmd.validate(); err != nil {
		if cmd.flags != nil {
			cmd.flags.Usage()
		}
		return errors.Join(ErrInvalidConfiguration, err)
	}

	if err := cmd.initialize(); err != nil {
		return err
	}

	configurations, err := readConfiguration(cmd.Configuration)
	if err != nil {
		return err
	}

	reader, err := crd.NewCrdReader(cmd.Logger)
	if err != nil {
		return err
	}

	for _, config := range splitConfigurations(configurations) {
		runtime.GC()

		build, err := generator.NewBuilder(config, reader, cmd.Output, cmd.Logger)
		if err != nil {
			continue
		}

		err = build.Build()
		if err != nil {
			continue
		}
	}

	return nil
}

func (cmd Updater) validate() error {
	if f, err := os.Stat(cmd.Output); err != nil || len(cmd.Output) == 0 || !f.IsDir() {
		return fmt.Errorf("'%s' is not a valid directory path", cmd.Output)
	}

	if f, err := os.Stat(cmd.Configuration); err != nil || len(cmd.Configuration) == 0 || f.IsDir() {
		return fmt.Errorf("'%s' is not a valid file path", cmd.Configuration)
	}

	return nil
}

func (cmd *Updater) initialize() error {
	if cmd.Logger == nil {
		cmd.Logger = os.Stderr
	}

	reader, err := crd.NewCrdReader(cmd.Logger)
	if err != nil {
		return err
	}
	cmd.reader = reader

	return nil
}

func readConfiguration(file string) ([]configuration.Configuration, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	configs, err := configuration.UnmarshalConfigurations(f)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// splitConfigurations will duplicate all entries in Entries, so each configuration only has a single entry in Entries
func splitConfigurations(configurations []configuration.Configuration) []configuration.Configuration {
	updated := make([]configuration.Configuration, 0)
	for _, c := range configurations {
		if len(c.Entries) == 0 {
			updated = append(updated, c)
			continue
		}
		for _, e := range c.Entries {
			copy := c
			copy.Entries = []string{e}
			updated = append(updated, copy)
		}
	}

	return updated
}
