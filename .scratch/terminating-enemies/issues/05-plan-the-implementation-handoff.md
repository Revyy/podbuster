# Plan the implementation handoff

Type: grilling
Status: resolved
Blocked by: 01, 02, 03, 04

## Question

Given the settled scope, state machine, cluster/bridge signaling, and mode-simulation decisions,
what implementation ticket boundaries, dependency order, and acceptance checks will let another
agent build the terminating-enemy feature without reopening design questions?

## Answer

Six implementation tickets, created as `ready-for-agent` child issues:

1. [Add the Terminating signal to the cluster provider](06-add-terminating-signal-to-cluster-provider.md) —
   `PodInfo.Terminating`, `PodProvider` doc updates, and `kube.go` changes so a real Pod's
   `DeletionTimestamp` produces an `EventModified{Terminating: true}` instead of collapsing into
   `EventDeleted`. Blocked by nothing; this is the foundation everything else observes.
2. [Simulate Terminating in the mock provider](07-simulate-terminating-in-mock-provider.md) —
   `MockProvider.DeletePod`'s 2-second delayed-removal window, and the new `TerminatePod` helper
   for externally-caused terminations. Blocked by 06 (needs `PodInfo.Terminating`).
3. [Add bridge event types for Terminating and deletion failure](08-add-bridge-event-types.md) —
   `EventPodTerminating`, `EventDeletionFailed`, and the bridge translation/routing logic
   (`handlePodEvent`, `handleKill`). Blocked by 06.
4. [Build the enemy state machine and scene wiring](09-build-enemy-state-machine-and-scene-wiring.md) —
   the `EnemyState` enum and timer on `Enemy`, every `Active`-gated call site in `scene.go`
   migrated to `EnemyActive`, the optimistic local transition on kill, bridge event handling in
   `drainEvents`, and the practice-mode fixed-tick simulation. Blocked by 08 (needs the new
   `GameEventType`s to consume).
5. [Render enemy Terminating and Exploding visuals](10-render-enemy-terminating-and-exploding-visuals.md) —
   the dimmed/pulsing Terminating sprite and the scaled-down explosion ring in `draw.go`. Blocked
   by 09 (needs `EnemyState` to render against).
6. [Final integration pass](11-final-integration-pass.md) — full test run, manual playtest in both
   modes (including an externally-triggered termination via the new mock helper), and doc
   cross-references. Blocked by 07 and 10.

Tickets 2 and 3 (mock simulation, bridge event types) are parallelizable once 1 lands: they touch
disjoint files (`cluster/mock.go` vs. `bridge/bridge.go` + `game/events.go`) and neither result
depends on the other. Ticket 4 is the integration point and must wait for 3; ticket 5 is purely
additive rendering and must wait for 4. Ticket 6 waits on both the mock simulation (2) and the
visuals (5) so a full manual playtest is possible.
