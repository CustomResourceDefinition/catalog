package configuration

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/CustomResourceDefinition/catalog/internal/semver"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

type Kind string

const (
	Git     Kind = "git"
	Http    Kind = "http"
	Helm    Kind = "helm"
	HelmOci Kind = "helm-oci"
)

type Configuration struct {
	ApiGroups      []string                `yaml:"apiGroups"`
	Downloads      []ConfigurationDownload `yaml:"crds"`
	Entries        []string                `yaml:"entries"`
	GenPaths       []string                `yaml:"genPaths"`
	Kind           Kind                    `yaml:"kind"`
	KustomizePaths []string                `yaml:"kustomizationPaths"`
	Name           string                  `yaml:"name"`
	Namespace      string                  `yaml:"namespace"`
	Repository     string                  `yaml:"repository"`
	SearchPaths    []string                `yaml:"searchPaths"`
	Values         []ConfigurationValues   `yaml:"valuesFiles"`
	VersionPattern string                  `yaml:"versionPattern"`
}

type ConfigurationDownload struct {
	BaseUri string   `yaml:"baseUri"`
	Paths   []string `yaml:"paths"`
	Version string   `yaml:"version"`
}

type ConfigurationValues struct {
	ValuesFile string `yaml:"valuesFile"`
	Version    string `yaml:"version"`
}

//go:embed schema.json
var schema string

// UnmarshalConfigurations validates and unpacks bytes from the reader as a list of configurations.
//
// The jsonschema used for validating configurations is embedded.
func UnmarshalConfigurations(reader io.Reader) ([]Configuration, error) {
	if reader == nil {
		return nil, fmt.Errorf("nothing to unmarshal, reader was nil")
	}

	const uri = "schema.json"
	sch := jsonschema.MustCompileString(uri, schema)

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var dict []any
	var conf []Configuration
	err = errors.Join(yaml.Unmarshal(bytes, &dict), yaml.Unmarshal(bytes, &conf))
	if err != nil {
		return nil, err
	}

	err = sch.Validate(dict)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// ValuesFile resolves the most relevant values file content from the configuration,
// note that versions are treated as semver, but any prefix/suffix extras or leading zeroes will result in improper sorting
func (c *Configuration) ValuesFile(version string) (map[string]any, error) {
	sort.Slice(c.Values, func(i, j int) bool {
		return semver.Compare(c.Values[i].Version, c.Values[j].Version) > 0
	})

	for _, v := range c.Values {
		if semver.Compare(v.Version, version) > 0 {
			continue
		}

		var result map[string]any
		err := yaml.Unmarshal([]byte(v.ValuesFile), &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	return nil, nil
}
