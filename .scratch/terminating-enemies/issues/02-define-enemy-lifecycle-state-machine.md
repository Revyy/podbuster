# Define the enemy lifecycle state machine

Type: grilling
Status: resolved
Blocked by: 01

## Question

Given the scope settled in "Define scope and triggers for the terminating-enemy visual", what
concrete states does `Enemy` need, what does each one exclude from today's gameplay systems
(homing, contact, Separation, collisions), what do the visuals look like, and does anything about
kill counting, session reset, or the player's own death-animation freeze need to change?

## Answer

`Enemy` replaces its current `Active bool` with an explicit `EnemyState` enum plus a state timer:

- `EnemyActive` — today's existing behavior, unchanged.
- `EnemyTerminating` — entered the instant `Hit()` drives `Hits` to zero, **or** when the bridge
  reports the Pod already terminating for any other reason (see ticket 03), with no distinction
  between the two triggers from this point on. A `Terminating` enemy is frozen: it does not home
  toward the player, deals no contact damage, and is excluded from Separation steering, matching
  exactly how inactive enemies are excluded today. It renders as a dimmed/desaturated sprite with a
  slow pulse, distinct from the existing hit-flash. It advances only its own state timer.
- `EnemyExploding` — entered once the Pod is confirmed gone (the real `EventDespawnEnemy`, or the
  practice-mode simulated equivalent from ticket 04). A brief one-shot beat, 24 ticks, reusing the
  player's existing explosion ring effect (`drawExplosion`) scaled down to enemy size, then the
  enemy is removed from the pool.
- A `Terminating` enemy whose chaos-mode deletion request fails synchronously reverts straight back
  to `EnemyActive` (ticket 01).

Every call site currently gated on `Active` (`updateEnemies`, `avoidanceVector`, `resolveCollisions`,
`resolveContact`, `separate`, `prune`) is updated to gate on `EnemyActive` specifically;
`Terminating` and `Exploding` enemies get their own small update/prune paths that only advance their
timer and, for `Exploding`, remove the enemy once it expires.

Unaffected, deliberately:

- **Kill counting**: `Scene.Kills()` still increments in `recordKill`, at the moment `Hit()` reaches
  zero (i.e. the instant an enemy enters `Terminating`), exactly as today. Externally-caused
  Terminating (no player `Hit()` involved) never increments it.
- **Session reset**: `Scene.Reset()` still nils the entire enemy slice on session end; a
  `Terminating` or `Exploding` enemy is discarded along with everything else, no special-casing.
- **The player's own death animation**: `UpdateWithoutPlayer` (run during `StateDying`) keeps
  advancing enemy state timers normally. Real cluster time and practice mode's fixed-tick timers
  have no reason to pause for the player's unrelated death beat.
