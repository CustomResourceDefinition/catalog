package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type Converter struct {
	Input, Output string
	flags         *flag.FlagSet
}

func (c Converter) Run() error {
	if err := c.validate(); err != nil {
		if c.flags != nil {
			c.flags.Usage()
		}
		return errors.Join(errInvalidConfiguration, err)
	}

	crd, err := decodeCRD(c.Input)
	if err != nil {
		return err
	}

	for _, v := range crd.Spec.Versions {
		schema := v.Schema.OpenAPIV3Schema
		applyDefaults(schema, true)

		b, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return err
		}
		b = append(b, []byte("\n")...)

		filename := path.Join(c.Output, strings.ToLower(fmt.Sprintf("%s_%s.json", crd.Spec.Names.Kind, v.Name)))
		err = os.WriteFile(filename, b, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Converter) validate() error {
	directories := []string{c.Output}
	for _, d := range directories {
		if f, err := os.Stat(d); err != nil || len(d) == 0 || !f.IsDir() {
			return fmt.Errorf("'%s' is not a valid directory path", d)
		}
	}
	if f, err := os.Stat(c.Input); err != nil || f.IsDir() {
		return fmt.Errorf("'%s' is not a valid file path", c.Input)
	}
	return nil
}

func decodeCRD(file string) (*v1.CustomResourceDefinition, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	codecs := serializer.NewCodecFactory(scheme)
	decoder := codecs.UniversalDeserializer()

	obj, _, err := decoder.Decode(b, nil, nil)
	if err != nil {
		return nil, err
	}

	crd, ok := obj.(*v1.CustomResourceDefinition)
	if !ok {
		return nil, errors.New("unable to convert to CRD")
	}

	return crd, nil
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
