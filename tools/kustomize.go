package main

import (
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func renderKustomize(path string) ([]byte, error) {
	fSys := filesys.MakeFsOnDisk()
	k := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	resMap, err := k.Run(fSys, path)
	if err != nil {
		return nil, err
	}

	return resMap.AsYaml()
}
