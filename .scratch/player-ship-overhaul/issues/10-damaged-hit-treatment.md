# 10: Damaged hit treatment

**What to build:** Replace the Invulnerability blink with the "Damaged" treatment. See [Prototyping the ship's damage and death treatments](04-prototyping-the-ship-damage-and-death-treatments.md) and branch `prototype/ship-damage`.

**Blocked by:** 05, 08

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] During Invulnerability the ship is never hidden. `playerVisible` and the blink are removed.
- [x] The ship shakes for 10 ticks, easing out, up to about 1.6 px.
- [x] A red-tinted hull pulse runs in 6-tick phases for the whole window.
- [x] For 45 ticks, six sparks spray from the hull and the engine glow drops out on about 30% of frames.
- [x] Smoke puffs trail behind the engines, fading with the remaining timer.
- [x] While rolling, only the roll animation is drawn.
- [x] `go test ./...` is green.

## Completion

Completed: 2026-10-06
Commit: ee88c58
Notes: Replaced the blink with the damaged hit treatment so the ship stays visible under pressure.
