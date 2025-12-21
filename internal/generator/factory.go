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

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
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
}

func NewBuilder(config configuration.Configuration, reader crd.CrdReader, generatedRepository, schemaRepository, definitionRepository string, logger io.Writer) (*Builder, error) {
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

	pattern := fmt.Sprintf(`^%s([0-9]{1,})\.([0-9]{1,})\.([0-9]{1,})%s`, config.VersionPrefix, config.VersionSuffix)
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
	}, nil
}

func (builder Builder) Build() error {
	defer builder.generator.Close()

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
		a := normalizeVersion(builder.versionFilter.FindAllStringSubmatch(filtered[i], -1))
		b := normalizeVersion(builder.versionFilter.FindAllStringSubmatch(filtered[j], -1))
		return semver.Compare(a, b) > 0
	})
	return filtered, nil
}

func normalizeVersion(matches [][]string) string {
	ints := make([]int, 3)
	for i := 1; i <= 3; i++ {
		n, _ := strconv.Atoi(matches[0][i])
		ints[i-1] = n
	}

	return fmt.Sprintf("v%d.%d.%d", ints[0], ints[1], ints[2])
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
