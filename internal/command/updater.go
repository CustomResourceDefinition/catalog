package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"strings"

	"github.com/CustomResourceDefinition/catalog/internal/crd"
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

	if cmd.Logger == nil {
		cmd.Logger = os.Stderr
	}

	reader, err := crd.NewCrdReader(cmd.Logger)
	if err != nil {
		return err
	}
	cmd.reader = reader

	fmt.Fprintf(cmd.Logger, "Updating:\n")
	dir := cmd.Configuration
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		fmt.Fprintf(cmd.Logger, " - group: %s\n", e.Name())
		if e.IsDir() {
			err := cmd.handleGroup(path.Join(dir, e.Name()), cmd.Output, cmd.Logger)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (cmd Updater) validate() error {
	directories := []string{cmd.Configuration, cmd.Output}
	for _, d := range directories {
		if f, err := os.Stat(d); err != nil || len(d) == 0 || !f.IsDir() {
			return fmt.Errorf("'%s' is not a valid directory path", d)
		}
	}
	return nil
}

func (cmd Updater) handleGroup(dir string, output string, logger io.Writer) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			next := path.Join(dir, e.Name())
			fmt.Fprintf(logger, "   - using application: %s\n", e.Name())
			err := cmd.handleApplication(next, output, logger)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (cmd Updater) handleApplication(dir string, output string, logger io.Writer) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	collection := make(map[string]crd.CrdSchema, 0)
	for _, e := range entries {
		item := path.Join(dir, e.Name())
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(item), ".yaml") || strings.HasSuffix(strings.ToLower(item), ".yml") {
			items, err := cmd.resolve(item)
			if err != nil {
				return err
			}
			maps.Copy(collection, items)
		} else {
			fmt.Fprintf(logger, "     - skipping: %s\n", e.Name())
		}
	}

	for filename, crd := range collection {
		file := path.Join(output, filename)
		os.MkdirAll(path.Dir(file), 0755)
		err := os.WriteFile(file, crd.Bytes, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cmd Updater) resolve(path string) (map[string]crd.CrdSchema, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	list, err := cmd.reader.Read(file, path)
	if err != nil {
		return nil, err
	}

	items := make(map[string]crd.CrdSchema, 0)
	for _, item := range list {
		schema, err := item.Schema()
		if err != nil {
			return nil, err
		}

		for _, s := range schema {
			file := fmt.Sprintf("%s/%s_%s.json", s.Group, s.Kind, s.Version)
			items[file] = s
		}
	}

	return items, nil
}
