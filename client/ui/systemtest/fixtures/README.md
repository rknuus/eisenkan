# System Test Fixtures

Seed repositories used by in-process system tests.

Planned sets:

- `board_eisenhower_minimal/` – minimal columns/sections, few/no tasks
- `board_eisenhower_populated/` – representative board with tasks and subtasks
- `board_eisenhower_large/` – large dataset for scalability/perf checks

Each fixture directory will be copied into a temp repo, then `git init` + initial commit performed in tests.


