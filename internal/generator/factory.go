package generator

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"slices"
	"sort"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/semver"
)

type Builder struct {
	generatedRepository string
	schemaRepository    string
	logger              io.Writer
	config              configuration.Configuration
	generator           Generator
	versionFilter       *regexp.Regexp
}

func NewBuilder(config configuration.Configuration, reader crd.CrdReader, generatedRepository, schemaRepository string, logger io.Writer) (*Builder, error) {
	generator, err := resolveGenerator(config, reader)
	if err != nil {
		return nil, err
	}

	if len(config.VersionPrefix) == 0 {
		if config.Kind == configuration.Helm || config.Kind == configuration.HelmOci {
			config.VersionPrefix = "v?"
		} else {
			config.VersionPrefix = ""
		}
	}

	if len(config.VersionSuffix) == 0 {
		config.VersionSuffix = "$"
	}

	if len(config.Namespace) == 0 {
		config.Namespace = "namespace"
	}

	pattern := fmt.Sprintf(`^%s([0-9]{1,}\.[0-9]{1,}\.[0-9]{1,})%s`, config.VersionPrefix, config.VersionSuffix)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &Builder{
		config:              config,
		generator:           generator,
		versionFilter:       re,
		logger:              logger,
		schemaRepository:    schemaRepository,
		generatedRepository: generatedRepository,
	}, nil
}

func (builder Builder) Build() error {
	logger := builder.logger

	fmt.Fprintf(logger, "Producing for %s@%s:\n", builder.config.Name, builder.config.Kind)
	defer fmt.Fprintf(logger, "End.\n")

	versions, err := builder.versions()
	if err != nil {
		return err
	}
	count := len(versions)
	fmt.Fprintf(logger, " - found %d versions.\n", count)

	if count <= 0 {
		return fmt.Errorf("empty list of versions for '%s'", builder.config.Name)
	}
	version := versions[0]

	fmt.Fprintf(logger, " - checking version %s for completeness.\n", version)
	metadata, err := builder.generator.MetaData(version)
	if err != nil {
		fmt.Fprintf(logger, " ! failed: %s\n", err.Error())
		return err
	}

	missing, known := verifyKnownMetadata(metadata, builder.schemaRepository)
	if known {
		fmt.Fprintf(logger, " - complete -> render only latest version.\n")
		versions = []string{version}
	} else {
		fmt.Fprintf(logger, " - missing %s -> render all versions.\n", missing)
		slices.Reverse(versions)
	}

	for _, version := range versions {
		fmt.Fprintf(logger, " - render version %s.\n", version)
		schemas, err := builder.generator.Schemas(version)
		if err != nil {
			fmt.Fprintf(logger, " - - discarded due to error: %s.\n", err.Error())
			continue
		}

		fmt.Fprintf(logger, " - - rendered %d schema.\n", len(schemas))
		for _, schema := range schemas {
			file := path.Join(builder.generatedRepository, schema.Filepath())
			os.MkdirAll(path.Dir(file), 0755)
			err := os.WriteFile(file, schema.Bytes, 0644)
			if err != nil {
				return err
			}
		}
	}

	builder.generator.Close()

	return nil
}

func (builder Builder) versions() ([]string, error) {
	versions, err := builder.generator.Versions()
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0)
	for _, v := range versions {
		if builder.versionFilter.MatchString(v) || v == referenceHead {
			filtered = append(filtered, v)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i] == referenceHead {
			return true
		}
		if filtered[j] == referenceHead {
			return false
		}

		return semver.Compare(builder.versionFilter.FindAllStringSubmatch(filtered[i], 1)[0][1], builder.versionFilter.FindAllStringSubmatch(filtered[j], 1)[0][1]) > 0
	})
	return filtered, nil
}

func resolveGenerator(config configuration.Configuration, reader crd.CrdReader) (Generator, error) {
	switch config.Kind {
	case configuration.Git:
		return NewGitGenerator(config, reader), nil
	case configuration.Http:
		return NewHttpGenerator(config, reader), nil
	case configuration.Helm:
		target := config.Entries[len(config.Entries)-1]
		return NewHelmGenerator(target, config, reader), nil
	case configuration.HelmOci:
		return NewOciGenerator(config, reader), nil
	default:
		return nil, fmt.Errorf("no generators matched for kind '%s'", config.Kind)
	}
}

func verifyKnownMetadata(metadata []crd.CrdMetaSchema, schemaRepository string) (string, bool) {
	for _, m := range metadata {
		filename := m.Filepath()
		fullPath := path.Join(schemaRepository, filename)
		if f, err := os.Stat(fullPath); err != nil || f.IsDir() {
			return filename, false
		}
	}
	return "", true
}
