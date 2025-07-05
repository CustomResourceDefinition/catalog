package generator

import (
	"bytes"
	"io"
	"os"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdMarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

type GenAllGenerator struct{}

func (generator GenAllGenerator) NeededVersions(config configuration.Configuration, schemaRepository string) ([]string, error) {
	return nil, nil
}

func (generator GenAllGenerator) Read(config configuration.Configuration, version string) ([]byte, error) {
	return nil, nil
}

// GenAllCrd mimics the following command
// `controller-gen crd "paths=/input" output:crd:dir=/output output:none`
// and returns bytes for generated CRD yaml documents, if possible.
func GenAllCrd(input string) []byte {
	buf := bytes.Buffer{}

	oldWD, _ := os.Getwd()
	defer os.Chdir(oldWD)
	err := os.Chdir(input)
	if err != nil {
		return buf.Bytes()
	}

	roots, _ := loader.LoadRoots(".")
	reg := &markers.Registry{}
	crdMarkers.Register(reg)

	out := &outputRule{
		buf: &buf,
	}

	ctx := &genall.GenerationContext{
		Collector:  &markers.Collector{Registry: reg},
		Roots:      roots,
		Checker:    &loader.TypeChecker{},
		OutputRule: out,
	}

	length := 0
	gen := &crd.Generator{
		CRDVersions: []string{"v1"},
		MaxDescLen:  &length,
	}

	gen.Generate(ctx)

	return buf.Bytes()
}

// outputRule is wrapping a byte buffer up so we can use it as an OutputRule
type outputRule struct {
	buf *bytes.Buffer
}

func (o *outputRule) Open(_ *loader.Package, _ string) (io.WriteCloser, error) {
	return nopCloser{o.buf}, nil
}

type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}
