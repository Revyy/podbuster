# 09: Barrel roll animation

**What to build:** The squash-flip roll animation on the Interceptor. See [Prototyping the Barrel roll and HUD indicator](03-prototyping-the-barrel-roll-and-hud-indicator.md) and branch `prototype/barrel-roll`.

**Blocked by:** 05, 08

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] Over roll progress t, the hull width scales by |cos 2πt| with a 0.12 floor.
- [x] The darker belly (gunmetal, panel lines, no cockpit) shows while cos < 0.
- [x] The ship scales up by 1 + 0.18·sin πt, a white highlight sweeps the hull as it banks, and the engine plumes keep drawing.
- [x] There are no trails, afterimages or shimmer.
- [x] `go test ./...` is green.

## Completion

Completed: 2026-10-06
Commit: 426123e
Notes: Swapped the placeholder for the squash-flip roll animation and kept the effect clean and readable.
