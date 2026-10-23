# Define practice-mode and mock-provider simulation

Type: grilling
Status: resolved
Blocked by: 01, 02

## Question

Practice mode requests no deletions, so nothing will ever report a player-killed enemy's Pod as
really Terminating or really gone. Should practice mode play the new `Terminating` → `Exploding`
sequence at all, and if so, how is its timing driven? Separately, the primary dev loop
(`go run . --mock`) has no real cluster: does `MockProvider` need to simulate the Terminating
window so the whole feature is visible and testable without one?

## Answer

Practice mode plays the **identical** sequence to chaos mode, for visual consistency, rather than
skipping straight from `Active` to `Exploding`: the instant `Hit()` reaches zero, the enemy enters
`EnemyTerminating` for a short fixed **24 ticks** (0.4s at 60fps), purely simulated locally since
there's no real cluster latency to reflect, then `EnemyExploding` for another fixed **24 ticks**,
then removal. This keeps the sequence looking the same regardless of mode; only the trigger and
duration source differ (a local fixed timer vs. real cluster state). The existing 180-tick
practice-mode respawn countdown (`respawnDelayTicks`) is unaffected: it still starts counting at the
same instant as the kill, independent of the new animation.

`MockProvider.DeletePod` stops deleting synchronously. Instead: it marks the pod terminating
internally and broadcasts a `Terminating: true` signal immediately, then after a fixed **2-second**
delay (its own goroutine/timer) actually removes the pod and broadcasts `EventDeleted`, mirroring a
real cluster's graceful-shutdown window. This always succeeds after its fixed delay in this
iteration — no simulated failure or stuck-Terminating pod is needed yet (out of scope for now, see
the map).

A new helper (`TerminatePod`, or an extension to `RemovePod`) lets tests and manual dev sessions
exercise an **externally**-caused termination the same way: it goes through the same
terminating-then-delayed-removal path as `DeletePod`, so the ticket-01 scenario (a Pod terminated by
something other than the player) is exercisable in mock/tests too. The existing `RemovePod` keeps
its current instant, synchronous behavior unchanged, for any test that needs immediate removal
without the new terminating step.
