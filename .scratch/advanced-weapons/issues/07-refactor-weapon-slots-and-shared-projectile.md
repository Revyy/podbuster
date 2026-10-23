# Refactor weapon slots and shared projectile model

Status: resolved
Blocked by: None

## Context

Implements the ownership and impact-cost model settled in [Define weapon slot
ownership](04-design-weapon-status-in-the-hud.md) — see
[01](01-define-weapon-slot-ownership.md) and [03](03-define-projectile-impact-rules.md) for the
full decisions. Pure refactor: no new player-visible behavior, no new inputs.

## Scope

- Introduce a `Weapon` interface in the `game` package owning firing state, cooldown, and
  projectile creation from a firing pose. `Player` holds exactly two `Weapon` slots (main,
  secondary) and routes Space/X to them during input update; it no longer retains a `shootTimer`
  or constructs `Bullet` directly.
- Replace the single-purpose `Bullet` type with a shared projectile model where every projectile
  (including the existing basic shot) carries a weapon-specific hit cost. Keep the current basic
  shot's cost at 1 so its behavior is unchanged.
- Update `Enemy.Hit()` (or its caller) to accept and apply a variable cost, clamping remaining
  hits at zero — no negative hits, no double kill events.
- Update `Scene.resolveCollisions` so each active projectile resolves against enemies in firing
  order, deactivates itself on its first overlap, and applies its cost exactly once; an enemy
  already inactive from an earlier hit in the same tick is skipped for subsequent projectiles.
  `Scene` remains the only place that may call the code path that can mark a kill; contact damage
  must not reach it.
- No secondary/rocket weapon yet — this ticket only prepares the shared model.

## Acceptance criteria

- `go test ./...` passes with no behavior change to the existing basic-shot Space input (same
  cooldown, same damage, same visuals).
- `Player` has no direct reference to `Bullet`/`shootTimer`; firing and cooldown state live on the
  `Weapon` implementation(s).
- A projectile's hit cost is data on the projectile, not a special case in the collision loop.
- Contact (`resolveContact`) still never calls the hit/kill path.
