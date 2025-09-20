## NEXT STEPS: System Test Expansion Plan

### Principles
- **Deterministic UI-level tests** using `fyne.io/fyne/v2/test` (no parallelism), minimal sleeps, explicit waits.
- **UI actions first, repo-state verification second** (via go-git) as STR evidence.
- **SRS-driven coverage**: Each test cites requirement IDs; destructive and negative cases included.
- **Performance budgets** enforced by lightweight timing checks.

### Scope
This plan expands system tests in `client/ui/systemtest` to validate the Client/UI SRS documents:
- `ApplicationRoot_SRS.md`
- `BoardSelectionView_SRS.md`
- `BoardView_SRS.md`
- `CreateTaskDialog_SRS.md`
- `TaskWidget_SRS.md` (indirectly via embedding and integration)

---

### Harness Upgrades (Enablers)
Add capabilities to `client/ui/systemtest/harness/`:
- **UI gestures**: drag and drop helpers (`Drag`, `DragBy`, drop targeting), hover/overlay assertions.
- **Keyboard/shortcuts**: `SendKey`, `SendShortcut` (e.g., Cmd+Q), Enter/ESC, Tab traversal.
- **Timing utilities**: `Measure(fn) time.Duration`, `AssertUnder(d, budget)`.
- **Engine/manager fakes**: Minimal fakes for `WorkflowManager`, `FormValidationEngine`, `DragDropEngine`, `LayoutEngine` with programmable outcomes (success, validation error, latency).
- **Finders**: Labels by text, list items by content, error overlays.
- **Data factories**: Builders for tasks/subtasks with priorities to seed views without file I/O when appropriate.

### Fixtures Updates
- Boards: `board_invalid/` (missing structure), `board_corrupt/` (bad JSON), `board_large/` (≈1000 tasks spread across quadrants).
- Recent boards: shim/registry to emulate OS “recently used” list for tests.

---

### ApplicationRoot Journeys (new `application_root_journey_test.go`)
- **Startup shows selection**: AR-REQ-001..004; title set to "EisenKan" (AR-REQ-002/003).
- **Select board → BoardView**: AR-REQ-005..008; calls `BoardView.LoadBoard()` (AR-REQ-007); title includes board name (AR-REQ-012).
- **Back navigation**: AR-REQ-009..013; refresh BoardSelectionView (AR-REQ-011).
- **Transition guards**: AR-REQ-014..018; complete ≤500ms; prevent reentrancy.
- **Shutdown**: AR-REQ-019..023 via close and Cmd+Q; fatal error dialogs AR-REQ-024..028.

### BoardSelectionView Journeys (expand `board_selection_journey_test.go`)
- **Browse valid/invalid**: BSV-REQ-001..006, 009. Error messaging for invalid.
- **Empty state**: BSV-REQ-008.
- **Metadata display**: BSV-REQ-007 with FormattingEngine fake.
- **Search/filter/sort**: BSV-REQ-011..019, highlight via FormattingEngine.
- **Recent boards (10)**: BSV-REQ-021..023.
- **Selection validation**: BSV-REQ-020, 024; activation by double-click/Enter (BSV-REQ-021).
- **CRUD**: BSV-REQ-025..033 (create/edit/delete with confirmation and error states).
- **App control + keyboard**: BSV-REQ-034..038.
- **Performance**: BSV-REQ-043..046 (init ≤2s, filter ≤500ms, selection ≤200ms, CRUD ≤1s with faked latency).

### BoardView Journeys (expand `board_view_journey_test.go`)
- **Board display**: BV-REQ-001..005 (labels, separation, responsive layout basics).
- **Task display + refresh**: BV-REQ-006..010, 021..025, 029.
- **Column coordination + WIP**: BV-REQ-016..020; enforce limits, visual feedback.
- **Drag-drop happy path**: BV-REQ-011..015; visual cues; WorkflowManager update.
- **Validation integration**: BV-REQ-031..035; FormValidationEngine blocks invalid drops; error UI assertions.
- **Events**: BV-REQ-036..040 (selection, move, state change callbacks).
- **Performance/scalability**: BV-REQ-041..050 using `board_large` and timing checks.

### CreateTaskDialog Journeys (expand `create_task_dialog_journey_test.go`)
- **2x2 matrix with creation quadrant**: CTD-REQ-001..005.
- **CreateMode TaskWidget**: CTD-REQ-006..010 with real-time validation feedback.
- **Drag newly created task**: CTD-REQ-011..015; visual feedback (CTD-REQ-015).
- **Drag-drop engine coordination**: CTD-REQ-016..020 (sequential handling).
- **Lifecycle**: CTD-REQ-021..025 (open/with-data/close/cancel; cleanup).
- **Errors/fallbacks**: CTD-REQ-026..030 (validation, workflow, DnD failures).
- **Integrations + perf/usability**: CTD-REQ-031..045, 046..050 (keyboard navigation).

