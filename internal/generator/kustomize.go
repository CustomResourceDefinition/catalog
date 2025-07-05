package generator

import (
	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type KustomizeGenerator struct{}

func (generator KustomizeGenerator) NeededVersions(config configuration.Configuration, schemaRepository string) ([]string, error) {
	return nil, nil
}

func (generator KustomizeGenerator) Read(config configuration.Configuration, version string) ([]byte, error) {
	return nil, nil
}

func renderKustomize(path string) ([]byte, error) {
	fSys := filesys.MakeFsOnDisk()
	k := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	resMap, err := k.Run(fSys, path)
	if err != nil {
		return nil, err
	}

	return resMap.AsYaml()
}
