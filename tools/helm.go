package main

import (
	"fmt"
	"os"
	"path"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
)

func pullOCIChart(ociRef string, version string) (string, error) {
	settings := cli.EnvSettings{}

	regClient, err := registry.NewClient(
		registry.ClientOptDebug(false),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stdout),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create registry client: %w", err)
	}

	// Set up the ChartDownloader with OCI support
	chartDownloader := downloader.ChartDownloader{
		Getters:        getter.All(&settings),
		RegistryClient: regClient,
	}

	// Destination to save the chart
	tmpDir, err := os.MkdirTemp("", "helm-oci-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Pull the chart into the temp dir
	savedPath, _, err := chartDownloader.DownloadTo(ociRef, version, tmpDir)
	if err != nil {
		return "", fmt.Errorf("failed to download chart: %w", err)
	}

	// Returns full path to .tgz file
	return savedPath, nil
}

// pullHelmChart downloads a Helm chart from an HTTP repo and returns the path to the .tgz file.
func pullHelmChart(repoURL, version, name, chart string) (string, error) {
	// Create a temporary directory to store the downloaded chart
	tmpDir, err := os.MkdirTemp("", "helm-pull-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	settings := cli.EnvSettings{}
	getters := getter.All(&settings)
	entry := repo.Entry{
		Name: name,
		URL:  repoURL,
	}

	repoFilePath := path.Join(tmpDir, "rps.yaml")
	repoFile := repo.NewFile()
	repoFile.Add(&entry)
	repoFile.WriteFile(repoFilePath, 0644)

	chartRepo, err := repo.NewChartRepository(&entry, getters)
	if err != nil {
		return "", err
	}

	chartRepo.CachePath = tmpDir
	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return "", err
	}
	// tmp + repoName + -index.yaml
	// LoadIndexFile
	// SortEntries

	chartDownloader := downloader.ChartDownloader{
		Getters:          getters,
		RepositoryConfig: repoFilePath,
		RepositoryCache:  tmpDir,
	}

	// Pull chart
	destPath := tmpDir
	if version == "" {
		version = ">0.0.0" // fetch latest if not provided
	}

	ref := fmt.Sprintf("%s/%s", name, chart)
	filename, _, err := chartDownloader.DownloadTo(ref, version, destPath)
	if err != nil {
		return "", fmt.Errorf("failed to download chart: %w", err)
	}

	// Validate it loads
	_, err = loader.Load(filename)
	if err != nil {
		return "", fmt.Errorf("downloaded chart is invalid: %w", err)
	}

	return filename, nil
}

func renderChart(chartPath, releaseName, namespace string, values map[string]interface{}) (string, error) {
	// Load chart from .tgz or directory
	chart, err := loader.Load(chartPath)
	if err != nil {
		return "", fmt.Errorf("failed to load chart: %w", err)
	}

	if chart.Metadata == nil {
		return "", fmt.Errorf("chart metadata is missing")
	}

	// Prepare chart values context
	options := chartutil.ReleaseOptions{
		Name:      releaseName,
		Namespace: namespace,
		Revision:  1,
		IsInstall: true,
	}

	// Render values.yaml + --set values
	vals, err := chartutil.ToRenderValues(chart, values, options, nil)
	if err != nil {
		return "", fmt.Errorf("failed to render values: %w", err)
	}

	// Use the engine to render templates
	rendered, err := engine.Render(chart, vals)
	if err != nil {
		return "", fmt.Errorf("failed to render templates: %w", err)
	}

	// Flatten into a single manifest string (like helm template output)
	final := ""
	for _, content := range rendered {
		final += content + "\n---\n"
	}

	return final, nil
}
