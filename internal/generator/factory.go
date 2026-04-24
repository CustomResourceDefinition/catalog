package generator

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"runtime"
	"slices"
	"time"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/registry"
	"github.com/CustomResourceDefinition/catalog/internal/timing"
)

type Builder struct {
	generatedRepository  string
	schemaRepository     string
	definitionRepository string
	logger               io.Writer
	config               configuration.Configuration
	generator            Generator
	registry             *registry.SourceRegistry
	stats                *timing.Stats
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
		stats:                timing.NewStats(),
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

	if _, ok := builder.generator.(*PreparedGitGenerator); ok {
		fmt.Fprintf(logger, " - using prepared git generator\n")
	}

	start := time.Now()
	latestVersion, isUpdated, entry, err := builder.registryStatus()
	builder.stats.Record(timing.CategoryMisc, timing.OperationTypeStatus, "registry_status", time.Since(start), err == nil, start)
	if err != nil {
		return err
	}
	if isUpdated {
		fmt.Fprintf(logger, " - skipping %s@%s (version %s unchanged)\n", builder.config.Name, builder.config.Kind, latestVersion)
		return nil
	}

	versions, err := builder.fetchVersions(logger)
	if err != nil {
		return err
	}

	metadata, err := builder.fetchMetadata(logger, latestVersion)
	if err != nil {
		return err
	}

	savedPastRefs := make([]registry.CrdRef, 0)
	missing, known := verifyKnownMetadata(metadata, builder.schemaRepository)
	if !known {
		fmt.Fprintf(logger, " - missing %s -> render all versions.\n", missing)
	} else {
		if entry == nil {
			fmt.Fprintf(logger, " - complete -> render only latest version.\n")
			versions = []string{latestVersion}
		} else if len(entry.Refs) == 0 {
			fmt.Fprintf(logger, " - missing reference metadata -> render all versions.\n")
		} else {
			fmt.Fprintf(logger, " - complete -> render only latest version.\n")
			savedPastRefs = entry.PastRefs
			versions = []string{latestVersion}
		}
	}

	refs, pastRefs := builder.render(versions, savedPastRefs, logger)

	start = time.Now()
	builder.updateRegistry(latestVersion, refs, pastRefs)
	builder.stats.Record(timing.CategoryMisc, timing.OperationTypeUpdate, "update_registry", time.Since(start), true, start)

	return nil
}

func (builder Builder) render(versions []string, pastRefs []registry.CrdRef, logger io.Writer) ([]registry.CrdRef, []registry.CrdRef) {
	currentRefs := make([]registry.CrdRef, 0)
	previousRefs := pastRefs
	for _, version := range versions {
		runtime.GC()

		refs, err := builder.renderVersion(logger, version)
		if err != nil {
			continue
		}

		currentRefs = refs
		pastRefs = computeRemoved(pastRefs, previousRefs, refs)
		previousRefs = refs
	}
	return currentRefs, pastRefs
}

func (builder Builder) fetchVersions(logger io.Writer) ([]string, error) {
	start := time.Now()
	versions, err := builder.generator.Versions()
	builder.stats.Record(builder.operationCategory(), timing.OperationTypeAPIFetch, "versions", time.Since(start), err == nil, start)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(logger, " - found %d versions.\n", len(versions))
	slices.Reverse(versions)
	return versions, nil
}

func (builder Builder) fetchMetadata(logger io.Writer, latestVersion string) ([]crd.CrdMetaSchema, error) {
	fmt.Fprintf(logger, " - checking version %s for completeness.\n", latestVersion)
	start := time.Now()
	metadata, err := builder.generator.MetaData(latestVersion)
	builder.stats.Record(builder.operationCategory(), timing.OperationTypeAPIFetch, "metadata", time.Since(start), err == nil, start)
	if err != nil {
		fmt.Fprintf(logger, " ! failed: %s\n", err.Error())
		return nil, err
	}
	return metadata, nil
}

func (builder Builder) renderVersion(logger io.Writer, version string) ([]registry.CrdRef, error) {
	fmt.Fprintf(logger, " - render version %s.\n", version)

	crds, err := builder.generateCrds(version)
	if err != nil {
		fmt.Fprintf(logger, " - - discarded due to error: %s.\n", err.Error())
		return nil, err
	}

	schemas := builder.extractSchemas(logger, crds)

	if len(crds) > 0 {
		if err := builder.writeDefinitions(crds); err != nil {
			return nil, err
		}
	}

	if len(schemas) > 0 {
		if err := builder.writeSchemas(schemas); err != nil {
			return nil, err
		}
	}

	return builder.extractCrdRefs(crds), nil
}

