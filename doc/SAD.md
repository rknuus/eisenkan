# EisenKan Software Architecture Document
## Volatility analysis
* notification customization, e.g. email
* customize rules and triggers, e.g. Scrum, SAFe
* automation, e.g. automated status transitions or automatic generation of commit messages
* external data ingestion
* versioning incl. conflict resolution
* schema migrations
* from local to distributed deployment
* integration into 3rd party systems like JIRA, GitHub, Slack
* data customization, e.g. card type
* view customization, e.g. filtering/grouping or visual theming
* usage customization, e.g. keyboard shortcuts

The following potential volatilities are not considered, because EisenKan covers the author's personal needs and is not meant to be used in a team/business context. I.e. such a change is considered a change of the "nature of business".

* scaling from single- to multi-user
* scaling from single-board to multiboard with dependencies
* data volume grows beyond capacity of system
* scaling personal => team => enterprise
* turns into project management suite

## Use cases
### Primary Use Cases (Core Business Value)

1. Organize Tasks by Eisenhower Priority - Create tasks into Eisenhower quadrants, which determines the tasks positions in the "todo" column
(Urgent/Important combinations)
2. Visualize Work Progress - View tasks flowing through workflow columns (typically: todo → doing → done) to understand work status at a glance
3. Move Tasks Through Workflow - Transition tasks between columns to reflect changing work states and progress
4. Manage Task Details - Update and delete task information including descriptions, due dates, tags, and priorities

### Secondary Use Cases (Supporting Functionality)

5. Customize Board Structure - Create, modify, and delete columns to match specific workflows
6. Search and Filter Tasks - Find specific tasks based on various criteria (tags, priority, assignee, due date, etc.)
7. Manage extra task details - Add comments and attachments to tasks for enhanced communication
8. Export/Import Data - Move task data between systems or create backups in various formats

### Activity Diagrams for Core Use Cases

#### Use Case 1: Organize Tasks by Eisenhower Priority

```
    (●) Start
     │
     ▼
┌─────────────────┐
│ Task creation   │
│ requested       │
└─────┬───────────┘
      │
      ▼
┌─────────────────┐
│ Priority matrix │
│ assessment      │
└─────┬───────────┘
      │
      ▼
┌─────────────────┐
│ Position task   │
│ in todo based   │
│ on priority     │
└─────┬───────────┘
      │
      ▼
┌─────────────────┐
│ Task visible    │
│ in board        │
└─────┬───────────┘
      │
      ▼
    (◉) End
```

#### Use Case 2: Visualize Work Progress

```
    (●) Start
     │
     ▼
┌─────────────────┐
│ Board view      │
│ requested       │
└─────┬───────────┘
      │
      ▼
┌─────────────────┐
│ Current state   │
│ retrieved       │
└─────┬───────────┘
      │
      ▼
┌─────────────────┐
│ Progress flow   │
│ visualized      │
└─────┬───────────┘
      │
      ▼
    (◉) End
```

#### Use Case 3: Move Tasks Through Workflow

```
    (●) Start
     │
     ▼
┌─────────────────┐
│ Task movement   │
│ initiated       │
└─────┬───────────┘
      │
      ▼
     ◊ Transition valid?
    ╱ ╲
   ╱   ╲
  ╱     ╲
No╱       ╲Yes
 ╱         ╲
▼           ▼
┌─────────┐ ┌─────────────────┐
│ Movement│ │ Task position   │
│ rejected│ │ updated         │
└─────┬───┘ └─────┬───────────┘
      │           │
      │           ▼
      │     ┌─────────────────┐
      │     │ Board state     │
      │     │ refreshed       │
      │     └─────┬───────────┘
      │           │
      └───────────┘
                  │
                  ▼
                (◉) End
```

#### Use Case 4: Manage Task Details

```
    (●) Start
     │
     ▼
┌─────────────────┐
│ Task operation  │
│ requested       │
└─────┬───────────┘
      │
      ▼
┌─────────────────┐
│ Update/deletion │
│ performed       │
└─────┬───────────┘
      │
      ▼
┌─────────────────┐
│ Task state      │
│ updated         │
└─────┬───────────┘
      │
      ▼
    (◉) End
```

### Architectural Note

These use cases align with the iDesign principle of capturing "required behavior, not required functionality." Each use case represents a complete business scenario that delivers value to the user, rather than just technical operations. The unique aspect of EisenKan is the integration of the Eisenhower matrix prioritization system with traditional Kanban workflow management.

The volatility areas identified in the SAD.md (notifications, rules/triggers, automation, etc.) represent potential future extensions to these core use cases without changing the fundamental business nature of the system.

## Decomposition
### Complete decomposition
The following decomposition contains all components to encapsulate *all* volatilities. Even though the implementation might never reach this degree of compleness, it is important to consider all volatilities from the very beginning ("design big").

