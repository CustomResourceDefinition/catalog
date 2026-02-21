package generator

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

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
	plainHttp   bool
}

const HELM_OCI_PLAIN_HTTP = "HELM_OCI_PLAIN_HTTP"

func NewOciGenerator(config configuration.Configuration, reader crd.CrdReader) Generator {
	plainHttp := false
	env, found := os.LookupEnv(HELM_OCI_PLAIN_HTTP)
	value, err := strconv.ParseBool(env)
	if found && err == nil && value {
		plainHttp = true
	}

	return &OciGenerator{
		realmClient: newRealmClient(plainHttp),
		config:      config,
		reader:      reader,
		plainHttp:   plainHttp,
	}
}

func (generator *OciGenerator) Close() error {
	return os.RemoveAll(generator.tmpDir)
}

func (generator *OciGenerator) VersionSortKey(version string) (int64, error) {
	return 0, nil
}

func (generator *OciGenerator) MetaData(version string) ([]crd.CrdMetaSchema, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	crds, err := generator.Crds(version)
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

	metadata := make([]crd.CrdMetaSchema, len(schemas))
	for i, s := range schemas {
		metadata[i] = s.CrdMetaSchema
	}
	return metadata, nil
}

func (generator *OciGenerator) Crds(version string) ([]crd.Crd, error) {
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

	return crds, nil
}

func (generator *OciGenerator) Versions() ([]string, error) {
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

	opts := []registry.ClientOption{
		registry.ClientOptDebug(false),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stdout),
	}

	if generator.plainHttp {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	regClient, err := registry.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	provider := getter.Provider{
		Schemes: []string{registry.OCIScheme},
		New: func(options ...getter.Option) (getter.Getter, error) {
			return getter.NewOCIGetter(getter.WithPlainHTTP(generator.plainHttp))
		},
	}

	chartDownloader := downloader.ChartDownloader{
		Getters:        []getter.Provider{provider},
		RegistryClient: regClient,
	}

	generator.downloader = chartDownloader
	generator.tmpDir = tmpDir

	return nil
}
