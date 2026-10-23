# 12: Final integration pass

**What to build:** Docs and a manual playtest of the whole overhaul.

**Blocked by:** 07, 09, 10, 11

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] AGENTS.md describes the Barrel roll (Z, timings, immunity), and the lunge rule now says a roll overlap is not contact.
- [x] AGENTS.md also describes the new ship, shot and impact visuals, and the removal of the blink.
- [x] The CONTEXT.md entries for Barrel roll and Invulnerability still match the behaviour.
- [ ] Manual check with `go run . --mock`:
  - glow across speeds;
  - shots and impacts against every archetype;
  - rolling through a Turret ring and a Dart charge;
  - rolling straight after a hit;
  - death, retry and roll ready again.
- [x] Tuning changes are recorded in the spec's Implementation section.
- [x] `go test ./...` is green.

## Completion

Completed: 2026-10-06
Commit: 5425f21
Notes: AGENTS.md is updated and CONTEXT.md still matches, but the user still needs to run the manual `go run . --mock` playtest and no tuning changes were made.
