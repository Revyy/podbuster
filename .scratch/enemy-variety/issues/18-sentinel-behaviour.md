# 18: Sentinel behaviour

**What to build:** Sentinels orbit the player and periodically dash in, reusing the lunge rules from ticket 17. See decision 06, including the slower speeds from playtesting.

**Blocked by:** 17

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] The Sentinel cruises to an orbit around the player at 0.8 px/tick tangential, with the direction picked at random at spawn. The cruise speed scales with tier.
- [x] Orbit radius and dash cadence per tier: Light 200 px / 240 ticks, Standard 180 / 180, Heavy 150 / 120. The first cycle has a 0–60 tick offset from the scene RNG.
- [x] A 30-tick telegraph (the ring spins up and brightens) comes before each dash. The Sentinel locks the player's position and dashes at 2.7 px/tick to 40 px past it, then rejoins its orbit.
- [x] The lunge rules apply. Timers restart on revert from Terminating and start fresh on a practice respawn.
- [x] Scene tests cover the orbit radius, cadence, dash target and lunge termination. `go test ./...` is green.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `d003b9d` (merged into `feat/enemy-variety-impl`)
- **Notes:** Logic lives in game/sentinel.go and reuses the lunge rules. The interval runs from one dash start to the next. An off-screen Sentinel holds its telegraph until it is inside the screen. With the spec numbers, a Standard Sentinel may not regain its full radius between dashes. Checking it by hand is deferred to ticket 22.
