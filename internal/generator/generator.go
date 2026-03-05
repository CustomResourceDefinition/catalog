package generator

import (
	"io"
	"regexp"

	"github.com/CustomResourceDefinition/catalog/internal/crd"
)

type Generator interface {
	Versions(filter *regexp.Regexp) ([]string, error)
	LatestVersion(filter *regexp.Regexp) (string, error)
	MetaData(version string) ([]crd.CrdMetaSchema, error)
	Crds(version string) ([]crd.Crd, error)
	io.Closer
}
