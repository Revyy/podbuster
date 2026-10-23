# Define cluster and bridge signaling for Terminating pods

Type: grilling
Status: resolved
Blocked by: 01

## Question

What new signal does the cluster layer need to expose so the bridge and game can observe a Pod
entering Terminating (as opposed to being fully gone), and what new events does the bridge emit to
the game to drive the state machine defined in "Define the enemy lifecycle state machine"?

## Answer

`PodInfo` gains a `Terminating bool` field, kept orthogonal to `Phase` (real Pods keep their
existing Phase — Running, Pending, etc. — even while `DeletionTimestamp` is set; the two are
independent in the Kubernetes API).

In `kube.go`, `shouldIgnorePod` and the informer's `UpdateFunc` currently collapse any pod with
`DeletionTimestamp` set straight into `EventDeleted`, treating it as already gone. This stops:
`shouldIgnorePod` no longer treats a Terminating pod as ignorable for `ListPods`/`AddFunc` purposes
(it's still a real, eligible enemy), and `UpdateFunc` emits an `EventModified` carrying
`Terminating: true` the moment `DeletionTimestamp` appears, instead of an `EventDeleted`. The real
informer `DeleteFunc` remains the sole source of the final `EventDeleted`, meaning "actually gone."

Two new `GameEventType`s are added to the bridge's vocabulary, alongside the existing
`EventSpawnEnemy` / `EventDespawnEnemy` / `EventError`:

- `EventPodTerminating` — the bridge translates a `Terminating: true` `EventModified` into this;
  the scene moves the matching enemy into `EnemyTerminating` (or no-ops if it's already there via
  the optimistic path below, or if it's already `Exploding`/gone).
- `EventDeletionFailed` — emitted when a chaos-mode `DeletePod` call returns an error synchronously;
  the scene reverts the matching enemy from `EnemyTerminating` back to `EnemyActive`.

`EventDespawnEnemy` (from the real `EventDeleted`) keeps its existing meaning, "confirmed gone," but
now drives the enemy into `EnemyExploding` instead of removing it immediately.

The game does **not** wait for a round-trip to start looking dying: the instant an enemy's own
`Hit()` reaches zero, the scene moves it to `EnemyTerminating` optimistically, on its own. The
bridge's `EventPodTerminating` / `EventDeletionFailed` / `EventDespawnEnemy` only confirm, revert, or
finalize an already-`Terminating` enemy afterward; they never need to be what *starts* the
transition for a player-caused kill (only for externally-caused Terminating, where there is no
local kill to react to).

An `EventPodTerminating` or `EventDespawnEnemy` for a `PodName`/`Namespace` the scene has no enemy
for is a silent no-op, matching `despawnEnemy`'s existing behavior over an empty match.
