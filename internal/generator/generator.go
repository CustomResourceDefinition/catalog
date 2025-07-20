package generator

import (
	"io"

	"github.com/CustomResourceDefinition/catalog/internal/crd"
)

type Generator interface {
	Versions() ([]string, error)
	MetaData(version string) ([]crd.CrdMetaSchema, error)
	Schemas(version string) ([]crd.CrdSchema, error)
	io.Closer
}
