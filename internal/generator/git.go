package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/CustomResourceDefinition/catalog/internal/crd"
	"github.com/CustomResourceDefinition/catalog/internal/genall"
	"github.com/CustomResourceDefinition/catalog/internal/kustomize"
)

type GitGenerator struct {
	config         configuration.Configuration
	reader         crd.CrdReader
	tmpDir, gitDir string
	tags           []string
}

func NewGitGenerator(config configuration.Configuration, reader crd.CrdReader) Generator {
	return &GitGenerator{
		config: config,
		reader: reader,
	}
}

func (generator *GitGenerator) Close() error {
	return os.RemoveAll(generator.tmpDir)
}

func (generator *GitGenerator) VersionSortKey(version string) (int64, error) {
	if err := generator.ensureLoaded(); err != nil {
		return 0, err
	}

	out, err := exec.Command("git", "--git-dir", generator.gitDir, "log", "-1", "--format=%ct", version).Output()
	if err != nil {
		return 0, fmt.Errorf("unable to get commit date for '%s': %w", version, err)
	}

	ts, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse commit date for '%s': %w", version, err)
	}

	return ts, nil
}

func (generator *GitGenerator) MetaData(version string) ([]crd.CrdMetaSchema, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	crds, err := generator.Crds(version)
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

	metadata := make([]crd.CrdMetaSchema, len(schemas))
	for i, s := range schemas {
		metadata[i] = s.CrdMetaSchema
	}
	return metadata, nil
}

func (generator *GitGenerator) Crds(version string) ([]crd.Crd, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	if len(version) == 0 {
		versions, err := generator.Versions()
		if err != nil {
			return nil, err
		}
		version = versions[0]
	}

	err := exec.Command("git", "--git-dir", generator.gitDir, "--work-tree", generator.tmpDir, "checkout", "--force", version).Run()
	if err != nil {
		return nil, fmt.Errorf("unable to checkout '%s': %w", version, err)
	}

	_ = exec.Command("git", "--git-dir", generator.gitDir, "--work-tree", generator.tmpDir, "submodule", "update", "--init", "--recursive").Run()

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

	return crds, nil
}

func (generator *GitGenerator) Versions() ([]string, error) {
	if err := generator.ensureLoaded(); err != nil {
		return nil, err
	}

	return generator.tags, nil
}

func (generator *GitGenerator) ensureLoaded() error {
	if len(generator.tmpDir) != 0 {
		return nil
	}

	tmpDir, err := os.MkdirTemp("", generator.config.Name)
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	gitDir := fmt.Sprintf("%s/.git", tmpDir)

	// while now irrelevant, this is left as a warning - go-git cloning balloons into too much memory usage, so exec out and use git cli for cloning
	err = exec.Command("git", "clone", "--quiet", "--recursive", generator.config.Repository, tmpDir).Run()
	if err != nil {
		return fmt.Errorf("unable to clone: %w", err)
	}

	out, err := exec.Command("git", "--git-dir", gitDir, "tag").Output()
	if err != nil {
		return fmt.Errorf("unable to list tags: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(string(out)), "\n")

	generator.tags = tags
	generator.gitDir = gitDir
	generator.tmpDir = tmpDir

	return nil
}
