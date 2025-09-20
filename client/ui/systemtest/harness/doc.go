// Package harness provides deterministic helpers for in-process Fyne system tests.
//
// Characteristics:
//   - Non-parallel: Fyne app is a singleton; system tests must not use t.Parallel().
//   - Deterministic: fixed theme, scale, and window size; animation settling helpers.
//   - Isolation: each test uses a temporary git-backed repo seeded from fixtures.
//   - Verification: repository assertions (go-git) and UI-state helpers.
package harness
