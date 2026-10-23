# Define scope and triggers for the terminating-enemy visual

Type: grilling
Status: resolved
Blocked by: None

## Question

Should visually distinguishing a Pod that is genuinely in the process of being killed (blocked on
a PodDisruptionBudget, running a pre-stop hook, or catching the shutdown signal) apply only to
chaos-mode kills caused by the player, or more broadly? And what happens at the edges — a
synchronous delete failure, and a Pod stuck Terminating indefinitely?

## Answer

Applies in **both** practice and chaos mode, and to **any** Pod the namespace watch reports as
Terminating, regardless of cause: a player's kill in chaos mode, or an external cluster actor
(another controller, a scale-down, a manual `kubectl delete`) deleting a Pod the player never shot,
in either mode. This is a general visualization of cluster reality, not just feedback for the
player's own actions.

We rely purely on Kubernetes' generic Terminating signal (`DeletionTimestamp` set but the Pod not
yet gone) rather than special-casing PodDisruptionBudgets, pre-stop hooks, or SIGTERM handling
individually — they all look identical from the watch's point of view, and that's enough.

Edge cases:

- If a chaos-mode delete request fails synchronously (RBAC error, network error) before any
  Terminating state is ever observed, the enemy reverts straight back to `Active`: nothing is
  actually happening to the Pod, so the enemy shouldn't look like it is either.
- If a Pod is stuck Terminating indefinitely (e.g. a stuck finalizer), the enemy waits forever.
  There is no timeout that forces it to explode or despawn — a stuck Terminating Pod is itself
  real, meaningful information about the cluster, and papering over it would hide that.
- Practice mode requests no deletions (per `CONTEXT.md`'s existing definition), but its watch is
  still live and reports real cluster truth, so an externally-terminated Pod during a practice
  session animates identically to chaos mode.
