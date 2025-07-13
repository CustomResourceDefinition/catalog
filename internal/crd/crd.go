package crd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type CrdReader interface {
	Read(reader io.Reader, file string) ([]Crd, error)
}

type crdReader struct {
	decoder runtime.Decoder
	logger  io.Writer
	matcher *regexp.Regexp
}

type Crd struct {
	Group, Kind string
	definition  *v1.CustomResourceDefinition
}

type CrdSchema struct {
	CrdMetaSchema
	Bytes []byte
}

type CrdMetaSchema struct {
	Group, Kind, Version string
}

func NewCrdReader(logger io.Writer) (CrdReader, error) {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	codecs := serializer.NewCodecFactory(scheme)
	decoder := codecs.UniversalDeserializer()

	matcher, err := regexp.Compile(`[^\t\n\v\f\r ]+`)
	if err != nil {
		return nil, err
	}

	return &crdReader{decoder: decoder, logger: logger, matcher: matcher}, nil
}

func (r *crdReader) Read(reader io.Reader, file string) ([]Crd, error) {
	yr := yaml.NewYAMLReader(bufio.NewReader(reader))

	list := make([]Crd, 0)
	index := -1
	for {
		index++
		doc, err := yr.Read()
		if err == io.EOF {
			break
		}
		if !r.matcher.Match(doc) {
			continue
		}

		obj, _, err := r.decoder.Decode(doc, nil, nil)
		if err != nil {
			if !runtime.IsMissingKind(err) && !runtime.IsNotRegisteredError(err) {
				fmt.Fprintf(r.logger, "   unable to decode document #%d at %s, err: %s\n", index, file, strings.ReplaceAll(err.Error(), "\n", "\n   "))
			}
			continue
		}

		crd, ok := obj.(*v1.CustomResourceDefinition)
		if !ok {
			fmt.Fprintf(r.logger, "   invalid document #%d at %s\n", index, file)
			continue
		}

		c := Crd{
			Group:      strings.ToLower(crd.Spec.Group),
			Kind:       strings.ToLower(crd.Spec.Names.Kind),
			definition: crd,
		}

		list = append(list, c)
	}

	return list, nil
}

func (c *Crd) MetaSchema() ([]CrdMetaSchema, error) {
	list := make([]CrdMetaSchema, 0)

	for _, v := range c.definition.Spec.Versions {
		item := CrdMetaSchema{
			Group:   strings.ToLower(c.definition.Spec.Group),
			Kind:    strings.ToLower(c.definition.Spec.Names.Kind),
			Version: strings.ToLower(v.Name),
		}

		list = append(list, item)
	}

	return list, nil
}

func (c *Crd) Schema() ([]CrdSchema, error) {
	list := make([]CrdSchema, 0)

	for _, v := range c.definition.Spec.Versions {
		schema := v.Schema.OpenAPIV3Schema
		applyDefaults(schema, true)

		b, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return nil, err
		}
		b = append(b, []byte("\n")...)

		item := CrdSchema{
			CrdMetaSchema: CrdMetaSchema{
				Group:   strings.ToLower(c.definition.Spec.Group),
				Kind:    strings.ToLower(c.definition.Spec.Names.Kind),
				Version: strings.ToLower(v.Name),
			},
			Bytes: b,
		}

		list = append(list, item)
	}

	return list, nil
}

func (s *CrdMetaSchema) Filepath() string {
	return fmt.Sprintf("%s/%s_%s.json", s.Group, s.Kind, s.Version)
}

func applyDefaults(schema *v1.JSONSchemaProps, skip bool) {
	if !skip && len(schema.Properties) > 0 && schema.AdditionalProperties == nil {
		schema.AdditionalProperties = &v1.JSONSchemaPropsOrBool{Allows: false}
	}

	schema.Description = ""
	if schema.Format == "int-or-string" {
		schema.Type = ""
		schema.Format = ""
		schema.OneOf = []v1.JSONSchemaProps{
			{Type: "string"},
			{Type: "integer"},
		}
	}

	if schema.Items != nil {
		if schema.Items.Schema != nil {
			applyDefaults(schema.Items.Schema, false)
		}
		for i := range schema.Items.JSONSchemas {
			applyDefaults(&schema.Items.JSONSchemas[i], false)
		}
	}
	if schema.AllOf != nil {
		for i := range schema.AllOf {
			applyDefaults(&schema.AllOf[i], false)
		}
	}
	if schema.AnyOf != nil {
		for i := range schema.AnyOf {
			applyDefaults(&schema.AnyOf[i], false)
		}
	}
	if schema.OneOf != nil {
		for i := range schema.OneOf {
			applyDefaults(&schema.OneOf[i], false)
		}
	}
	if schema.Not != nil {
		applyDefaults(schema.Not, false)
	}
	for k, v := range schema.Properties {
		applyDefaults(&v, false)
		schema.Properties[k] = v
	}
}
