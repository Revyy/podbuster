# Chart visual feedback for terminating Pods

Label: wayfinder:map

## Destination

An implementation-ready design and ordered ticket plan for a new enemy lifecycle —
`Active → Terminating → Exploding → removed` — so that Pods which are genuinely in the process of
being killed (blocked on a PodDisruptionBudget, running a pre-stop hook, or catching the shutdown
signal for graceful termination) are visibly frozen and harmless in-game until they are actually
gone, rather than vanishing instantly the moment a kill or external deletion is observed.

## Notes

- Consult `domain-modeling` in every session and keep `CONTEXT.md` current as terms settle.
- Use `grilling` for decision tickets.
- Applies in both practice and chaos mode, and to any Pod entering Terminating for any reason
  (player kill or external cluster actor) — not just chaos-mode kills.
- "Terminating" is Kubernetes' own generic signal (`DeletionTimestamp` set, Pod not yet gone). No
  per-cause logic for PDBs, pre-stop hooks, or SIGTERM handling: they all look the same to us.
- The enemy-side state is named `Terminating`, never "Dying" — that word stays reserved for the
  player's own `StateDying` death animation.
- A stuck Terminating Pod (e.g. a stuck finalizer) waits forever; there is no timeout.

## Decisions so far

- [Define scope and triggers for the terminating-enemy visual](issues/01-define-scope-and-triggers.md):
  applies in both modes and to any observed Pod entering Terminating for any reason, player-caused
  or not; a synchronous chaos-mode delete failure reverts the enemy straight back to `Active`; a
  stuck Terminating Pod waits forever with no timeout.
- [Define the enemy lifecycle state machine](issues/02-define-enemy-lifecycle-state-machine.md):
  `Enemy` gains an `EnemyState` enum (`EnemyActive`, `EnemyTerminating`, `EnemyExploding`) replacing
  `Active bool`; `Terminating` is frozen (no homing, no contact damage, excluded from Separation)
  and renders dimmed with a slow pulse; `Exploding` is a brief one-shot beat reusing the player's
  explosion ring, scaled down; kill counting stays exactly as today, at the moment `Hits` reaches
  zero; session reset and the player's own death-animation freeze need no special-casing — state
  timers advance normally regardless.
- [Define cluster and bridge signaling for Terminating pods](issues/03-define-cluster-and-bridge-signaling.md):
  `PodInfo` gains an orthogonal `Terminating bool`; `kube.go` stops collapsing terminating pods into
  `EventDeleted` and instead emits `EventModified` carrying the flag; two new `GameEventType`s,
  `EventPodTerminating` and `EventDeletionFailed`, are added to the bridge vocabulary;
  `EventDespawnEnemy` keeps meaning "confirmed gone" but now triggers `Exploding` instead of instant
  removal; the game moves an enemy to `Terminating` optimistically on its own `Hit()` reaching zero,
  with bridge events only confirming, reverting, or finalizing afterward; events for an unknown
  Pod are a silent no-op, matching today's behavior.
- [Define practice-mode and mock-provider simulation](issues/04-define-practice-mode-and-mock-simulation.md):
  practice mode plays the identical `Terminating` (24 ticks) then `Exploding` (24 ticks) sequence on
  a local fixed timer for every kill, purely for visual consistency, independent of and without
  delaying the existing 180-tick respawn countdown; `MockProvider.DeletePod` stops deleting
  synchronously and instead simulates a 2-second Terminating window before removal; a new
  `TerminatePod`-style helper lets external terminations (Q14/Q18-style scenarios) be exercised the
  same way, while `RemovePod` keeps its existing instant behavior for tests that need it.
- [Plan the implementation handoff](issues/05-plan-the-implementation-handoff.md): six ordered
  implementation tickets (cluster signal → mock simulation → bridge event types → enemy state
  machine + scene wiring → enemy visuals → final integration), each `ready-for-agent` with its own
  acceptance criteria.
- [Add the Terminating signal to the cluster provider](issues/06-add-terminating-signal-to-cluster-provider.md):
  `PodInfo.Terminating` and informer events now distinguish a Pod entering termination from its
  confirmed removal.
- [Simulate Terminating in the mock provider](issues/07-simulate-terminating-in-mock-provider.md):
  mock deletions now emit an immediate terminating event and a delayed confirmed-removal event,
  with `TerminatePod` covering external termination scenarios.
- [Add bridge event types for Terminating and deletion failure](issues/08-add-bridge-event-types.md):
  bridge/game channels now carry terminating Pod observations and synchronous deletion failures.
- [Build the enemy state machine and scene wiring](issues/09-build-enemy-state-machine-and-scene-wiring.md):
  enemies now freeze in Terminating, animate through Exploding on confirmation or practice timing,
  and remain excluded from active combat systems throughout.
- [Render enemy Terminating and Exploding visuals](issues/10-render-enemy-terminating-and-exploding-visuals.md):
  terminating enemies now pulse dimmed and desaturated, while confirmed removals show a compact
  expanding explosion ring.
- [Final integration pass](issues/11-final-integration-pass.md): full tests and build pass, mock
  launch succeeds, lifecycle behavior is covered end-to-end by package and rendering tests, and
  documentation now matches the terminating-Pod model.

## Not yet specified

- None at the current frontier.

## Out of scope

- Per-cause detection or handling of PodDisruptionBudgets, pre-stop hooks, or SIGTERM/graceful
  shutdown specifically — all covered generically via the Kubernetes Terminating signal.
- Timeouts or forced removal for Pods stuck Terminating indefinitely.
- Simulating deletion failure or a stuck Terminating pod in `MockProvider` — mock always succeeds
  after its fixed delay.
