package main

import (
	_ "embed"
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

func UnmarshalConfigurations(reader io.Reader) (*[]Configuration, error) {
	sch := jsonschema.MustCompileString("schema.json", schema)

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var dict []any
	var conf []Configuration
	err = yaml.Unmarshal(bytes, &dict)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		return nil, err
	}

	err = sch.Validate(dict)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
