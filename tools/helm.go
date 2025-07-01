package main

import (
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

func renderHelmChart(chartPath string, releaseName string, namespace string) (string, error) {
	settings := cli.New()
	actionConfig := new(action.Configuration)

	err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {})
	if err != nil {
		return "", err
	}

	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.ReleaseName = releaseName
	client.Replace = true
	client.ClientOnly = true
	client.Namespace = namespace

	chart, err := loader.Load(chartPath)
	if err != nil {
		return "", err
	}

	release, err := client.Run(chart, nil)
	if err != nil {
		return "", err
	}

	return release.Manifest, nil
}
