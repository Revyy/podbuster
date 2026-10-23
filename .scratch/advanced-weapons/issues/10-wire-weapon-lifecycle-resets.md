# Wire weapon lifecycle resets

**Status:** resolved

## Context

Implements [Define weapon lifecycle rules](05-define-weapon-lifecycle-rules.md).

## Scope

- `Player.Reset` (called on session start in `startSession` and on retry in `updateGameOver`)
  resets both weapons' cooldown/firing state to fully ready, the same way it already zeroes the
  main weapon's `shootTimer`.
- Weapon equip (which weapon occupies each slot) is set once at `Player` construction and is never
  touched by `Reset`, `startSession`, or retry.
- No change needed for the death-animation freeze: confirm cooldown ticking lives only inside
  `Player.Update()`, which already doesn't run during `StateDying`, so this is a verification step,
  not new code.
- Confirm no code path in practice-mode respawn (`Scene.updateRespawns`) or chaos-mode deletion
  (`bridge`/`cluster` delete flow) reads or writes Player weapon state.

## Acceptance criteria

- A test asserts that after `Player.Reset`, both weapons report ready-to-fire cooldown state.
- A test (or inspection noted in the PR) confirms cooldown does not advance while `StateDying` is
  active, mirroring existing main-weapon behavior.
- `go test ./...` passes.

## Completion

Completed: 2026-09-01
Commit: d77cd7d
Notes: Weapon lifecycle resets wired with tests verifying cooldown reset and StateDying freeze.
