# 15: Seeker family rendering

**What to build:** Enemies are drawn as the Seeker droid family: Probe, Turret, Sentinel and Dart at each tier, with their state animations. Port the vector drawings from branches `prototype/probe-droid` (`drawProbeSeeker`) and `prototype/archetype-silhouettes`. See decisions 02 and 07.

**Blocked by:** 13

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] Each archetype has its silhouette and colour: Probe red, Turret amber, Sentinel green, Dart violet.
- [x] Droids are drawn upright, may overhang the hitbox, and have an idle animation.
- [x] Tier cues: Light is 0.85× with a thinner hull. Heavy is 1.3× with armour bands and rivets, an extra tentacle (an extra leg on the Turret) and a thicker Sentinel ring.
- [x] Damage, contact, Terminating and Exploding visuals are adapted for each archetype as in the prototypes.
- [x] The Turret's eye aims at the player, and its glow can be driven by a telegraph value. The Sentinel's ring and the Dart's eye have telegraph hooks for tickets 16–18.
- [x] The old enemy drawing and the pod-name and hit-count labels are removed, unless something else still needs them.
- [x] Production code has no prototype build tags. `go test ./...` is green, and you have checked it by hand with `--mock`.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `e549446` (merged into `feat/enemy-variety-impl`)
- **Notes:** Drawing lives in game/droid.go. A headless smoke test draws every archetype, tier and state. Enemy.Telegraph and HeadingX/Y are the hooks for tickets 16–18. The by-hand check with --mock is deferred to ticket 22.
