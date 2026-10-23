# Define projectile impact rules

Type: grilling
Status: resolved
Blocked by: 01

## Question

How should the shared projectile model apply different Hit costs, resolve multiple rockets touching
one enemy in the same tick, handle overkill, and preserve the rule that only deliberate weapon
impacts can cause a kill?

## Answer

Use a single shared projectile model in which each projectile carries a weapon-specific impact cost,
not a separate impact system per weapon. A standard bullet does 1 hit; a rocket does 2 hits; all
projectiles share the same collision loop and `Enemy.Hit()` damage gate. The current `Scene`
resolution stays the only place that may mark an enemy as killed: contact damage never calls `Hit()`,
so ram damage can never delete a Pod.

During collision resolution, each active projectile checks the active enemies in order. When it finds
its first overlap, it deactivates itself immediately and applies its impact to that enemy once. If a
rocket and a bullet both overlap the same enemy in the same tick, the loop resolves them
sequentially in firing order, but each projectile is processed at most once. If the enemy is already
inactive after a prior hit, later projectiles are skipped because the enemy is no longer active.

Overkill is handled by clamping the enemy's remaining hit count at zero; a projectile cannot produce
negative hits or trigger a second kill event. The rule is: the only path to a kill is a deliberate
weapon impact that reaches `Enemy.Hit()`, while contact, respawn timing, and session state remain
outside the kill decision.
