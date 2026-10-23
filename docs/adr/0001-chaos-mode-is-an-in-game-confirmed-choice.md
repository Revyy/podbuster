# Chaos mode is an in-game, confirmed choice

Chaos mode deletes real Pods, so the path into it must be deliberate. We removed the `--chaos`
CLI flag entirely: the mode is now selected on an in-game mode-select screen that defaults to
practice mode and requires an explicit confirmation naming the target namespace before play
starts. A flag would let a shell-history recall start deleting infrastructure with no
acknowledgement on screen, and there is deliberately no bypass.

The same reasoning shapes two rules elsewhere in the game that would otherwise look arbitrary:
contact with an enemy never reduces its hits, so only a projectile can cause a kill and only a kill
can cause a deletion; and chaos mode is announced by a permanent HUD banner for the whole
session. Together these guarantee that a Pod dies only when the player deliberately shot it
dead, in a mode they deliberately confirmed.

## Consequences

Provider selection stays a flag (`--mock`), because it decides where data comes from rather than
what the game does to it. Keeping the two separate leaves "mock provider plus chaos mode"
available as a safe demonstration of deletion behaviour.
