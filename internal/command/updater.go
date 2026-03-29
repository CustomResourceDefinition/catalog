package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/generator"
	"github.com/CustomResourceDefinition/catalog/internal/registry"
	"github.com/CustomResourceDefinition/catalog/internal/timing"
)

type teeWriter struct {
	Writers []io.Writer
}

func (t teeWriter) Write(p []byte) (n int, err error) {
	for _, w := range t.Writers {
		if _, err := w.Write(p); err != nil {
			return n, err
		}
		n = len(p)
	}
	return n, nil
}

func (cmd Updater) createSummaryWriter() (io.Writer, func(), error) {
	summaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryPath == "" {
		return cmd.Logger, func() {}, nil
	}

	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	return teeWriter{Writers: []io.Writer{cmd.Logger, f}}, func() { f.Close() }, nil
}

type Updater struct {
	Configuration, Schema, Definitions string
	Logger                             io.Writer
	flags                              *flag.FlagSet
	reader                             crd.CrdReader
	registry                           *registry.SourceRegistry
	registryPath                       string
	performanceLog                     string
}

func NewUpdater(configuration, schema, definitions, registryPath, performancePath string, logger io.Writer, flags *flag.FlagSet) Updater {
	return Updater{
		flags:          flags,
		Configuration:  configuration,
		Schema:         schema,
		Definitions:    definitions,
		registryPath:   registryPath,
		Logger:         logger,
		performanceLog: performancePath,
	}
}

func (cmd Updater) Run() error {
	if err := cmd.validate(); err != nil {
		if cmd.flags != nil {
			cmd.flags.Usage()
		}
		return errors.Join(ErrInvalidConfiguration, err)
	}

	if err := cmd.initialize(); err != nil {
		return err
	}

	configurations, err := readConfiguration(cmd.Configuration)
	if err != nil {
		return err
	}

	if cmd.registry != nil {
		validKeys := validSourceKeys(configurations)
		for key := range cmd.registry.Sources {
			if !validKeys[key] {
				delete(cmd.registry.Sources, key)
				fmt.Fprintf(cmd.Logger, "Removing stale registry entry: %s\n", key)
			}
		}
	}

	tmpDir, err := os.MkdirTemp("", "output")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	totalStats := timing.NewStats()

	if err := totalStats.OpenLogFile(cmd.performanceLog); err != nil {
		return fmt.Errorf("failed to open performance log: %w", err)
	}
	defer totalStats.CloseLogFile()

	for _, config := range splitConfigurations(configurations) {
		runtime.GC()

		build, err := generator.NewBuilder(config, cmd.reader, tmpDir, cmd.Schema, cmd.Definitions, cmd.Logger, cmd.registry)
		if err != nil {
			fmt.Fprintf(cmd.Logger, "::warning:: unable to create builder for %s: %v\n", config.Name, err)
			continue
		}

		err = build.Build()
		if err != nil {
			fmt.Fprintf(cmd.Logger, "::warning:: build of %s failed: %v\n", config.Name, err)
			continue
		}

		stats := build.Stats()
		for _, op := range stats.GetAllStats() {
			totalStats.Record(op.Category, op.Type, op.Name, op.Duration, op.Success, op.StartTime)
		}
	}

	if cmd.registry != nil && cmd.registryPath != "" {
		if err := cmd.registry.Save(cmd.registryPath); err != nil {
			fmt.Fprintf(cmd.Logger, "::warning:: failed to save registry: %v\n", err)
		}
	}

	writer, closer, err := cmd.createSummaryWriter()
	if err != nil {
		return err
	}
	defer closer()
	totalStats.PrintSummary(writer)

	return merge(tmpDir, cmd.Schema)
}

func (cmd Updater) validate() error {
	if f, err := os.Stat(cmd.Schema); err != nil || len(cmd.Schema) == 0 || !f.IsDir() {
		return fmt.Errorf("'%s' is not a valid directory path", cmd.Schema)
	}

	if f, err := os.Stat(cmd.Definitions); err != nil || len(cmd.Definitions) == 0 || !f.IsDir() {
		return fmt.Errorf("'%s' is not a valid directory path", cmd.Definitions)
	}

	if f, err := os.Stat(cmd.Configuration); err != nil || len(cmd.Configuration) == 0 || f.IsDir() {
		return fmt.Errorf("'%s' is not a valid file path", cmd.Configuration)
	}

	return nil
}

func (cmd *Updater) initialize() error {
	if cmd.Logger == nil {
		cmd.Logger = os.Stderr
	}

	reader, err := crd.NewCrdReader(cmd.Logger)
	if err != nil {
		return err
	}
	cmd.reader = reader

	if cmd.registryPath != "" {
		reg, err := registry.Load(cmd.registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
		cmd.registry = reg
	}

	return nil
}

func readConfiguration(file string) ([]configuration.Configuration, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	configs, err := configuration.UnmarshalConfigurations(f)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// splitConfigurations will duplicate all entries in Entries, so each configuration only has a single entry in Entries
func splitConfigurations(configurations []configuration.Configuration) []configuration.Configuration {
	updated := make([]configuration.Configuration, 0)
	for _, c := range configurations {
		if len(c.Entries) == 0 {
			updated = append(updated, c)
			continue
		}
		for _, e := range c.Entries {
			cfg := c
			cfg.Name = fmt.Sprintf("%s.%s", c.Name, e)
			cfg.Entries = []string{e}
			updated = append(updated, cfg)
		}
	}

	return updated
}

func validSourceKeys(configs []configuration.Configuration) map[string]bool {
	valid := make(map[string]bool)
	for _, cfg := range splitConfigurations(configs) {
		valid[cfg.Name] = true
	}
	return valid
}

// merge will move created files in generated into the schema repository,
// and will attempt to write a file to schema repository if moving it fails
func merge(generatedRepository, schemaRepository string) error {
	groups, err := os.ReadDir(generatedRepository)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if group.IsDir() {
			_ = os.Mkdir(path.Join(schemaRepository, group.Name()), 0755)
			files, err := os.ReadDir(path.Join(generatedRepository, group.Name()))
			if err != nil {
				return err
			}

			for _, f := range files {
				if f.Type().IsRegular() {
					origin := path.Join(generatedRepository, group.Name(), f.Name())
					destination := path.Join(schemaRepository, group.Name(), f.Name())
					err = os.Rename(origin, destination)
					if err != nil {
						bytes, err := os.ReadFile(origin)
						if err != nil {
							return err
						}
						err = os.WriteFile(destination, bytes, 0644)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