```
    ┌─────────────┐    ┌──────────────────┐    ┌──────────────────┐
    │ API Client  │    │ EisenKan Admin   │    │ EisenKan Client  │
    │             │    │ Client           │    │                  │
    └──┬───────┬──┘    └────────┬─────────┘    └─────────┬────────┘
       │       │                │                        │
       │       │                │                        │
       │       └────────────────┴────────────────────────┘
       │                        │
       │                        │
       ▼                        ▼
   ┌───────────┐      ┌──────────────────┐     ┌─────────────────┐     ┌───────────┐
   │  Feed     │•••••▶│  Task Manager    │••••▶│  Notification   │────▶│ Messaging │
   │  Manager  │◀•••••│                  │     │  Manager        │     │ Utility   │
   └─┬───────┬─┘      └─┬────┬────┬────┬─┘     └──────────────┬──┘     └───────────┘
     │       │          │    │    │    │                      │
     │       │          │    │    │    │                      │
     │       │          │    │    │    └─────────────────┐    │
     │       ▼          ▼    │    ▼                      │    │
     │     ┌──────────────┐  │  ┌───────────────┐        │    │
     │     │  Validation  │  │  │  Rule Engine  │        │    │
     │     │  Engine      │  │  │               │        │    │
     │     └──────────────┘  │  └───────┬───────┘        │    │
     │                       │          │                │    │
     └───────┐         ┌─────┘          │                │    │
             │         │                │                │    │
             │         │                │                │    │
             ▼         ▼                ▼                ▼    ▼
           ┌─────────────┐      ┌───────────────┐      ┌───────────────┐
           │  Files      │      │  Rules Access │      │  Board Access │
           │  Access     │      │               │      │               │ 
           └─────┬───────┘      └───────┬───────┘      └───────┬───────┘ 
                 │                      │                      │
                 │                      │                      │
                 ▼                      ▼                      ▼
           ┌─────────────┐      ┌───────────────┐      ┌───────────────┐
           │ Storage API │      │  Rules Repo   │      │  Board Repo   │
           │             │      │               │      │               │
           └─────────────┘      └───────────────┘      └───────────────┘

    Legend:
    - Solid lines (│, ─): Direct dependencies/calls
    - Dashed lines (•): Asynchronous/queued calls
```

The graphical diagram is colorized:

[!decomposition](./media/decomposition.svg)

### Currently supported decomposition

```
      ┌──────────────────┐    ┌──────────────────┐
      │ EisenKan Admin   │    │ EisenKan Client  │
      │ Client           │    │                  │
      └────────┬─────────┘    └─────────┬────────┘
               │                        │
               │                        │
               └────────────────────────┘
                            │
                            │
                            ▼
                  ┌──────────────────┐
                  │  Task Manager    │
                  │                  │
                  └────┬───────────┬─┘
                       │           │
                       │           │
                       │           │
                       │           ▼
                       │        ┌───────────────┐
                       │        │  Rule Engine  │
                       │        │               │
                       │        └───────┬───────┘
                       │                │
                       │                │
                       │                │
                       ▼                ▼
               ┌───────────────┐   ┌───────────────┐
               │  Board Access │   │  Rules Access │
               │               │   │               │
               └───────┬───────┘   └───────┬───────┘
                       │                   │
                       │                   │
                       ▼                   ▼
               ┌───────────────┐   ┌───────────────┐
               │  Board Repo   │   │  Rules Repo   │
               │               │   │               │
               └───────────────┘   └───────────────┘

    Legend:
    - Solid lines (│, ─): Direct dependencies/calls
```

## System Testing Architecture

### Determinism and Isolation
- In‑process UI testing with `fyne.io/fyne/v2/test` to maximize determinism.
- Single app instance per test run; never use parallel execution for system tests.
- Fixed environment: Light theme, scale 1.0, window size 1024×768.
- Each test operates in an isolated temporary directory seeded from fixtures; git repository initialized for verifiable state transitions.

### Suite Structure
- `client/ui/systemtest/` contains the system testing suite:
  - `harness/`: deterministic app/window setup, widget traversal/finders, wait utilities, basic DnD gestures, and repository assertions (go‑git).
  - `fixtures/`: seedable board repositories (minimal, populated, large, invalid/corrupt) and optional JSON schemas for documentation.
  - `journeys/`: end‑to‑end user journeys for `BoardSelectionView`, `BoardView`, `CreateTaskDialog`, and `ApplicationRoot`.
  - `dnd/`: drag‑and‑drop acceptance scenarios for tasks and parent‑anchored subtasks.

### Verification Strategy
- UI‑level assertions for presence, counts, and basic invariants where applicable.
- Repository‑level assertions as durable evidence of workflow operations:
  - Commit advancement, file presence/absence, path moves, order by filename prefix, and clean working tree checks.
- Performance guardrails are measured with lightweight timing utilities to surface regressions without flakiness.

### Drag‑and‑Drop Testing Approach
- Initial phase uses repository‑level simulations to validate persistence semantics.
- Target state replaces simulations with UI‑level drag gestures using Fyne’s test APIs and validates both UI feedback and repository changes.

### Fixtures Strategy
- Minimal fixture for fast journeys; populated and large fixtures for coverage and performance; invalid and corrupt fixtures for negative cases (discovery/validation).
- Optional OS “recent boards” behavior is emulated by a shim used only in tests to avoid platform dependencies.

### Build and Execution
- `make test-system` runs the systemtest packages with linker flags adjusted for macOS to suppress duplicate Objective‑C warnings.
- CI integration is staged: system tests run on PRs; heavier load/performance scenarios can be scheduled for nightly runs.
