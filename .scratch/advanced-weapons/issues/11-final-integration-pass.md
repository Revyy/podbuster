# Final integration pass for the weapons system

**Status:** resolved
Blocked by: 09, 10

## Context

Closes out the advanced-weapons effort once the HUD panel and lifecycle wiring have both landed.

## Scope

- Sweep for any remaining direct references to the old `Bullet` type (tests, other packages) and
  update them to the shared projectile model from
  [07](07-refactor-weapon-slots-and-shared-projectile.md).
- Confirm `game/` still imports neither `cluster/` nor `bridge/` (repo-wide invariant).
- Run the full test suite and a manual `go run . --mock` pass in both practice and chaos mode,
  checking: basic shot unchanged, rocket volley fires/cools down correctly, HUD panel reflects
  both weapons accurately, and retry/death/session-end all leave weapon state as specified in
  [05](05-define-weapon-lifecycle-rules.md).
- Update `docs/agents/domain.md` or `CONTEXT.md` cross-references if any weapon terminology
  changed during implementation from what tickets 01–06 settled.

## Acceptance criteria

- [x] `go build ./...`, `go vet ./...`, and `go test ./...` all pass.
- [x] Manual playtest confirms the five behaviors above.
- [x] No stray references to `Bullet` remain outside of the shared projectile model (or `Bullet` is
  cleanly retired if fully superseded).

## Completion

Completed: 2026-09-01
Commit: d6367ee
Notes: Retired the Bullet compatibility API and completed the mock-game practice and chaos pass.
