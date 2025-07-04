package generator

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

func asdf() error {
	// Destination to save the chart
	tmpDir, err := os.MkdirTemp("", "git")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	repo, err := git.PlainClone(tmpDir, &git.CloneOptions{
		URL: "https://github.com/codereaper/lane",
	})
	if err != nil {
		return err
	}

	iter, err := repo.Tags()
	if err != nil {
		return err
	}

	tags := make([]string, 0)
	iter.ForEach(func(r *plumbing.Reference) error {
		tags = append(tags, r.Name().Short())
		return nil
	})

	return nil
}