### DnD Acceptance Tests (replace repo-only proxies with UI)
- **Inter-quadrant task move**: UI drag, visual feedback, manager update, repo verification.
- **Inter-quadrant subtask move**: parent-anchored variant.
- **Intra-quadrant reorder**: UI reorder, order assertion.
- **Invalid drop reverts**: visible error; working tree unchanged.
- **Between sections/columns**: section-only and section+column moves.
- Map to BV-REQ-011..015, 031..035 and STR evidence.

---

### Performance Under Load (aligns with NEXT_STEPS.md item 4)
- **Sustained multi-component operations**
  - Scenario: Open BoardView with `board_large`, run ~50 DnD ops (mixed valid/invalid), frequent refreshes.
  - Budgets: render ≤300ms, DnD feedback <50ms, priority updates ≤500ms.
  - Assert no leaks or UI degradation; use harness timing utilities and periodic GC stabilization.

---

### Prioritization & Milestones
**P0 (foundation + core journeys)**
- Harness: DnD + shortcuts + fakes; Fixtures: invalid/corrupt/large.
- AppRoot: startup/transition/back/shutdown.
- BoardView: 4-column basics + happy-path DnD.
- CreateTaskDialog: creation + DnD.
- BoardSelectionView: browse valid/invalid + selection + recent.

**P1 (validation, errors, WIP, performance)**
- BoardView: validation/WIP enforcement, event callbacks, perf checks.
- BoardSelectionView: search/filter/sort/CRUD.
- CreateTaskDialog: lifecycle/errors/integrations.

**P2 (accessibility + scalability + long-run)**
- Keyboard navigation coverage; 1000-task scenarios; sustained ops suite.

---

### Deliverables & Metrics
- New/expanded tests with SRS IDs in comments.
- Harness utilities and fakes with README and examples.
- STR-ready evidence: repo assertions for state changes; logs for timing/perf.
- Metrics: coverage of requirement groups, pass/fail per SRS cluster, timing histograms for perf budgets.

### Risks & Dependencies

---

### Progress
- Implemented harness enablers: DnD (placeholder APIs), keyboard (placeholder), timing utilities, and basic fakes.
- Added fixtures: `board_eisenhower_invalid` and `board_eisenhower_corrupt`.
 - Scaffolded journeys: BoardSelectionView (browse/invalid/recent) and ApplicationRoot (startup/transition/back/shutdown) — currently skipped pending wiring.
 - Scaffolded BoardView DnD happy-path test (skipped pending wiring).
 - Scaffolded CreateTaskDialog journeys: creation flow and DnD organization (skipped pending wiring).
 - Scaffolded performance under load journey (skipped pending wiring).
 - Added harness recent-boards shim for BoardSelectionView tests.

### Next Actions (P0)
1. Scaffold BoardSelectionView journeys covering:
   - Browse valid/invalid directories (BSV-REQ-001..006, 009) and error messaging.
   - Recent boards list behavior (BSV-REQ-021..023).
   - Basic selection activation via double-click/Enter (BSV-REQ-020..021).
   - Temporarily mark as skipped where implementation is pending.
2. Enhance BoardSelectionView recent journey to use recent store (still skipped). [Done]

### P1 Progress
- Completed: scaffolding BoardView validation/WIP/event callback tests (skipped pending wiring).
- Completed: scaffolding BoardSelectionView search/filter/sort and CRUD tests (skipped pending wiring).
- Started: wiring BoardSelectionView to un-skip first journey.

### Next Actions (P1)
1. Scaffold CreateTaskDialog tests for:
2. Begin wiring BoardSelectionView minimal constructor, list, and recent provider; un-skip one journey.
   - Lifecycle: open/with-data/close/cancel; cleanup (CTD-REQ-021..025).
   - Error handling: validation/workflow/DnD failures (CTD-REQ-026..030).
   - Integrations: DragDropEngine, FormValidationEngine, WorkflowManager (CTD-REQ-031..035).

- UI DnD simulation limitations in `fyne/test` → mitigate with gesture helpers and deterministic waits.
- Integration fakes drifting from real interfaces → pin to interfaces and add compile-time checks in tests.
- Perf flakiness in CI → run timing with tolerances, isolate heavy tests under a dedicated build tag.


