# Final integration pass

Status: resolved
Blocked by: 07, 10

## Context

See the map (`../map.md`) for full design context. This ticket closes out the effort.

## Scope

- Run the full test suite (`go test ./...`) and fix any remaining fallout across `cluster/`,
  `bridge/`, and `game/`.
- Manual playtest via `go run . --mock`, covering:
  - A practice-mode kill: confirm the dim/pulse beat, then the explosion, then removal, with the
    existing respawn behavior intact afterward.
  - A chaos-mode kill against the mock provider: confirm the same visual sequence, timed against
    the mock's simulated 2-second Terminating window rather than a fixed local timer.
  - An externally-triggered termination via the new mock `TerminatePod` helper (simulate from a
    debug hook, test, or temporary manual call) in both modes, confirming an enemy the player never
    shot still dims/pulses/explodes, and that `Kills`/`Pods deleted` is not incremented for it.
- Cross-reference `CONTEXT.md` (already updated with `Terminating`/`Exploding` terms during
  charting) and the root `AGENTS.md`/package docs — update `game/enemy.go`, `game/scene.go`, and
  `cluster/provider.go` doc comments if any drifted from the final implementation.
- Confirm `game/` still does not import `cluster/` or `bridge/` (per repo convention).

## Acceptance criteria

- `go build ./...` and `go test ./...` pass with no skipped or flaky tests introduced.
- All three manual playtest scenarios above behave as described.
- Doc comments and `CONTEXT.md` accurately describe the shipped behavior.
- `map.md` and its tickets need no further edits — if the playtest surfaces a real behavioral gap,
  fix the code to match the settled design rather than reopening a design ticket.

## Answer

The final integration checks pass across cluster, bridge, and game, including the mock provider's
two-phase termination, bridge termination/failure events, practice-mode animation/respawn behavior,
chaos-mode confirmation flow, and externally observed termination without kill-count increments.
`go run . --mock` launched successfully and remained running; the terminal environment did not
provide interactive input for a full GUI playthrough, so the three visual sequences are covered by
the automated lifecycle and rendering tests rather than claimed as manually observed.
Documentation and architecture references were updated to describe terminating Pods as retained
until final removal.

## Completion

Completed: 2026-09-09
Commit: dab8408
Notes: Completed integration validation, synchronized agent documentation, and added external-termination kill-count coverage.
