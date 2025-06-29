package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"strings"
)

type Updater struct {
	Input, Output string
	Logger        io.Writer
	flags         *flag.FlagSet
}

type crd struct {
	Group, Kind, Version string
	Bytes                []byte
}

func (u Updater) Run() error {
	if err := u.validate(); err != nil {
		if u.flags != nil {
			u.flags.Usage()
		}
		return errors.Join(errInvalidConfiguration, err)
	}

	if u.Logger == nil {
		u.Logger = os.Stderr
	}

	fmt.Fprintf(u.Logger, "Updating:\n")
	dir := u.Input
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		fmt.Fprintf(u.Logger, " - group: %s\n", e.Name())
		if e.IsDir() {
			err := handleGroup(path.Join(dir, e.Name()), u.Output, u.Logger)
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

func handleGroup(dir string, output string, logger io.Writer) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			next := path.Join(dir, e.Name())
			fmt.Fprintf(logger, "   - using application: %s\n", e.Name())
			err := handleApplication(next, output, logger)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func handleApplication(dir string, output string, logger io.Writer) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	collection := make(map[string]crd, 0)
	for _, e := range entries {
		item := path.Join(dir, e.Name())
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(item), ".yaml") || strings.HasSuffix(strings.ToLower(item), ".yml") {
			items, err := resolve(item, logger)
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

func resolve(path string, logger io.Writer) (map[string]crd, error) {
	list, err := decodeCRDs(path, logger)
	if err != nil {
		return nil, err
	}

	items := make(map[string]crd, 0)
	for _, c := range list {
		for _, v := range c.Spec.Versions {
			schema := v.Schema.OpenAPIV3Schema
			applyDefaults(schema, true)

			b, err := json.MarshalIndent(schema, "", "  ")
			if err != nil {
				return nil, err
			}
			b = append(b, []byte("\n")...)

			item := crd{
				Group:   strings.ToLower(c.Spec.Group),
				Kind:    strings.ToLower(c.Spec.Names.Kind),
				Version: strings.ToLower(v.Name),
				Bytes:   b,
			}

			file := fmt.Sprintf("%s/%s_%s.json", item.Group, item.Kind, item.Version)
			items[file] = item
		}
	}

	return items, nil
}
