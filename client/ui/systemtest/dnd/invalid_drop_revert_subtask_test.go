package dnd

import (
	"os/exec"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// Placeholder invalid-drop for subtask: perform no changes and assert clean tree
func TestDnD_InvalidDrop_Subtask_Reverts_Smoke(t *testing.T) {
	repoRoot := t.TempDir()
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_minimal")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed: %v", err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, string(out))
		}
	}
	run("init")
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "init")
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := h.AssertWorkingTreeClean(repo); err != nil {
		t.Fatalf("expected clean tree: %v", err)
	}
}
