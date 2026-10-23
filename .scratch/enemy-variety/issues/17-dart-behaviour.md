# 17: Dart behaviour

**What to build:** Darts wind up, charge through the player's locked position, overshoot, brake and repeat. This ticket also introduces the shared lunge rules that the Sentinel reuses. See decision 06.

**Blocked by:** 15

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] The cycle is wind-up (tentacles flared, eye locked, drift at most 0.8 px/tick), then lock the player's position, charge through it and overshoot, then brake for 20 ticks.
- [x] Charge speed, overshoot and wind-up per tier: Light 3.0 / 80 / 50, Standard 3.5 / 120 / 40, Heavy 4.0 / 160 / 30.
- [x] A charge stops at the screen edge, inset by half the Dart's size.
- [x] Lunge rules: contact uses normal contact damage and knockback and ends the lunge. Lunges skip soft avoidance but keep the hard no-overlap floor, and a lunge blocked by the floor ends early.
- [x] The Dart faces its direction of travel. The wind-up shows the eye glow.
- [x] Timers restart on revert from Terminating and start fresh on a practice respawn.
- [x] Scene tests cover the cycle timing, overshoot, edge stop and lunge termination. `go test ./...` is green.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `16817a1` (merged into `feat/enemy-variety-impl`)
- **Notes:** Lunge rules live in game/lunge.go (startLunge/endLunge/Lunging), reused by ticket 18. A lunge ends on contact only when damage lands. A Dart off-screen holds its wind-up until it has room to charge. Checking it by hand is deferred to ticket 22.
