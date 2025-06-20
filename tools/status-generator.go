package main

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"text/template"
)

type StatusGenerator struct {
	Remote, Current, Out string
	flags                *flag.FlagSet
}

type item struct {
	group, kind, version string
}

type markdownData struct {
	Name, Uri string
	Data      []markdownItem
}

type markdownItem struct {
	Group string
	Items []markdownItems
}

type markdownItems struct {
	Kind, Versions string
}

var errInvalidConfiguration = errors.New("invalid configuration")

const markdown = `
# Comparison

This page lists missing CRD validation schemas that are present in alternative catalogs. If anyone knows the source of any of these CRDs, please make a pull request with that source or create an issue explaining where or what the source is.

{{- range . }}

## [{{ .Name }}]({{ .Uri }})

{{- range .Data }}

| {{ .Group }} | |
| --- | --- |
{{- range .Items }}
| {{ .Kind }} | {{ .Versions }} |
{{- end }}
{{- end }}

{{- end }}
`

func (g StatusGenerator) Run() error {
	if err := g.Validate(); err != nil {
		if g.flags != nil {
			g.flags.Usage()
		}
		return errors.Join(errInvalidConfiguration, err)
	}

	data, err := find(g.Current, g.Remote, "datreeio/CRDs-catalog", "https://github.com/datreeio/CRDs-catalog")
	if err != nil {
		return err
	}

	return render([]markdownData{data}, g.Out)
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

func find(src string, remote string, name string, uri string) (markdownData, error) {
	data := markdownData{
		Name: name,
		Uri:  uri,
		Data: make([]markdownItem, 0),
	}

	current, err := findItems(src)
	if err != nil {
		return data, err
	}

	target, err := findItems(remote)
	if err != nil {
		return data, err
	}

	marker := make(map[string]bool, 0)
	for _, i := range current {
		marker[i.Id()] = false
	}

	missing := make([]item, 0)
	groups := make(map[string]string, 0)
	for _, item := range target {
		if _, ok := marker[item.Id()]; !ok {
			missing = append(missing, item)
			groups[item.group] = item.group
		}
	}

	for group := range groups {
		item := markdownItem{Group: group, Items: []markdownItems{}}

		kinds := make(map[string]string, 0)
		for _, i := range missing {
			if i.group != group {
				continue
			}
			kinds[i.kind] = i.kind
		}

		versions := make([]string, 0)
		for kind := range kinds {
			groupItem := markdownItems{Kind: kind}

			for _, i := range missing {
				if i.group != group || i.kind != kind {
					continue
				}
				versions = append(versions, i.version)
			}

			slices.Sort(versions)

			groupItem.Versions = strings.Join(versions, ", ")
			item.Items = append(item.Items, groupItem)
		}

		slices.SortFunc(item.Items, func(a, b markdownItems) int { return cmp.Compare(a.Kind, b.Kind) })

		data.Data = append(data.Data, item)
	}

	slices.SortFunc(data.Data, func(a, b markdownItem) int { return cmp.Compare(a.Group, b.Group) })

	return data, nil
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

	re := regexp.MustCompile(`(.+)_(v.+)\.json`)

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

func render(list []markdownData, out string) error {
	f, err := os.CreateTemp("", "*.md")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	generator, err := template.New("markdown").Parse(markdown)
	if err != nil {
		return err
	}

	err = generator.Execute(f, list)
	if err != nil {
		return err
	}

	return os.Rename(f.Name(), out)
}

func (i *item) Id() string {
	return fmt.Sprintf("%s/%s_%s.json", i.group, i.kind, i.version)
}
