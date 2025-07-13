package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/genall"
	"github.com/CustomResourceDefinition/catalog/internal/kustomize"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

type GitGenerator struct {
	config configuration.Configuration
	repo   *git.Repository
	reader crd.CrdReader
	head   plumbing.Hash
	tmpDir string
}

const referenceHead = "head"

func NewGitGenerator(config configuration.Configuration, reader crd.CrdReader) Generator {
	return GitGenerator{
		config: config,
		reader: reader,
	}
}

func (generator GitGenerator) Close() error {
	return os.Remove(generator.tmpDir)
}

func (generator GitGenerator) MetaData(version string) ([]crd.CrdMetaSchema, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

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

func (generator GitGenerator) Schemas(version string) ([]crd.CrdSchema, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	tree, err := generator.repo.Worktree()
	if err != nil {
		return nil, err
	}

	if len(version) == 0 {
		versions, err := generator.Versions()
		if err != nil {
			return nil, err
		}
		version = versions[0]
	}

	var opts git.CheckoutOptions
	if version == referenceHead {
		opts = git.CheckoutOptions{
			Hash: generator.head,
		}
	} else {
		opts = git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(version),
		}
	}
	err = tree.Checkout(&opts)
	if err != nil {
		return nil, err
	}

	buf := Buffer{}
	if len(generator.config.SearchPaths) > 0 {
		for _, sp := range generator.config.SearchPaths {
			filepath.Walk(path.Join(generator.tmpDir, sp), func(p string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}

				name := strings.ToLower(info.Name())
				if info.IsDir() || !(strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")) {
					return nil
				}

				bytes, err := os.ReadFile(p)
				if err != nil {
					return nil
				}
				buf.Write(bytes)

				return nil
			})
		}
	}

	if len(generator.config.KustomizePaths) > 0 {
		for _, sp := range generator.config.KustomizePaths {
			directory := path.Join(generator.tmpDir, sp)
			bytes, err := kustomize.Render(directory)
			if err != nil {
				continue
			}
			buf.Write(bytes)
		}
	}

	if len(generator.config.GenPaths) > 0 {
		for _, sp := range generator.config.GenPaths {
			directory := path.Join(generator.tmpDir, strings.TrimSuffix(sp, "..."))
			bytes, err := genall.Render(directory)
			if err != nil {
				continue
			}
			buf.Write(bytes)
		}
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

func (generator GitGenerator) Versions() ([]string, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	iter, err := generator.repo.Tags()
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0)
	if generator.config.IncludeHead {
		tags = append(tags, referenceHead)
	}
	iter.ForEach(func(r *plumbing.Reference) error {
		tags = append(tags, r.Name().Short())
		return nil
	})

	return tags, nil
}

func (generator *GitGenerator) ensureLoaded() error {
	if len(generator.tmpDir) != 0 {
		return nil
	}

	tmpDir, err := os.MkdirTemp("", generator.config.Name)
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	repo, err := git.PlainClone(tmpDir, &git.CloneOptions{
		URL:               generator.config.Repository,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	generator.head = head.Hash()
	generator.repo = repo
	generator.tmpDir = tmpDir

	return nil
}
