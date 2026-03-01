package registry

import (
	"errors"
	"io/fs"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type SourceRegistry struct {
	Sources map[string]SourceEntry `yaml:"sources"`
}

type SourceEntry struct {
	Kind        string `yaml:"kind"`
	Version     string `yaml:"version"`
	LastUpdated string `yaml:"lastUpdated"`
}

func Load(path string) (*SourceRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &SourceRegistry{Sources: make(map[string]SourceEntry)}, nil
		}
		return nil, err
	}

	var reg SourceRegistry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, err
	}

	if reg.Sources == nil {
		reg.Sources = make(map[string]SourceEntry)
	}

	return &reg, nil
}

func (r *SourceRegistry) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(r)
}

func (r *SourceRegistry) Get(name string) (SourceEntry, bool) {
	entry, ok := r.Sources[name]
	return entry, ok
}

func (r *SourceRegistry) Set(name, kind, version string) {
	r.Sources[name] = SourceEntry{
		Kind:        kind,
		Version:     version,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}
}
