# 11: Break-apart death

**What to build:** Replace the Dying ring with the break-apart death. See [Prototyping the ship's damage and death treatments](04-prototyping-the-ship-damage-and-death-treatments.md) and branch `prototype/ship-damage`.

**Blocked by:** 05

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] Over the existing 60-tick Dying window, the Interceptor splits into nose, rear hull and two wings.
- [x] Each piece flies out from the centre, easing out, travels 22–40 px plus some of the ship's last velocity, and spins at its own rate.
- [x] Pieces glow orange-hot, cool to their hull colours, then char and fade.
- [x] An orange and yellow fireball covers the first 35% and eight embers scatter.
- [x] There's no engine glow during Dying.
- [x] `go test ./...` is green.

## Completion

Completed: 2026-10-06
Commit: c3f8c9f
Notes: Replaced the old death ring with the full break-apart Interceptor explosion.
