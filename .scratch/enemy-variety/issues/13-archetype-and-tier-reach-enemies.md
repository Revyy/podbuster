# 13: Archetype and Tier reach enemies

**What to build:** Every enemy spawns with an Archetype and a Tier derived in the bridge, and plays with the shared tier stats. See decisions 03, 04, 06 and 09.

**Blocked by:** 12

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] `PodInfo` carries effective CPU and memory requests, summed over regular and sidecar containers, from the real provider and the mock.
- [x] Game events carry the Archetype and Tier enums defined on the game side. `game/` imports neither `cluster/` nor `bridge/`.
- [x] The bridge maps owner kind to Archetype and computes Tier against the median of the session's known eligible Pods with requests. Initial Pods are banded together after sync, and later Pods on arrival. Request changes update the median only.
- [x] Enemies store Archetype and Tier. Hits use rows of 2/3/5 (Dart 1/2/3). Size is 0.85/1/1.3× with a scaled hitbox. Cruise speed is 1.1/1/0.85×. Contact damage is 1/1/2.
- [x] Separation accounts for each pair of enemy sizes.
- [x] A practice respawn keeps its Archetype and Tier.
- [x] Until rendering lands, archetypes may draw with the existing enemy visuals, scaled by tier.
- [x] Bridge tests cover banding: a uniform namespace, no requests, a lone Pod, Light and Heavy, and the median after deletes. Scene tests cover tier stats. `go test ./...` is green.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `f0f4a8a` (merged into `feat/enemy-variety-impl`)
- **Notes:** The median is computed per resource over Pods that request it, and the score is the larger ratio. Exactly ½ and exactly 2× count as Standard. `NewEnemy` still returns a Standard Probe.
