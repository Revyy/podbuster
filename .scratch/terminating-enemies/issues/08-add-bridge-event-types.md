# Add bridge event types for Terminating and deletion failure

Status: resolved
Blocked by: 06

## Context

See the map (`../map.md`) and ticket 03 for full design context.

## Scope

- In `game/events.go`, add two new `GameEventType` constants: `EventPodTerminating` and
  `EventDeletionFailed`, each documented per the existing style (see `EventSpawnEnemy`,
  `EventDespawnEnemy`, `EventError`). Both carry `PodName`/`Namespace` on the existing `GameEvent`
  struct — no new fields needed.
- In `bridge/bridge.go`:
  - `handlePodEvent`: route a `cluster.EventModified` event with `Terminating: true` to
    `game.EventPodTerminating` (routed by `PodName`/`Namespace`). A `cluster.EventModified` event
    with `Terminating: false` (the pre-existing modified-but-not-terminating case) is still
    ignored, unchanged.
  - `handleKill`: when `b.provider.DeletePod` returns an error, in addition to the existing log
    line, emit `game.EventDeletionFailed{PodName: kill.PodName, Namespace: kill.Namespace}` via
    `b.emit`, so the scene can revert the enemy back to active.
  - `EventDespawnEnemy`'s existing meaning and emission (from `cluster.EventDeleted`) is unchanged.

## Out of scope

- Any scene/game-side consumption of the new event types — that's ticket 09.

## Acceptance criteria

- `game.EventPodTerminating` and `game.EventDeletionFailed` exist with doc comments matching the
  existing style.
- A `cluster.EventModified{Pod: {Terminating: true}}` event produces exactly one
  `game.EventPodTerminating` on the bridge's game-events channel.
- A `DeletePod` error in `handleKill` produces exactly one `game.EventDeletionFailed` in addition to
  the existing log line; success produces neither.
- Existing `bridge_test.go` coverage still passes; add tests covering both new emission paths.
- `go build ./...` and `go test ./...` pass.

## Answer

Added `EventPodTerminating` and `EventDeletionFailed` to the game event vocabulary. The bridge now
translates terminating provider modifications, ignores non-terminating modifications, and emits a
failure event when a chaos-mode deletion request fails. Bridge tests cover both emission paths,
successful deletion behavior, exact failure emission, and ignored modifications.

## Completion

Completed: 2026-09-09
Commit: 2b0fec7
Notes: Added bridge routing and game events for Pod termination and deletion failure.
