package generator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
)

type HttpGenerator struct {
	client http.Client
	config configuration.Configuration
	reader crd.CrdReader
}

func NewHttpGenerator(config configuration.Configuration, reader crd.CrdReader) Generator {
	return HttpGenerator{
		client: http.Client{
			Timeout: 15 * time.Second,
		},
		config: config,
		reader: reader,
	}
}

func (generator HttpGenerator) MetaData(version string) ([]crd.CrdMetaSchema, error) {
	schemas, err := generator.Schemas(version)
	if err != nil {
		return nil, err
	}
	metadata := make([]crd.CrdMetaSchema, len(schemas))
	for i, s := range schemas {
		metadata[i] = s.CrdMetaSchema
	}
	return metadata, nil
}

func (generator HttpGenerator) Schemas(version string) ([]crd.CrdSchema, error) {
	download, err := resolveDownload(version, generator.config.Downloads)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	for _, path := range download.Paths {
		request := fmt.Sprintf("%s/%s", download.BaseUri, path)
		resp, err := generator.client.Get(request)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			continue
		}

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		buf.Write(bytes)
		buf.WriteString("---\n")
	}

	crds, err := generator.reader.Read(bytes.NewReader(buf.Bytes()), fmt.Sprintf("buffered bytes for %s", generator.config.Name))
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

	return schemas, nil
}

func (generator HttpGenerator) Versions() ([]string, error) {
	versions := make([]string, len(generator.config.Downloads))

	for i, download := range generator.config.Downloads {
		versions[i] = download.Version
	}

	return versions, nil
}

func (generator HttpGenerator) Close() error {
	return nil
}

func resolveDownload(version string, downloads []configuration.ConfigurationDownload) (*configuration.ConfigurationDownload, error) {
	for _, download := range downloads {
		if download.Version == version {
			return &download, nil
		}
	}
	return nil, fmt.Errorf("version '%s' was not found", version)
}
