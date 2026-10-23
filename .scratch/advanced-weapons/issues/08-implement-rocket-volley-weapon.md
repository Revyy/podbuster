# Implement the rocket volley secondary weapon

Status: resolved
Blocked by: 07

## Context

Implements the tuning settled in [Tune the rocket volley](02-tune-the-rocket-volley.md), built on
the `Weapon` interface and shared projectile model from
[07](07-refactor-weapon-slots-and-shared-projectile.md).

## Scope

- A `RocketWeapon` (or similarly named) `Weapon` implementation equipped in the Player's secondary
  slot, fired on `X` (`inpututil.IsKeyJustPressed`).
- One fire creates three rocket projectiles: a center shot plus two side shots at roughly ±8° to
  the ship's heading, offset a little left/right of the nose so the pattern reads as one burst.
- Each rocket is a 10×10 projectile, 9.0 px/tick, hit cost 2, no homing/splash/pierce/knockback.
- A 180-tick cooldown between volleys; unlimited volleys otherwise.
- Rocket rendering: short orange-auburn bolt, a small bright exhaust trail, and a minimal muzzle
  flash on fire, distinguishable at a glance from the basic shot.

## Acceptance criteria

- Pressing X when off cooldown spawns exactly three rockets in the described pattern and starts
  the 180-tick cooldown; pressing X again before it elapses fires nothing.
- Each rocket costs a hit enemy exactly 2 hits per the shared collision loop from
  [07](07-refactor-weapon-slots-and-shared-projectile.md); overkill still clamps at zero.
- The main weapon's own cooldown and firing are unaffected by the secondary firing or its cooldown.
- `go test ./...` passes; add coverage for the volley's spawn pattern, cost, and cooldown gating.

## Completion

Completed: 2026-09-01
Commit: 29ce41c
Notes: Added a three-round rocket volley with a 45-tick cooldown, shared 2-hit impact cost, and test coverage for pattern/cost/cooldown behavior. Reviewed and fixed a 10x10 collision-size spec gap plus standards/rendering findings. Cooldown later retuned to 180 ticks per updated [02](02-tune-the-rocket-volley.md); see follow-up commit.
