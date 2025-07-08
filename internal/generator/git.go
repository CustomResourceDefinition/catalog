package generator

import (
	"fmt"
	"os"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

type GitGenerator struct {
	config configuration.Configuration
	repo   *git.Repository
	reader crd.CrdReader
	tmpDir string
}

func NewGitGenerator(config configuration.Configuration, reader crd.CrdReader) Generator {
	return GitGenerator{
		config: config,
		reader: reader,
	}
}

func (generator GitGenerator) Close() error {
	panic("unimplemented")
}

func (generator GitGenerator) MetaData(version string) ([]crd.CrdMetaSchema, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	panic("unimplemented")
}

func (generator GitGenerator) Schemas(version string) ([]crd.CrdSchema, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	panic("unimplemented")
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
		URL: generator.config.Repository,
	})
	if err != nil {
		return err
	}

	generator.repo = repo
	generator.tmpDir = tmpDir

	return nil
}
