package generator

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/CustomResourceDefinition/catalog/internal/configuration"
	"github.com/go-git/go-billy/v6/osfs"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/storage/filesystem"
	"github.com/stretchr/testify/assert"
)

// FIXME: more tests
func TestGit(t *testing.T) {
	p, err := setupGit()
	assert.Nil(t, err)
	assert.NotNil(t, p)

	expectedVersions := []string{"1.0.0"}

	config := configuration.Configuration{
		Kind:       configuration.Helm,
		Repository: fmt.Sprintf("file://%s", *p),
	}

	generator := NewGitGenerator(config, nil)

	versions, err := generator.Versions()
	assert.Nil(t, err)
	assert.Equal(t, expectedVersions, versions)
}

func setupGit() (*string, error) {
	tmpDir, err := os.MkdirTemp("", "generator.config.Name")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	dotgitdir := path.Join(tmpDir, ".git")
	store := filesystem.NewStorage(osfs.New(dotgitdir), nil)

	repo, err := git.Init(store, git.WithDefaultBranch("refs/heads/main"), git.WithWorkTree(osfs.New(tmpDir)))
	if err != nil {
		return nil, err
	}

	cfg, err := repo.Config()
	if err != nil {
		return nil, err
	}

	cfg.User.Name = "runner"
	err = repo.Storer.SetConfig(cfg)
	if err != nil {
		return nil, err
	}

	tree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile("testdata/test-crd.yaml")
	if err != nil {
		return nil, err
	}

	filename := path.Join(tmpDir, "test-crd.yaml")
	err = os.WriteFile(filename, bytes, 0644)
	if err != nil {
		return nil, err
	}

	_, err = tree.Add("test-crd.yaml")
	if err != nil {
		return nil, err
	}

	commit, err := tree.Commit("Add CRD", &git.CommitOptions{Author: &object.Signature{Name: "runner", Email: "runner"}})
	if err != nil {
		return nil, err
	}

	_, err = repo.CommitObject(commit)
	if err != nil {
		return nil, err
	}

	_, err = repo.CreateTag("1.0.0", commit, nil)
	if err != nil {
		return nil, err
	}

	return &dotgitdir, nil
}
