package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"
)

type StatusGenerator struct {
	Remote, Current, Out string
	flags                *flag.FlagSet
}

type item struct {
	group, kind, version string
}

var errInvalidConfiguration = errors.New("invalid configuration")

func (g StatusGenerator) Run() error {
	if err := g.Validate(); err != nil {
		if g.flags != nil {
			g.flags.Usage()
		}
		return errors.Join(errInvalidConfiguration, err)
	}

	current, err := findItems(g.Current)
	if err != nil {
		return nil
	}

	remote, err := findItems(g.Remote)
	if err != nil {
		return nil
	}

	fqrf := make(map[string]bool, 0)
	klrf := make(map[string]bool, 0)
	tlrf := make(map[string]bool, 0)
	for _, i := range remote {
		fqrf[i.Id()] = false
		klrf[i.Kind()] = false
		tlrf[i.group] = false
	}
	for _, i := range current {
		fqrf[i.Id()] = true
		klrf[i.Kind()] = true
		tlrf[i.group] = true
	}

	// FIXME: write template report with numbers for FQ, kinds and groups
	// 		  update README too about looking for migration help

	return nil
}

func (g StatusGenerator) Validate() error {
	directories := []string{g.Current, g.Remote}
	for _, d := range directories {
		if _, err := os.Stat(d); err != nil || len(d) == 0 {
			return fmt.Errorf("'%s' is not a valid directory path", d)
		}
	}
	if len(g.Out) == 0 {
		return fmt.Errorf("no output path was configured")
	}
	return nil
}

func findItems(src string) ([]item, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return nil, err
	}

	items := make([]item, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			group, err := findGroupItems(path.Join(src, entry.Name()))
			if err != nil {
				return nil, err
			}
			items = append(items, group...)
		}
	}
	return items, nil
}

func findGroupItems(src string) ([]item, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`(.+)_v(.+)\.json`)

	items := make([]item, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			continue
		}

		items = append(items, item{
			group:   path.Base(src),
			kind:    matches[1],
			version: matches[2],
		})
	}
	return items, nil
}

func (i *item) Id() string {
	return fmt.Sprintf("%s/%s_%s.json", i.group, i.kind, i.version)
}

func (i *item) Kind() string {
	return fmt.Sprintf("%s/%s", i.group, i.kind)
}
