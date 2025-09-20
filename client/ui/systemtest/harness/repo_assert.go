package harness

import (
	"errors"
	"fmt"

	git "github.com/go-git/go-git/v5"
)

// OpenRepo opens a git repository at path.
func OpenRepo(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

// LastCommit returns the latest commit hash on HEAD.
func LastCommit(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", err
	}
	return head.Hash().String(), nil
}

// AssertCommitAdvanced asserts that the commit hash advanced.
func AssertCommitAdvanced(prev, curr string) error {
	if prev == curr {
		return errors.New("commit did not advance")
	}
	return nil
}

// AssertFileExists ensures the given relative path exists in worktree.
func AssertFileExists(repo *git.Repository, rel string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	_, err = wt.Filesystem.Stat(rel)
	if err != nil {
		return fmt.Errorf("expected file exists: %s: %w", rel, err)
	}
	return nil
}

// AssertFileAbsent ensures the given relative path does not exist in worktree.
func AssertFileAbsent(repo *git.Repository, rel string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	_, err = wt.Filesystem.Stat(rel)
	if err == nil {
		return fmt.Errorf("expected file absent: %s", rel)
	}
	return nil
}

// AssertMoved checks that oldRel no longer exists and newRel exists.
func AssertMoved(repo *git.Repository, oldRel, newRel string) error {
	if err := AssertFileAbsent(repo, oldRel); err != nil {
		return err
	}
	if err := AssertFileExists(repo, newRel); err != nil {
		return err
	}
	return nil
}

// AssertOrderByPrefix checks that the files in dir have the expected order by numeric prefix.
func AssertOrderByPrefix(repo *git.Repository, dir string, expected []string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	entries, err := wt.Filesystem.ReadDir(dir)
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	if len(names) < len(expected) {
		return fmt.Errorf("directory %s has fewer files than expected", dir)
	}
	for i, want := range expected {
		if i >= len(names) {
			return fmt.Errorf("missing file for index %d", i)
		}
		if names[i] != want {
			return fmt.Errorf("order mismatch at %d: have %s want %s", i, names[i], want)
		}
	}
	return nil
}

// AssertWorkingTreeClean checks if there are no unstaged/staged changes.
func AssertWorkingTreeClean(repo *git.Repository) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	st, err := wt.Status()
	if err != nil {
		return err
	}
	if !st.IsClean() {
		return errors.New("working tree not clean")
	}
	return nil
}

// Helper: substring contains (no import for strings just for this)
// contains helper removed; order check now uses exact filenames
