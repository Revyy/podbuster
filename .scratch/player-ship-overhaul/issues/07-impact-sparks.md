# 07: Impact sparks

**What to build:** A short effect where a player projectile hits an enemy. See the impact section of [Prototyping Blaster shots and Rockets](02-prototyping-blaster-shots-and-rockets.md).

**Blocked by:** 06

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] When a projectile hits an enemy, the Scene records an impact effect with the position, kind and direction, much as Bolt bursts are recorded. Effects age out after about 18 ticks.
- [x] A Blaster hit shows a small cyan flash with five white sparks spraying backward. A Rocket hit shows an orange and yellow fireball with a grey smoke ring.
- [x] Shots leaving the screen leave no effect. Retry, scene reset and session end clear effects.
- [x] Scene tests cover recording on a hit, no recording on leaving the screen, ageing out, and clearing on reset.
- [x] `go test ./...` is green.

## Completion

Completed: 2026-10-06
Commit: 989f379
Notes: Added short-lived projectile impact effects and the scene coverage that keeps them honest.
