# 01: Reduce player speed and turn rate

**What to build:** The player ship should feel easier to aim while flying and turning. Reduce the ship's maximum flight speed by 20% and its turn (rotation) rate by roughly 40%, so players can more precisely line up shots without overshooting or spinning too fast.

**Blocked by:** None (can start immediately)

**Status:** resolved

- [x] Player max speed constant is reduced by 20% from its current value
- [x] Player turn rate constant is reduced by roughly 40% from its current value
- [x] Existing player/game tests pass with the updated values
- [x] Manual playtest via `go run . --mock` confirms the ship turns and accelerates noticeably slower and feels easier to aim

<!-- Implementing skills set Status to resolved and add a Completion section after the final commit. -->

## Completion

Completed: 2026-08-28
Commit: c13f1a7
Notes: playerMaxSpeed lowered 5.0 -> 4.0 (-20%) and playerTurnSpeed lowered 0.08 -> 0.048 (-40%) in game/player.go; verified via go test ./... and a `go run . --mock` launch check, then merged feat/test-ticket-skill-updates into main.
