# Add the Terminating signal to the cluster provider

Status: resolved
Blocked by: None

## Context

See the map (`../map.md`) and tickets 01–03 for full design context. This ticket implements the
cluster-layer signal. Reading `CONTEXT.md` before starting is recommended.

## Scope

- Add `Terminating bool` to `cluster.PodInfo` (`provider.go`), documented as orthogonal to `Phase`.
- In `kube.go`:
  - `shouldIgnorePod` no longer treats a pod with `DeletionTimestamp` set as ignorable — it's still
    a real, eligible enemy while terminating. Update its doc comment accordingly.
  - `toPodInfo` sets `Terminating: pod.DeletionTimestamp != nil`.
  - The informer's `UpdateFunc`: when the new pod has `DeletionTimestamp` set (and is not otherwise
    ignorable, e.g. not Succeeded/Failed), emit `PodEvent{Type: EventModified, Pod: toPodInfo(pod)}`
    with `Terminating: true` instead of falling through to the `EventDeleted` branch it currently
    reaches via `shouldIgnorePod`.
  - The real informer `DeleteFunc` is unchanged: it remains the sole source of the final
    `EventDeleted`, meaning "actually gone."
- Update the `PodProvider` and `PodEvent`/`PodInfo` doc comments to describe the new Terminating
  signal and how `EventModified` now carries it.

## Out of scope

- Any bridge or game-side handling of the new field/event — that's tickets 03/08 and 09.
- `MockProvider` changes — that's ticket 07.

## Acceptance criteria

- `cluster.PodInfo` has a `Terminating bool` field.
- A real Pod that acquires a `DeletionTimestamp` (but has not yet been removed) produces an
  `EventModified` with `Terminating: true`, not an `EventDeleted`.
- A Pod's final removal still produces exactly one `EventDeleted`.
- Existing `kube_test.go` coverage still passes; add or update tests covering: a pod transitioning
  to Terminating (expect `EventModified{Terminating: true}`), and a pod's final removal (expect
  `EventDeleted`), distinguishing the two.
- `go build ./...` and `go test ./...` pass.

## Answer

`PodInfo` now carries an orthogonal `Terminating` flag. The Kubernetes provider keeps
non-terminal Pods with a `DeletionTimestamp` eligible, reports their transition as
`EventModified`, and reserves `EventDeleted` for final removal. Coverage distinguishes both
watch events.

## Completion

Completed: 2026-09-09
Commit: fc8bbbb
Notes: Added terminating Pod metadata, informer signaling, provider documentation, and watch coverage.
