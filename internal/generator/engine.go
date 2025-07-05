package generator

import (
	"fmt"
	"io"
	"os"
	"path"
	"sort"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"golang.org/x/mod/semver"
)

type Factory struct {
	schemaRepository string
	crdReader        crd.CrdReader
	logger           io.Writer
}

type Build struct {
	Factory
	config    configuration.Configuration
	generator Generator
}

func NewFactory(schemaRepository string, logger io.Writer) (*Factory, error) {
	reader, err := crd.NewCrdReader(logger)
	if err != nil {
		return nil, err
	}
	return &Factory{
		schemaRepository: schemaRepository,
		crdReader:        reader,
		logger:           logger,
	}, nil
}

func (f Factory) NewBuild(config configuration.Configuration) (*Build, error) {
	generator, err := resolveGenerator(config, f.crdReader)
	if err != nil {
		return nil, err
	}
	return &Build{
		Factory:   f,
		config:    config,
		generator: generator,
	}, nil
}

func (build Build) Complete() error {
	logger := build.Factory.logger

	fmt.Fprintf(logger, "Producing for %s:\n", build.config.Name)
	defer fmt.Fprintf(logger, "End.\n")

	versions, err := build.versions()
	if err != nil {
		return err
	}
	count := len(versions)
	fmt.Fprintf(logger, " - found %d versions.\n", count)

	if count <= 0 {
		return fmt.Errorf("empty list of versions for '%s'", build.config.Name)
	}
	version := versions[0]

	fmt.Fprintf(logger, " - checking version %s for completeness.\n", version)
	metadata, err := build.generator.MetaData(version)
	if err != nil {
		return err
	}

	missing, known := verifyKnownMetadata(metadata, build.schemaRepository)
	if known {
		fmt.Fprintf(logger, " - complete -> render only latest version.\n")
		versions = []string{version}
	} else {
		fmt.Fprintf(logger, " - missing %s -> render all versions.\n", missing)
	}

	for _, version := range versions {
		fmt.Fprintf(logger, " - render version %s.\n", version)
		schemas, err := build.generator.Schemas(version)
		if err != nil {
			fmt.Fprintf(logger, " - - discarded due to error: %s.\n", err.Error())
			continue
		}

		fmt.Fprintf(logger, " - - rendered %d schema.\n", len(schemas))
		for _, schema := range schemas {
			file := path.Join(build.schemaRepository, schema.Filepath())
			os.MkdirAll(path.Dir(file), 0755)
			err := os.WriteFile(file, schema.Bytes, 0644)
			if err != nil {
				return err
			}
		}
	}

	build.generator.Close()

	return nil
}

func (build Build) versions() ([]string, error) {
	versions, err := build.generator.Versions()
	if err != nil {
		return nil, err
	}
	sort.Slice(versions, func(i, j int) bool {
		return semver.Compare(versions[i], versions[j]) > 0
	})
	return versions, nil
}

func resolveGenerator(config configuration.Configuration, reader crd.CrdReader) (Generator, error) {
	switch config.Kind {
	// FIXME: no no
	// case configuration.Git:
	// 	return GitGenerator{}, nil
	case configuration.Http:
		return NewHttpGenerator(config, reader), nil
	case configuration.Helm:
		target := config.Entries[len(config.Entries)-1] // FIXME: do config splitting
		return NewHelmGenerator(target, config, reader), nil
	// case configuration.HelmOci:
	default:
		return nil, fmt.Errorf("no generators matched for kind '%s'", config.Kind)
	}
}

func verifyKnownMetadata(metadata []crd.CrdMetaSchema, schemaRepository string) (string, bool) {
	for _, m := range metadata {
		filename := m.Filepath()
		fullPath := path.Join(schemaRepository, filename)
		if f, err := os.Stat(fullPath); err != nil || !f.IsDir() {
			return filename, false
		}
	}
	return "", true
}
