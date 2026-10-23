# Build the enemy state machine and scene wiring

Status: resolved
Blocked by: 08

## Context

See the map (`../map.md`) and tickets 01, 02, 04 for full design context. This is the central
integration ticket: it does not need visuals (ticket 10 handles that), only correct state and
timing.

## Scope

`game/enemy.go`:

- Replace `Enemy.Active bool` with an `EnemyState` type (e.g. `int`/`iota`-based) with values
  `EnemyActive`, `EnemyTerminating`, `EnemyExploding`, plus a state timer field (e.g.
  `StateTimer int`) counting down ticks remaining in the current state (unused/zero for
  `EnemyActive`).
- `Hit(cost int) *KillEvent`: on reaching zero Hits, instead of setting `Active = false` and
  returning a `KillEvent` for immediate removal, transition the enemy to `EnemyTerminating` (with
  its state timer set per mode — see below) and still return the `KillEvent` so `Scene.recordKill`
  fires exactly as today (kill counting is unaffected, per ticket 02).
- Add whatever small helper methods scene.go needs (e.g. `BeginTerminating(ticks int)`,
  `BeginExploding(ticks int)`, `RevertToActive()`, `AdvanceState() (expired bool)`), matching the
  existing style of small, well-named `Enemy` methods.

`game/scene.go`:

- Every call site currently gated on `!e.Active` (`updateEnemies`, `avoidanceVector`,
  `resolveCollisions`, `resolveContact`, `separate`, `prune`) is updated to gate on
  `e.State != EnemyActive` (or equivalent), preserving current behavior for `EnemyActive` enemies
  exactly.
- `Terminating` and `Exploding` enemies get their own per-tick update (advance state timer only —
  no position change, no collisions) run from `Scene.Update`/`UpdateWithoutPlayer` alongside the
  existing per-enemy loop. When an `Exploding` enemy's timer expires, it becomes eligible for
  removal in `prune` (mirroring how `!Active` enemies are pruned today).
- `drainEvents`: handle the two new `GameEventType`s —
  - `EventPodTerminating`: find the matching enemy by `PodName`/`Namespace`; if it's `EnemyActive`,
    move it to `EnemyTerminating` with **no fixed timer** (chaos/real-cluster path — it waits for
    the real `EventDespawnEnemy` to arrive, however long that takes, per ticket 01's "waits
    forever" rule). If no matching enemy exists, or it's already `Terminating`/`Exploding`, no-op
    (ticket 03).
  - `EventDeletionFailed`: find the matching enemy; if it's `EnemyTerminating`, revert it to
    `EnemyActive`. No-op if not found or not currently `Terminating`.
  - `EventDespawnEnemy` (existing case): instead of setting `Active = false` for immediate removal,
    move the matching enemy into `EnemyExploding` with its state timer (24 ticks, per ticket 04),
    reusing whatever mode-appropriate constant ticket 04 calls for — chaos mode and practice mode
    use the same 24-tick `Exploding` duration. No-op if no matching enemy (ticket 03).
- `recordKill`/`Hit()`'s optimistic local transition: the moment `Hit()` reaches zero, the enemy is
  already `EnemyTerminating` before any bridge event arrives (per ticket 03's "optimistic, no
  round-trip needed"). Its state timer depends on mode:
  - **Chaos mode**: no fixed timer — it waits for a real `EventPodTerminating`/`EventDespawnEnemy`
    (which may arrive already, if the cluster is fast) or `EventDeletionFailed`.
  - **Practice mode**: fixed 24-tick `Terminating` timer (ticket 04), after which it transitions
    itself locally to `EnemyExploding` for another fixed 24 ticks, with no bridge involvement at
    all (practice mode requests no deletions, so nothing will ever confirm it externally). The
    existing 180-tick `respawnDelayTicks` countdown is untouched — it already starts at the same
    instant via `recordKill`, independent of this animation.
- `Scene.Reset()` is unchanged (still nils the whole slice) — confirm this in a test rather than
  needing new code (ticket 02).

## Out of scope

- Any rendering (dimming, pulsing, explosion ring) — ticket 10.
- Cluster/bridge-side changes — tickets 06/08 (already done, blocking this one).

## Acceptance criteria

- A chaos-mode kill: enemy enters `EnemyTerminating` immediately with no local timer; stays there
  until an `EventDespawnEnemy` (→ `EnemyExploding` → removed after 24 ticks) or
  `EventDeletionFailed` (→ back to `EnemyActive`) arrives; an unrelated `EventPodTerminating` for
  the same enemy while already `Terminating` is a no-op.
- A practice-mode kill: enemy enters `EnemyTerminating`, transitions itself to `EnemyExploding`
  after 24 ticks, then is removed after another 24 ticks — with no bridge event involved — while
  the existing respawn timer, started at the same kill, is unaffected and still returns the enemy
  to the pool per current behavior.
- An externally-caused `EventPodTerminating` for an enemy currently `EnemyActive` (no local kill
  involved) moves it to `EnemyTerminating`; it does not increment `Scene.Kills()`.
- `Terminating` and `Exploding` enemies are excluded from homing, contact damage, and Separation,
  verified by test.
- `Scene.Reset()` still discards all enemies regardless of state.
- Existing tests (`combat_test.go`, `state_test.go`, `separation_test.go`, `projectile_test.go`,
  etc.) still pass with `Active` references updated to the new state model; add new tests for the
  scenarios above.
- `go build ./...` and `go test ./...` pass.

## Answer

Enemies now use explicit `EnemyActive`, `EnemyTerminating`, and `EnemyExploding` states with
fixed practice/explosion timers and indefinite provider-driven termination. Scene movement,
collisions, contact, and separation only operate on active enemies; bridge events confirm,
revert, or finalize lifecycle transitions. Practice kills use the local 24-tick terminating and
exploding sequence while preserving the existing respawn countdown, and chaos kills wait for
provider confirmation. Repeated hits cannot emit duplicate kill events.

## Completion

Completed: 2026-09-09
Commit: cb674b6
Notes: Added the enemy lifecycle state machine, scene event wiring, mode-specific timing, and tests.
