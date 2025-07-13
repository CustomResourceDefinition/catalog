package command

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

	"gopkg.in/yaml.v3"
)

type Comparer struct {
	Datreeio, Current, Out, Ignore string
	flags                          *flag.FlagSet
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

type ignore struct {
	Description string       `yaml:"description"`
	Items       []ignoreItem `yaml:"items"`
}

type ignoreItem struct {
	Group   string `yaml:"group"`
	Kind    string `yaml:"kind"`
	Version string `yaml:"version"`
}

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

func NewComparer(datreeio, current, output, ignore string, flags *flag.FlagSet) Comparer {
	return Comparer{
		flags:    flags,
		Current:  current,
		Datreeio: datreeio,
		Ignore:   ignore,
		Out:      output,
	}
}

func (cmd Comparer) Run() error {
	if err := cmd.validate(); err != nil {
		if cmd.flags != nil {
			cmd.flags.Usage()
		}
		return errors.Join(ErrInvalidConfiguration, err)
	}

	ignores := make([]ignore, 0)
	if len(cmd.Ignore) > 0 {
		f, err := os.ReadFile(cmd.Ignore)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(f, &ignores)
		if err != nil {
			return err
		}
	}

	data, err := find(cmd.Current, ignores, cmd.Datreeio, "datreeio/CRDs-catalog", "https://github.com/datreeio/CRDs-catalog")
	if err != nil {
		return err
	}

	return render([]markdownData{data}, cmd.Out)
}

func (cmd Comparer) validate() error {
	directories := []string{cmd.Current, cmd.Datreeio}
	for _, d := range directories {
		if f, err := os.Stat(d); err != nil || len(d) == 0 || !f.IsDir() {
			return fmt.Errorf("'%s' is not a valid directory path", d)
		}
	}
	if len(cmd.Ignore) > 0 {
		if f, err := os.Stat(cmd.Ignore); err != nil || f.IsDir() {
			return fmt.Errorf("'%s' is not a valid file path", cmd.Ignore)
		}
	}
	if len(cmd.Out) == 0 {
		return fmt.Errorf("no output path was configured")
	}
	return nil
}

func find(src string, ignores []ignore, datreeio string, name string, uri string) (markdownData, error) {
	data := markdownData{
		Name: name,
		Uri:  uri,
	}

	current, err := findItems(src)
	if err != nil {
		return data, err
	}

	target, err := findItems(datreeio)
	if err != nil {
		return data, err
	}

	data.update(current, target, ignores)

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
	f, err := os.Create(fmt.Sprintf("%s.tmp", out))
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
	return fmt.Sprintf("%s:%s:%s", i.group, i.kind, i.version)
}

func (data *markdownData) update(current []item, target []item, ignores []ignore) {
	if data.Data == nil {
		data.Data = make([]markdownItem, 0)
	}

	res := make([]*regexp.Regexp, 0)
	for _, ignore := range ignores {
		for _, i := range ignore.Items {
			re := regexp.MustCompile(fmt.Sprintf("%s:%s:%s", i.Group, i.Kind, i.Version))
			res = append(res, re)
		}
	}

	marker := make(map[string]bool, 0)
	for _, i := range current {
		marker[i.Id()] = false
	}

	missing := make([]item, 0)
	groups := make(map[string]string, 0)
	for _, item := range target {
		id := item.Id()
		_, ok := marker[id]
		for _, re := range res {
			if !ok && re.MatchString(id) {
				ok = true
			}
		}
		if !ok {
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

		for kind := range kinds {
			versions := make([]string, 0)
			groupItem := markdownItems{Kind: kind}

			for _, i := range missing {
				if i.group != group || i.kind != kind || slices.Contains(versions, i.version) {
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
}