func (builder Builder) extractCrdRefs(crds []crd.Crd) []registry.CrdRef {
	refs := make([]registry.CrdRef, 0)
	for _, c := range crds {
		meta, err := c.MetaSchema()
		if err != nil {
			continue
		}
		for _, m := range meta {
			refs = append(refs, registry.CrdRef{
				Group:   m.Group,
				Kind:    m.Kind,
				Version: m.Version,
			})
		}
	}
	return refs
}

func (builder Builder) generateCrds(version string) ([]crd.Crd, error) {
	start := time.Now()
	crds, err := builder.generator.Crds(version)
	builder.stats.Record(builder.operationCategory(), timing.OperationTypeGenerate, fmt.Sprintf("crds_%s", version), time.Since(start), err == nil, start)
	return crds, err
}

func (builder Builder) extractSchemas(logger io.Writer, crds []crd.Crd) []crd.CrdSchema {
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
		return nil
	}
	return schemas
}

func (builder Builder) writeDefinitions(crds []crd.Crd) error {
	fmt.Fprintf(builder.logger, " - - rendered %d definitions.\n", len(crds))
	startTime := time.Now()
	for _, crd := range crds {
		file := path.Join(builder.definitionRepository, crd.Filepath())
		os.MkdirAll(path.Dir(file), 0755)
		if err := os.WriteFile(file, crd.Bytes, 0644); err != nil {
			return err
		}
	}
	duration := time.Since(startTime)
	if len(crds) > 0 {
		builder.stats.Record(timing.CategoryGeneration, timing.OperationTypeWrite, "definitions", duration, true, startTime)
	}
	return nil
}

func (builder Builder) writeSchemas(schemas []crd.CrdSchema) error {
	fmt.Fprintf(builder.logger, " - - rendered %d schema.\n", len(schemas))
	startTime := time.Now()
	for _, schema := range schemas {
		file := path.Join(builder.generatedRepository, schema.Filepath())
		os.MkdirAll(path.Dir(file), 0755)
		if err := os.WriteFile(file, schema.Bytes, 0644); err != nil {
			return err
		}
	}
	duration := time.Since(startTime)
	if len(schemas) > 0 {
		builder.stats.Record(timing.CategoryGeneration, timing.OperationTypeWrite, "schemas", duration, true, startTime)
	}
	return nil
}

func (builder Builder) Stats() *timing.Stats {
	return builder.stats
}

func (builder Builder) operationCategory() timing.Category {
	switch builder.config.Kind {
	case configuration.Git:
		return timing.CategoryGit
	case configuration.Http:
		return timing.CategoryHTTP
	case configuration.Helm:
		return timing.CategoryHelm
	case configuration.HelmOci:
		return timing.CategoryOCI
	}
	return timing.CategoryMisc
}

// registryStatus reports on the state in registry based on the latest version
// available and only uses the decided interface method for latest version information
func (builder Builder) registryStatus() (string, bool, *registry.SourceEntry, error) {
	version, err := builder.generator.LatestVersion()
	if err != nil {
		return "", false, nil, fmt.Errorf("unable to check registry: %w", err)
	}

	if builder.registry == nil {
		return version, false, nil, nil
	}

	if entry, ok := builder.registry.Get(builder.config.Name); ok {
		return version,
			entry.Kind == string(builder.config.Kind) && entry.Version == version,
			&entry,
			nil
	}

	return version, false, &registry.SourceEntry{}, nil
}

func (builder Builder) updateRegistry(version string, refs, pastRefs []registry.CrdRef) {
	if builder.registry == nil {
		return
	}
	sortRefs(refs)
	sortRefs(pastRefs)
	builder.registry.Set(builder.config.Name, string(builder.config.Kind), version, refs, pastRefs)
}

func computeRemoved(pastCrds, previousCrds, currentCrds []registry.CrdRef) []registry.CrdRef {
	everProduced := make(map[registry.CrdRef]bool)
	for _, c := range pastCrds {
		everProduced[c] = true
	}
	for _, c := range previousCrds {
		everProduced[c] = true
	}
	for _, c := range currentCrds {
		everProduced[c] = true
	}

	currentSet := make(map[registry.CrdRef]bool)
	for _, c := range currentCrds {
		currentSet[c] = true
	}

	removed := make([]registry.CrdRef, 0)
	for crd := range everProduced {
		if !currentSet[crd] {
			removed = append(removed, crd)
		}
	}

	return removed
}

func sortRefs(refs []registry.CrdRef) {
	slices.SortFunc(refs, func(a, b registry.CrdRef) int {
		if a.Group != b.Group {
			if a.Group < b.Group {
				return -1
			}
			return 1
		}
		if a.Kind != b.Kind {
			if a.Kind < b.Kind {
				return -1
			}
			return 1
		}
		if a.Version < b.Version {
			return -1
		}
		if a.Version > b.Version {
			return 1
		}
		return 0
	})
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
