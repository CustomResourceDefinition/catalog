package command

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

type Verifier struct {
	Schema, File string
	Logger       io.Writer
	flags        *flag.FlagSet
}

func NewVerifier(schema, file string, flags *flag.FlagSet) Verifier {
	return Verifier{
		Schema: schema,
		File:   file,
		flags:  flags,
	}
}

func (cmd Verifier) Run() error {
	if err := cmd.validate(); err != nil {
		if cmd.flags != nil {
			cmd.flags.Usage()
		}
		return errors.Join(ErrInvalidConfiguration, err)
	}

	schema, err := readFile(cmd.Schema)
	if err != nil {
		return err
	}

	uri := path.Base(cmd.Schema)
	sch := jsonschema.MustCompileString(uri, string(schema))

	bytes, err := readFile(cmd.File)
	if err != nil {
		return err
	}

	obj, err := unpack(cmd.File, bytes)
	if err != nil {
		return err
	}

	err = sch.Validate(obj)
	if err != nil {
		return err
	}

	return nil
}

func (cmd Verifier) validate() error {
	files := []string{cmd.Schema, cmd.File}
	for _, file := range files {
		if f, err := os.Stat(file); err != nil || len(file) == 0 || f.IsDir() {
			return fmt.Errorf("'%s' is not a valid file path", file)
		}
	}
	return nil
}

func readFile(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readBytes(f)
}

func readBytes(reader io.Reader) ([]byte, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func unpack(file string, data []byte) (any, error) {
	var out any
	var list []any

	if tryUnpack(data, &list) {
		return list, nil
	}
	if tryUnpack(data, &out) {
		if _, ok := out.(string); !ok {
			return out, nil
		}
	}

	return nil, fmt.Errorf("unable to unpack '%s' as either JSON or YAML", file)
}

func tryUnpack(data []byte, target any) bool {
	return yaml.Unmarshal(data, target) == nil || json.Unmarshal(data, target) == nil
}
