package generator

import (
	"bytes"
	"fmt"
	"os"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
)

type OciGenerator struct {
	realmClient realmClient
	config      configuration.Configuration
	reader      crd.CrdReader
	tmpDir      string
	downloader  downloader.ChartDownloader
}

const HELM_OCI_PLAIN_HTTP = "HELM_OCI_PLAIN_HTTP"

func NewOciGenerator(config configuration.Configuration, reader crd.CrdReader) Generator {
	return OciGenerator{
		realmClient: newRealmClient(),
		config:      config,
		reader:      reader,
	}
}

func (generator OciGenerator) Close() error {
	return os.Remove(generator.tmpDir)
}

func (generator OciGenerator) MetaData(version string) ([]crd.CrdMetaSchema, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	schemas, err := generator.Schemas(version)
	if err != nil {
		return nil, err
	}
	metadata := make([]crd.CrdMetaSchema, len(schemas))
	for i, s := range schemas {
		metadata[i] = s.CrdMetaSchema
	}
	return metadata, nil
}

func (generator OciGenerator) Schemas(version string) ([]crd.CrdSchema, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	if len(version) == 0 {
		versions, err := generator.Versions()
		if err != nil {
			return nil, err
		}
		version = versions[0]
	}

	savedPath, _, err := generator.downloader.DownloadTo(generator.config.Repository, version, generator.tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to download chart: %w", err)
	}

	values, err := generator.config.ValuesFile(version)
	if err != nil {
		return nil, err
	}

	rendered, err := renderChart(savedPath, "release", "namespace", values)
	if err != nil {
		return nil, err
	}

	crds, err := generator.reader.Read(bytes.NewReader(rendered), fmt.Sprintf("buffered bytes for %s", generator.config.Name))
	if err != nil {
		return nil, err
	}

	schemas := make([]crd.CrdSchema, 0)
	for _, c := range crds {
		schema, err := c.Schema()
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema...)
	}

	return schemas, nil
}

func (generator OciGenerator) Versions() ([]string, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	tags, err := generator.realmClient.ListOciTags(generator.config.Repository)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (generator *OciGenerator) ensureLoaded() error {
	if len(generator.tmpDir) != 0 {
		return nil
	}

	tmpDir, err := os.MkdirTemp("", generator.config.Name)
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	regClient, err := registry.NewClient(
		registry.ClientOptDebug(false),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stdout),
		registry.ClientOptPlainHTTP(), // FIXME: follow up
	)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	provider := getter.Provider{
		Schemes: []string{registry.OCIScheme},
		New:     newOCIGetter,
	}

	chartDownloader := downloader.ChartDownloader{
		Getters:        []getter.Provider{provider},
		RegistryClient: regClient,
	}

	generator.downloader = chartDownloader
	generator.tmpDir = tmpDir

	return nil
}

func newOCIGetter(options ...getter.Option) (getter.Getter, error) {
	return getter.NewOCIGetter(getter.WithPlainHTTP(true)) // FIXME: follow up
}
