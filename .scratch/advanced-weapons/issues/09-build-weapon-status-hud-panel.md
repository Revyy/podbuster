# Build the weapon status HUD panel

**Status:** resolved
Blocked by: 07, 08

## Context

Implements the design settled in [Design weapon status in the HUD](04-design-weapon-status-in-the-hud.md).
A prototype of three variants is archived on the throwaway branch `prototype/weapon-hud-status`
(commits `ab48d53`, `348028f`) for reference; do not merge that branch, reimplement variant A
cleanly against the real `Weapon`/cooldown state from
[07](07-refactor-weapon-slots-and-shared-projectile.md) and
[08](08-implement-rocket-volley-weapon.md).

## Scope

- A boxed HUD panel in the bottom-left corner (around (8, 500), sized ~170×46) with two lines:
  `SPACE  <main weapon name>` and `X  <secondary weapon name>`.
- A thin horizontal cooldown bar under the secondary line, filling left-to-right as the secondary
  recharges, colored amber while charging and green when ready.
- Panel reads real `Weapon` state (name, cooldown remaining vs. total) rather than fixed strings,
  so it stays correct if weapon names or the cooldown duration change later.
- No changes to the existing health bar, Enemies/Kills column, or chaos-mode banner; panel must
  not overlap them at the default 800×600 layout.

## Acceptance criteria

- [x] Panel renders during `StatePlaying`, `StateDying`, and `StateGameOver` alongside the rest of the
  HUD (matches where `drawHUD` is already called).
- [x] Cooldown bar visibly empties on secondary fire and refills over the 180-tick window, ending green
  at 100%.
- [x] `go vet`/`go build` clean; no new test framework needed for pure rendering, but if the panel
  exposes any pure helper (e.g. a fill-ratio function), unit test it.

## Completion

Completed: 2026-09-01
Commit: 2997bd3
Notes: Added a live weapon-status HUD with a secondary recharge bar and tested cooldown progress.
