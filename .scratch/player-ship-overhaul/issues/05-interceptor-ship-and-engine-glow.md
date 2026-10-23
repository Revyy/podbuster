# 05: Interceptor ship and engine glow

**What to build:** Replace the line triangle with the vector Interceptor and its speed-driven twin engine glow. See [Prototyping the ship and engine glow](01-prototyping-the-ship-and-engine-glow.md) and branch `prototype/player-ship`.

**Blocked by:** None

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] The Interceptor is drawn rotated to the heading, without wingtip lights, using a rotating pen helper alongside the droids' `pen`.
- [x] The 20×20 hitbox is unchanged.
- [x] Two blue plumes grow with the speed ratio, leave a faint glow at rest, flare and brighten while thrust is held, and flicker slightly out of phase.
- [x] `Player` records whether thrust is held, set from the controls in `updateWithControls`, and drawing reads it. A player test covers it.
- [x] `shipPoints` and the line drawing are removed if nothing else uses them. The Invulnerability blink stays until ticket 10.
- [x] `go test ./...` is green.

## Completion

Completed: 2026-10-06
Commit: b01ec96
Notes: Shipped the Interceptor redraw with speed-based engine glow and the new thrust-held player seam.
