package crd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	encoder "gopkg.in/yaml.v3"
	api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type CrdReader interface {
	Read(reader io.Reader, file string) ([]Crd, error)
}

type crdReader struct {
	decoder runtime.Decoder
	scheme  *runtime.Scheme
	logger  io.Writer
	matcher *regexp.Regexp
}

type Crd struct {
	Group, Kind string
	Bytes       []byte
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
	if err := v1beta1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := v1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := api.AddToScheme(scheme); err != nil {
		return nil, err
	}

	scheme.AddConversionFunc((*v1beta1.CustomResourceDefinition)(nil), (*api.CustomResourceColumnDefinition)(nil), func(in interface{}, out interface{}, s conversion.Scope) error {
		return v1beta1.Convert_v1beta1_CustomResourceColumnDefinition_To_apiextensions_CustomResourceColumnDefinition(in.(*v1beta1.CustomResourceColumnDefinition), out.(*api.CustomResourceColumnDefinition), s)
	})
	scheme.AddConversionFunc(
		(*api.CustomResourceDefinition)(nil),
		(*v1.CustomResourceDefinition)(nil),
		func(in interface{}, out interface{}, s conversion.Scope) error {
			return v1.Convert_apiextensions_CustomResourceDefinition_To_v1_CustomResourceDefinition(
				in.(*api.CustomResourceDefinition), out.(*v1.CustomResourceDefinition), s)
		},
	)

	codecs := serializer.NewCodecFactory(scheme)
	decoder := codecs.UniversalDeserializer()

	matcher, err := regexp.Compile(`[^\t\n\v\f\r ]+`)
	if err != nil {
		return nil, err
	}

	return &crdReader{decoder: decoder, logger: logger, matcher: matcher, scheme: scheme}, nil
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

		crd := r.convertToV1(obj)
		if crd == nil {
			fmt.Fprintf(r.logger, "   invalid document #%d at %s\n", index, file)
			continue
		}

		var data map[string]any
		if err := json.Unmarshal(doc, &data); err == nil {
			out, err := encoder.Marshal(data)
			if err != nil {
				fmt.Fprintf(r.logger, "   failed to convert #%d at %s to yaml\n", index, file)
				continue
			}
			doc = out
		}

		if crd.Spec.Group == "" {
			fmt.Fprintf(r.logger, "   empty group declared for %s at %s\n", crd.ObjectMeta.Name, file)
			continue
		}
		if crd.Spec.Names.Kind == "" {
			fmt.Fprintf(r.logger, "   empty group declared for %s at %s\n", crd.ObjectMeta.Name, file)
			continue
		}

		c := Crd{
			Group:      strings.ToLower(crd.Spec.Group),
			Kind:       strings.ToLower(crd.Spec.Names.Kind),
			Bytes:      doc,
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
		if v.Schema == nil || v.Schema.OpenAPIV3Schema == nil {
			continue
		}
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

func (c *Crd) Filepath() string {
	return fmt.Sprintf("%s/%s.yaml", c.Group, c.Kind)
}

func (s *CrdMetaSchema) Filepath() string {
	return fmt.Sprintf("%s/%s_%s.json", s.Group, s.Kind, s.Version)
}

func applyDefaults(schema *v1.JSONSchemaProps, skip bool) {
	if !skip && len(schema.Properties) > 0 && schema.AdditionalProperties == nil {
		schema.AdditionalProperties = &v1.JSONSchemaPropsOrBool{Allows: false}
	}

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

func (r crdReader) convertToV1(obj runtime.Object) *v1.CustomResourceDefinition {
	var (
		crd v1.CustomResourceDefinition
		err error
	)

	if beta, ok := obj.(*v1beta1.CustomResourceDefinition); ok {
		var extension api.CustomResourceDefinition
		err = r.scheme.Convert(beta, &extension, nil)
		if err == nil {
			err = r.scheme.Convert(&extension, &crd, nil)
			if err == nil {
				return &crd
			}
		}
	}

	if c, ok := obj.(*v1.CustomResourceDefinition); ok {
		return c
	}

	return nil
}
