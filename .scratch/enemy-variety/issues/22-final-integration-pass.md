# 22: Final integration pass

**What to build:** Close out the effort: docs, the full suite, and a manual end-to-end check.

**Blocked by:** 12, 13, 14, 15, 16, 17, 18, 19, 20, 21

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] The `cluster/`, `bridge/` and `game/` sections of AGENTS.md describe raw traits, bridge eligibility, the synced marker, the Archetype/Tier/Symptoms events, Bolts and the Field guide.
- [x] `go vet ./...` and `go test ./...` are green.
- [x] Manual `--mock` check: all archetypes and tiers, Bolts, lunges, Symptoms, replacement Pods in chaos mode, and the Field guide.
- [x] `spec.md` gets an `## Implementation` section.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `0d10a98` (AGENTS.md) and `2b6e535` (Bolt tuning), merged into `feat/enemy-variety-impl`
- **Notes:** A code review led to fixes: contact always ends a lunge, Bolts fly on during Game Over, the `--mock` seed follows `--namespace`, Turrets never overlap, and one tier-lookup helper. The user checked everything by hand with `--mock`. The only change from that check was Bolt speed, from 3 to 2.4 px/tick.
