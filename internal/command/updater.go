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

	"github.com/CustomResourceDefinition/catalog/internal/generator"
)

type Updater struct {
	Input, Output string
	Logger        io.Writer
	flags         *flag.FlagSet
	reader        generator.CrdReader
}

func NewUpdater(input, output string, logger io.Writer, flags *flag.FlagSet) Updater {
	return Updater{
		flags:  flags,
		Input:  input,
		Output: output,
		Logger: logger,
	}
}

func (u Updater) Run() error {
	if err := u.validate(); err != nil {
		if u.flags != nil {
			u.flags.Usage()
		}
		return errors.Join(ErrInvalidConfiguration, err)
	}

	if u.Logger == nil {
		u.Logger = os.Stderr
	}

	reader, err := generator.NewCrdReader(u.Logger)
	if err != nil {
		return err
	}
	u.reader = reader

	fmt.Fprintf(u.Logger, "Updating:\n")
	dir := u.Input
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		fmt.Fprintf(u.Logger, " - group: %s\n", e.Name())
		if e.IsDir() {
			err := u.handleGroup(path.Join(dir, e.Name()), u.Output, u.Logger)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (u Updater) validate() error {
	directories := []string{u.Input, u.Output}
	for _, d := range directories {
		if f, err := os.Stat(d); err != nil || len(d) == 0 || !f.IsDir() {
			return fmt.Errorf("'%s' is not a valid directory path", d)
		}
	}
	return nil
}

func (u Updater) handleGroup(dir string, output string, logger io.Writer) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			next := path.Join(dir, e.Name())
			fmt.Fprintf(logger, "   - using application: %s\n", e.Name())
			err := u.handleApplication(next, output, logger)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (u Updater) handleApplication(dir string, output string, logger io.Writer) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	collection := make(map[string]generator.CrdSchema, 0)
	for _, e := range entries {
		item := path.Join(dir, e.Name())
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(item), ".yaml") || strings.HasSuffix(strings.ToLower(item), ".yml") {
			items, err := u.resolve(item)
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

func (u Updater) resolve(path string) (map[string]generator.CrdSchema, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	list, err := u.reader.Read(file, path)
	if err != nil {
		return nil, err
	}

	items := make(map[string]generator.CrdSchema, 0)
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
