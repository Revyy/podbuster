# 19: Symptoms end to end

**What to build:** A Pod's readiness and restarts show on its enemy as cosmetic Symptoms, live. See decisions [Deciding live cosmetic overlays](08-deciding-live-cosmetic-overlays.md) and 09.

**Blocked by:** 14, 15

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] `PodInfo` carries Ready (the Pod's Ready condition is True) and Restarts (summed over regular and sidecar containers, excluding ordinary init containers) from the real provider and the mock.
- [x] Spawn events carry Symptoms. The bridge emits a Symptoms-changed event only when Ready or Restarts change.
- [x] Not ready makes the eye flicker irregularly and the antenna spark. Patch plates show none, one or two for 0, 1–4 or 5+ restarts. A restart increase during play triggers a 20-tick reboot blink. Restarts present at spawn show only as patches.
- [x] Terminating and Exploding visuals hide Symptoms. An enemy that reverts shows its current Symptoms.
- [x] A pending practice respawn's record is updated by Symptoms-changed events and shows no blink for restarts that happened while it was away.
- [x] The mock has a Symptoms setter that emits Modified events. The seed table's Symptoms (from decision 10) are applied. A `--mock`-only ticker flips `agent-y`'s readiness every ~10 s and adds a restart to `web-c` every ~20 s. Replacement Pods start not ready for ~3 s.
- [x] Symptoms never affect behaviour or stats.
- [x] Bridge and scene tests cover the above. `go test ./...` is green.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `d5320fd` (merged into `feat/enemy-variety-impl`)
- **Notes:** Adds MockProvider.SimulateActivity (the --mock ticker) and SetBootDelay for tests. A Pod with no Ready condition counts as not ready. A small antenna is drawn only while not ready. Pending respawns keep updated Symptoms and never blink.
