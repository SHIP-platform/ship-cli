# SHIP CLI Agent Instructions

Before implementing any task in this repository:

1. Read `../docs/.agents/context/README.md` and the canonical pages it routes to.
2. Read `../docs/planning/tasks/README.md` and the selected ticket.
3. Follow `../docs/.agents/skills/ship-task-workflow/SKILL.md`.
4. Record analysis and an implementation plan in the ticket before editing CLI
   files.
5. Work on a `task/<ticket-id>-<slug>` branch, never directly on `main`.

Run `go test ./...` for CLI changes and add focused tests for changed behavior.
Preserve unrelated user changes and update CLI documentation when commands,
flags, configuration, or user-facing behavior change.
