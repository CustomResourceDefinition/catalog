package generator

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/registry"
	"github.com/CustomResourceDefinition/catalog/internal/semver"
)

type Builder struct {
	generatedRepository  string
	schemaRepository     string
	definitionRepository string
	logger               io.Writer
	config               configuration.Configuration
	generator            Generator
	versionFilter        *regexp.Regexp
	registry             *registry.SourceRegistry
}

func NewBuilder(config configuration.Configuration, reader crd.CrdReader, generatedRepository, schemaRepository, definitionRepository string, logger io.Writer, reg *registry.SourceRegistry) (*Builder, error) {
	generator, err := resolveGenerator(config, reader, logger)
	if err != nil {
		return nil, err
	}

	if len(config.Namespace) == 0 {
		config.Namespace = "namespace"
	}

	pattern := defaultVersionPattern(config.Kind)
	if len(config.VersionPattern) > 0 {
		pattern = config.VersionPattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &Builder{
		config:               config,
		generator:            generator,
		versionFilter:        re,
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

	latestVersion, isUpdated := builder.registryStatus()
	if isUpdated {
		fmt.Fprintf(logger, " - skipping %s@%s (version %s unchanged)\n", builder.config.Name, builder.config.Kind, latestVersion)
		return nil
	}

	fmt.Fprintf(logger, " - checking version %s for completeness.\n", latestVersion)
	metadata, err := builder.generator.MetaData(latestVersion)
	if err != nil {
		fmt.Fprintf(logger, " ! failed: %s\n", err.Error())
		return err
	}

	versions := []string{}
	missing, known := verifyKnownMetadata(metadata, builder.schemaRepository)
	if known {
		fmt.Fprintf(logger, " - complete -> render only latest version.\n")
		versions = []string{latestVersion}
	} else {
		fmt.Fprintf(logger, " - missing %s -> render all versions.\n", missing)
		versions, err = builder.versions()
		if err != nil {
			return err
		}
		fmt.Fprintf(logger, " - found %d versions.\n", len(versions))
		slices.Reverse(versions)
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

func (builder Builder) versions() ([]string, error) {
	versions, err := builder.generator.Versions()
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0)
	for _, v := range versions {
		if builder.versionFilter.MatchString(v) {
			filtered = append(filtered, v)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		keyA, errA := builder.generator.VersionSortKey(filtered[i])
		keyB, errB := builder.generator.VersionSortKey(filtered[j])

		if errA == nil && errB == nil && keyA != 0 && keyB != 0 && keyA != keyB {
			return keyA > keyB
		}

		a := normalizeVersion(builder.versionFilter.FindAllStringSubmatch(filtered[i], -1))
		b := normalizeVersion(builder.versionFilter.FindAllStringSubmatch(filtered[j], -1))
		return semver.Compare(a, b) > 0
	})
	return filtered, nil
}

func (builder Builder) registryStatus() (string, bool) {
	versions, err := builder.versions()
	if err != nil || len(versions) == 0 {
		return "", false
	}

	version := versions[0]

	if builder.registry == nil {
		return version, false
	}

	if entry, ok := builder.registry.Get(builder.config.Name); ok {
		return version, entry.Kind == string(builder.config.Kind) && entry.Version == version
	}

	return version, false
}

func (builder Builder) updateRegistry(version string) {
	if builder.registry == nil {
		return
	}

	builder.registry.Set(builder.config.Name, string(builder.config.Kind), version)
}

func normalizeVersion(matches [][]string) string {
	if len(matches) == 0 || len(matches[0]) < 2 {
		return "v0.0.0"
	}

	version := matches[0][1]
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return "v0.0.0"
	}

	ints := make([]int, 3)
	for i := range 3 {
		n, _ := strconv.Atoi(parts[i])
		ints[i] = n
	}

	return fmt.Sprintf("v%d.%d.%d", ints[0], ints[1], ints[2])
}

func resolveGenerator(config configuration.Configuration, reader crd.CrdReader, logger io.Writer) (Generator, error) {
	switch config.Kind {
	case configuration.Git:
		return NewGitGeneratorFactory(config, reader, logger).Build()
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
