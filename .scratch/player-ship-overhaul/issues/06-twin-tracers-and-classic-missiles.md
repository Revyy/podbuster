# 06: Twin tracers and classic missiles

**What to build:** New looks for the Blaster shot and Rockets. See [Prototyping Blaster shots and Rockets](02-prototyping-blaster-shots-and-rockets.md) and branch `prototype/player-shots`.

**Blocked by:** None

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] Blaster shots draw as twin tracers: two glowing cyan lines about 2.5 px either side, with white tips and a tail growing to about 10 px. A small cyan muzzle flash shows for the first 3 ticks.
- [x] Rockets draw as the classic missile: white-grey body, red nose, dark-red fins, a flickering red-orange-yellow flame, smoke puffs that stay put in the world (derived from age), and a muzzle burst for the first 4 ticks.
- [x] Rendering only. Hitboxes, speeds, spread and hit costs are unchanged, and no test changes are expected apart from any test of the old drawing.
- [x] `go test ./...` is green.

## Completion

Completed: 2026-10-06
Commit: d7e3eb8
Notes: Replaced the old player shots with twin tracers and classic missiles without changing weapon behavior.
