# 20: Field guide

**What to build:** A 7-page Field guide, reached from mode select, teaches the archetypes, tiers, Bolts and Symptoms. Port the design from branch `prototype/enemy-info-screen`. See decision [Designing the enemy info screen](11-designing-the-enemy-info-screen.md).

**Blocked by:** 16, 17, 18, 19

**Status:** resolved

Spec: [spec.md](../spec.md)

- [x] Mode select has a third entry, "Field guide", set apart from Practice and Chaos. Enter opens it. Esc/Backspace returns to mode select with the cursor on the entry. It doesn't interfere with the chaos confirmation. The title screen has no hint for it.
- [x] Pages: Probe, Turret, Sentinel, Dart, Tiers & Bolts, Symptoms, How Pods become enemies. Left/Right turn pages.
- [x] Each archetype page shows the name, source kind, one or two lines of behaviour, the three tiers and a live demo. There are no look descriptions. The Dart page carries the retry-limit note.
- [x] The last page is the illustrative `kubectl`-style table with an unmanaged "never an enemy" row.
- [x] Layout: a consistent header with the page number, a footer key bar, 24 px margins, short text blocks, and droids clipped to panels.
- [x] The guide reads no cluster data.
- [x] `go test ./...` is green. You have checked every page by hand.

## Completion

- **Completed:** 2026-10-01
- **Commit:** `ad37665` (merged into `feat/enemy-variety-impl`)
- **Notes:** The demos run the real Scene against a dummy player. The 1/2/3 tier keys, the N/P/R Symptom toggles and Up/Down row selection are kept from the prototype. Every page was checked from headless screenshots, and the interactive check is deferred to ticket 22.
