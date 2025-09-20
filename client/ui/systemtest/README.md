# System Tests for EisenKan (Fyne)

This module contains in-process "system" tests built on `fyne.io/fyne/v2/test`.

- Non-parallel: Fyne uses a singleton app; do not run these tests with `t.Parallel()`.
- Deterministic: fixed theme, scale (1.0), and window size in the harness.
- Isolation: each test seeds a temporary git-backed board repository from fixtures.
- Verification: tests assert UI state and verify repository state using go-git.

Structure:

- `harness/`: deterministic app/window setup, widget finders, waits, DnD helpers, repo utilities
- `fixtures/`: seed repositories (minimal, populated, large)
- `journeys/`: end-to-end user journeys (board selection, board view, create dialog)
- `dnd/`: drag-and-drop acceptance tests (tasks/subtasks across columns/sections)


