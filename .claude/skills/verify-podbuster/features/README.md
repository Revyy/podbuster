# Podbuster verification map

This directory is the maintained source for verifying the user-facing behaviour of Podbuster. Read
the index before driving the game, then use the matching feature file as the recipe.

## Baseline preconditions

- Run everything from the repo root with `P=.claude/skills/verify-podbuster/scripts/podctl.sh`.
- `$P launch <run-id> mock` reached `ready:` and `$P doctor <run-id>` prints no `FAIL`.
- The game is on the title screen (`P O D B U S T E R` / `Press ENTER to start`).
- For chaos only: `make kind-up` succeeded and the run was launched with `kind`.
- Never drive a Podbuster window this run did not launch.

## Driving conventions

- Start every recipe from the title screen unless its preconditions say otherwise.
- Menu keys are edge-triggered: `$P key <id> 80 <key>`, then `sleep 0.3` before the next one.
- Weapons fire once per press. Steering and thrust act for as long as the hold lasts.
- The world keeps running between commands. Chain the action and its `shot` in one shell command.
- Leave play by dying: Escape is ignored during play and Dying.

## Proof and skip reporting

- Shoot the state before and after each user action, and quote the on-screen text that changed.
- Chaos proof adds a `kubectl` view of the Pod before and after deletion.
- Record the run id, its mode (`mock` or `kind`) and the git rev from `run.env` with every proof.
- Report an unreachable path with the attempted command and the unmet precondition. Don't
  report a journey as verified through a different entry point.

## Feature entry contract

Each feature file starts with an H1 title and one paragraph describing the user-visible
behaviour. It then uses exactly four H2 sections: `Sub-features`, `How to get to it (user POV)`,
`Driving it with podctl` (starting with `Preconditions:`), and `Gotchas`.

## Features

- [Title and mode select](./title-and-mode-select.md) covers the title screen, the mode cursor,
  the namespace line, and the chaos double-confirmation and back-out.
- [Practice session](./practice-session.md) covers starting practice, the HUD, firing, kills
  and practice respawns. **Proven 2026-10-22 in `mock`.**
- [Game over and retry](./game-over-and-retry.md) covers death, the game-over summary, `R`
  retry in the same mode and Escape to the title.
- [Chaos session](./chaos-session.md) covers the confirmed chaos start, `Pods deleted` and the
  real Pod deletion against the kind playground.
- [Field guide](./field-guide.md) covers opening the guide, paging, tier switching, the Symptoms
  controls and returning to mode select.
