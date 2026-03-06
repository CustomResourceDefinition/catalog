package generator

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"runtime"
	"slices"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/registry"
)

type Builder struct {
	generatedRepository  string
	schemaRepository     string
	definitionRepository string
	logger               io.Writer
	config               configuration.Configuration
	generator            Generator
	registry             *registry.SourceRegistry
}

func NewBuilder(
	config configuration.Configuration,
	reader crd.CrdReader,
	generatedRepository, schemaRepository, definitionRepository string,
	logger io.Writer,
	reg *registry.SourceRegistry,
) (*Builder, error) {
	generator, err := resolveGenerator(config, reader, logger)
	if err != nil {
		return nil, err
	}

	return &Builder{
		config:               config,
		generator:            generator,
		logger:               logger,
		schemaRepository:     schemaRepository,
		generatedRepository:  generatedRepository,
		definitionRepository: definitionRepository,
		registry:             reg,
	}, nil
}

func defaultVersionPattern(kind configuration.Kind) string {
	switch kind {
	case configuration.Helm, configuration.HelmOci:
		return `^v?([0-9]+\.[0-9]+\.[0-9]+)$`
	default:
		return `^([0-9]+\.[0-9]+\.[0-9]+)$`
	}
}

func (builder Builder) Build() error {
	defer builder.generator.Close()

	logger := builder.logger

	fmt.Fprintf(logger, "Producing for %s@%s:\n", builder.config.Name, builder.config.Kind)
	defer fmt.Fprintf(logger, "End.\n")

	latestVersion, isUpdated, err := builder.registryStatus()
	if err != nil {
		return err
	}
	if isUpdated {
		fmt.Fprintf(logger, " - skipping %s@%s (version %s unchanged)\n", builder.config.Name, builder.config.Kind, latestVersion)
		return nil
	}

	versions, err := builder.generator.Versions()
	if err != nil {
		return err
	}
	fmt.Fprintf(logger, " - found %d versions.\n", len(versions))
	slices.Reverse(versions)

	fmt.Fprintf(logger, " - checking version %s for completeness.\n", latestVersion)
	metadata, err := builder.generator.MetaData(latestVersion)
	if err != nil {
		fmt.Fprintf(logger, " ! failed: %s\n", err.Error())
		return err
	}

	missing, known := verifyKnownMetadata(metadata, builder.schemaRepository)
	if known {
		fmt.Fprintf(logger, " - complete -> render only latest version.\n")
		versions = []string{latestVersion}
	} else {
		fmt.Fprintf(logger, " - missing %s -> render all versions.\n", missing)
	}

	for _, version := range versions {
		runtime.GC()

		fmt.Fprintf(logger, " - render version %s.\n", version)
		crds, err := builder.generator.Crds(version)
		if err != nil {
			fmt.Fprintf(logger, " - - discarded due to error: %s.\n", err.Error())
			continue
		}

		schemas := make([]crd.CrdSchema, 0)
		valid := true
		for _, c := range crds {
			schema, err := c.Schema()
			if err != nil {
				fmt.Fprintf(logger, " - - discarding due to error: %s.\n", err.Error())
				valid = false
			}
			schemas = append(schemas, schema...)
		}
		if !valid {
			fmt.Fprintf(logger, " - - discarded.\n")
			continue
		}

		fmt.Fprintf(logger, " - - rendered %d definitions.\n", len(crds))
		for _, crd := range crds {
			file := path.Join(builder.definitionRepository, crd.Filepath())
			os.MkdirAll(path.Dir(file), 0755)
			err := os.WriteFile(file, crd.Bytes, 0644)
			if err != nil {
				return err
			}
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

	builder.updateRegistry(latestVersion)

	return nil
}

// registryStatus reports on the state in registry based on the latest version
// available and only uses the decided interface method for latest version information
func (builder Builder) registryStatus() (string, bool, error) {
	version, err := builder.generator.LatestVersion()
	if err != nil {
		return "", false, fmt.Errorf("unable to check registry: %w", err)
	}

	if builder.registry == nil {
		return version, false, nil
	}

	if entry, ok := builder.registry.Get(builder.config.Name); ok {
		return version, entry.Kind == string(builder.config.Kind) && entry.Version == version, nil
	}

	return version, false, nil
}

func (builder Builder) updateRegistry(version string) {
	if builder.registry == nil {
		return
	}

	builder.registry.Set(builder.config.Name, string(builder.config.Kind), version)
}

func resolveGenerator(config configuration.Configuration, reader crd.CrdReader, logger io.Writer) (Generator, error) {
	if len(config.Namespace) == 0 {
		config.Namespace = "namespace"
	}

	pattern := defaultVersionPattern(config.Kind)
	if len(config.VersionPattern) > 0 {
		pattern = config.VersionPattern
	}

	filter, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	switch config.Kind {
	case configuration.Git:
		return NewGitGeneratorFactory(config, reader, filter, logger).Build()
	case configuration.Http:
		return NewHttpGenerator(config, reader, filter), nil
	case configuration.Helm:
		target := config.Entries[len(config.Entries)-1]
		return NewHelmGenerator(target, config, reader, filter), nil
	case configuration.HelmOci:
		return NewOciGenerator(config, reader, filter), nil
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
