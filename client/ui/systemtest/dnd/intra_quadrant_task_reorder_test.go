package dnd

import (
	"os/exec"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// Placeholder reorder: simulate filename prefix swap and assert order
func TestDnD_IntraQuadrantTaskReorder_Smoke(t *testing.T) {
	repoRoot := t.TempDir()
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_populated")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
		}
	}
	run("init")
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "init")

	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}

	dir := filepath.Join("todo", "urgent-important")
	// Reorder: rename 001-task-u1.json -> 002-task-u1.json (simulate moving down)
	old := filepath.Join(dir, "001-task-u1.json")
	tmp := filepath.Join(dir, "tmp-task-u1.json")
	newName := filepath.Join(dir, "002-task-u1.json")
	if out, err := exec.Command("mv", filepath.Join(repoRoot, old), filepath.Join(repoRoot, tmp)).CombinedOutput(); err != nil {
		t.Fatalf("mv tmp: %v\n%s", err, string(out))
	}
	if out, err := exec.Command("mv", filepath.Join(repoRoot, tmp), filepath.Join(repoRoot, newName)).CombinedOutput(); err != nil {
		t.Fatalf("mv new: %v\n%s", err, string(out))
	}
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "reorder")

	// Expect order to start with 002-... in that directory
	if err := h.AssertOrderByPrefix(repo, dir, []string{"002-task-u1.json"}); err != nil {
		t.Fatalf("order assert failed: %v", err)
	}
}
