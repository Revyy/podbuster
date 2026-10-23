# Simulate Terminating in the mock provider

Status: resolved
Blocked by: 06

## Context

See the map (`../map.md`) and ticket 04 for full design context.

## Scope

- `MockProvider.DeletePod` stops removing the pod synchronously. Instead:
  1. Marks the pod terminating internally and broadcasts a `PodEvent{Type: EventModified, Pod: ...}`
     with `Terminating: true` immediately.
  2. Starts a 2-second timer (on its own goroutine) after which it actually removes the pod from
     `m.pods` and broadcasts `PodEvent{Type: EventDeleted, Pod: ...}` — matching a real cluster's
     graceful-shutdown window.
  3. Returns `nil` immediately (the delete request itself still "succeeds" synchronously; only the
     resulting termination is delayed).
- Add a new `TerminatePod(namespace, name string)` method that drives the same
  terminating-then-delayed-removal sequence as `DeletePod`, for simulating an externally-caused
  termination (no player action involved) in manual dev sessions and tests.
- `RemovePod` is unchanged: it keeps its existing instant, synchronous removal for tests that need
  immediate removal without the new terminating step.
- Guard against a pod being deleted/terminated twice (e.g. `DeletePod` called on an
  already-terminating pod) — treat it as a no-op or return an error consistent with the existing
  "not found" error style once actually removed.

## Out of scope

- Simulating deletion failure or a stuck Terminating pod — always succeeds after the fixed delay.
- Any bridge, scene, or game-side consumption of these events — tickets 08/09.

## Acceptance criteria

- Calling `DeletePod` on an existing mock pod: broadcasts `Terminating: true` immediately, then
  broadcasts `EventDeleted` roughly 2 seconds later (test with a short overridable delay or by
  waiting/polling with a generous timeout — avoid a flaky exact-timing assertion).
- Calling `TerminatePod` produces the identical two-phase broadcast sequence without requiring
  `DeletePod`'s "chaos mode only" framing.
- `RemovePod`'s existing instant behavior and existing tests are unaffected.
- Existing `mock_test.go` coverage still passes; add tests covering the new two-phase sequence for
  both `DeletePod` and `TerminatePod`.
- `go build ./...` and `go test ./...` pass.

## Answer

`MockProvider.DeletePod` and the new `TerminatePod` helper now mark Pods terminating and emit an
immediate `EventModified`, then remove them after a two-second simulated graceful-shutdown window
with `EventDeleted`. Repeated termination requests are idempotent, while `RemovePod` remains an
immediate deletion path. Tests cover both two-phase entry points, prompt return, delayed removal,
and the unchanged immediate removal behavior.

## Completion

Completed: 2026-09-09
Commit: dc4ea74
Notes: Added mock terminating lifecycle simulation and public external-termination helper.
