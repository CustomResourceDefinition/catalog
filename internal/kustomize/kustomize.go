package kustomize

import (
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func Render(path string) ([]byte, error) {
	fs := filesys.MakeFsOnDisk()
	k := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	mapping, err := k.Run(fs, path)
	if err != nil {
		return nil, err
	}

	return mapping.AsYaml()
}
