# 08: Barrel roll mechanics and HUD

**What to build:** The Barrel roll ability, its immunity and its HUD row, drawn with a placeholder until ticket 09. See [Prototyping the Barrel roll and HUD indicator](03-prototyping-the-barrel-roll-and-hud-indicator.md) and the Notes in `map.md`.

**Blocked by:** None

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] Z (just pressed) starts a 45-tick roll when the 180-tick cooldown is 0, and the cooldown starts at the same moment. Both are constants.
- [x] While rolling, `resolveContact` and `resolveBolts` skip the player entirely: no damage, no knockback, no lunge end, and Bolts stay live.
- [x] Rolling doesn't start, extend or pause the Invulnerability timer.
- [x] Steering, thrust and firing work during the roll, and the roll doesn't move the ship.
- [x] `Player.Reset` clears the roll and cooldown, so retry and new sessions start ready. There's no rolling during Dying.
- [x] A `Z      Roll` panel with a 4 px cooldown bar sits directly below the weapon status panel, using the existing charging and ready colours and `cooldownFillRatio`.
- [x] Player, scene and game tests cover all of the above.
- [x] `go test ./...` is green.

## Completion

Completed: 2026-10-06
Commit: d1e8b21
Notes: Landed the Barrel roll rules, immunity window, cooldown HUD, and the tests around them.
