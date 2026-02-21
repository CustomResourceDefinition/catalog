package generator

import (
	"io"

	"github.com/CustomResourceDefinition/catalog/internal/crd"
)

type Generator interface {
	Versions() ([]string, error)
	MetaData(version string) ([]crd.CrdMetaSchema, error)
	Crds(version string) ([]crd.Crd, error)
	VersionSortKey(version string) (int64, error)
	io.Closer
}
