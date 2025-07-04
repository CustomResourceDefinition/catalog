package configuration

import (
	_ "embed"
	"errors"
	"fmt"
	"io"

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
	AdditionalVersions []string                `yaml:"additionalVersions"`
	ApiGroups          []string                `yaml:"apiGroups"`
	Downloads          []ConfigurationDownload `yaml:"crds"`
	Entries            []string                `yaml:"entries"`
	GenPaths           []string                `yaml:"genPaths"`
	IncludeHead        bool                    `yaml:"includeHead"`
	Kind               Kind                    `yaml:"kind"`
	KustomizePaths     []string                `yaml:"kustomizationPaths"`
	Name               string                  `yaml:"name"`
	Repository         string                  `yaml:"repository"`
	SearchPaths        []string                `yaml:"searchPaths"`
	Values             []ConfigurationValues   `yaml:"valuesFiles"`
	VersionPrefix      string                  `yaml:"versionPrefix"`
	VersionSuffix      string                  `yaml:"versionSuffix"`
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
