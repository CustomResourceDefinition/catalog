package generator

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/semver"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

type HelmGenerator struct {
	target     string
	config     configuration.Configuration
	reader     crd.CrdReader
	tmpDir     string
	versions   repo.ChartVersions
	downloader downloader.ChartDownloader
}

func NewHelmGenerator(target string, config configuration.Configuration, reader crd.CrdReader) Generator {
	return &HelmGenerator{
		target: target,
		config: config,
		reader: reader,
	}
}

func (generator *HelmGenerator) MetaData(version string) ([]crd.CrdMetaSchema, error) {
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

func (generator *HelmGenerator) Crds(version string) ([]crd.Crd, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	if version == "" {
		version = ">0.0.0"
	}

	ref := fmt.Sprintf("%s/%s", generator.config.Name, generator.target)
	filename, _, err := generator.downloader.DownloadTo(ref, version, generator.tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to download chart: %w", err)
	}

	values, err := generator.config.ValuesFile(version)
	if err != nil {
		return nil, err
	}

	rendered, err := renderChart(filename, "release", generator.config.Namespace, values)
	if err != nil {
		return nil, err
	}

	crds, err := generator.reader.Read(bytes.NewReader(rendered), fmt.Sprintf("buffered bytes for %s", generator.config.Name))
	if err != nil {
		return nil, err
	}

	return crds, nil
}

func (generator *HelmGenerator) Versions() ([]string, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	versions := make([]string, 0)

	for _, version := range generator.versions {
		if !version.Removed && semver.IsCanonical(version.Version) {
			versions = append(versions, version.Version)
		}
	}

	return versions, nil
}

func (generator *HelmGenerator) Close() error {
	return os.Remove(generator.tmpDir)
}

func (generator *HelmGenerator) ensureLoaded() error {
	if len(generator.tmpDir) != 0 {
		return nil
	}

	tmpDir, err := os.MkdirTemp("", generator.config.Name)
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	settings := cli.EnvSettings{}
	getters := getter.All(&settings)
	entry := repo.Entry{
		Name: generator.config.Name,
		URL:  generator.config.Repository,
	}

	repoFilePath := path.Join(tmpDir, "repositories.yaml")
	repoFile := repo.NewFile()
	repoFile.Add(&entry)
	repoFile.WriteFile(repoFilePath, 0644)

	chartRepo, err := repo.NewChartRepository(&entry, getters)
	if err != nil {
		return err
	}

	chartRepo.CachePath = tmpDir
	indexFile, err := chartRepo.DownloadIndexFile()
	if err != nil {
		return err
	}

	index, err := repo.LoadIndexFile(indexFile)
	if err != nil {
		return err
	}

	versions, ok := index.Entries[generator.target]
	if !ok {
		return fmt.Errorf("target '%s' was not found in index file", generator.target)
	}

	chartDownloader := downloader.ChartDownloader{
		Getters:          getters,
		RepositoryConfig: repoFilePath,
		RepositoryCache:  tmpDir,
	}

	generator.downloader = chartDownloader
	generator.versions = versions
	generator.tmpDir = tmpDir

	return nil
}

func renderChart(chartPath, releaseName, namespace string, values map[string]any) ([]byte, error) {
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	if chart.Metadata == nil {
		return nil, fmt.Errorf("chart metadata is missing")
	}

	options := chartutil.ReleaseOptions{
		Name:      releaseName,
		Namespace: namespace,
		Revision:  1,
		IsInstall: true,
	}

	vals, err := chartutil.ToRenderValues(chart, values, options, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to render values: %w", err)
	}

	rendered, err := engine.Render(chart, vals)
	if err != nil {
		return nil, fmt.Errorf("failed to render templates: %w", err)
	}

	buf := Buffer{}
	for filename, content := range rendered {
		if len(strings.TrimSpace(content)) == 0 {
			continue
		}
		if strings.HasSuffix(filename, "NOTES.txt") {
			continue
		}
		buf.WriteString(content)
		buf.WriteString("---")
	}

	for _, obj := range chart.CRDObjects() {
		buf.Write(obj.File.Data)
		buf.WriteString("---")
	}

	return buf.Bytes(), nil
}
