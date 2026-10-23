# Define weapon lifecycle rules

Type: grilling
Status: resolved
Blocked by: 01

## Question

How should equipped weapons and cooldown state behave when a session starts, the player dies,
retries in the same session, leaves for the title, or encounters practice-mode respawns and
chaos-mode deletion?

## Answer

Weapon equip (which weapon occupies the main and secondary slot) is fixed for the lifetime of the
`Game` instance: set once when the Player is constructed, never re-equipped by `startSession` or
`Reset`. Equipping, swapping, and acquisition stay out of scope per the map.

Cooldown and firing-state resets follow the existing main-weapon `shootTimer` pattern exactly: both
`Player.Reset` (called on session start and on retry from game over) zero the secondary weapon's
cooldown to fully ready, so neither a fresh session nor a retry ever starts with cooldown debt
carried over.

Cooldown only advances inside `Player.Update()`, which already doesn't run during `StateDying`
(only `Scene.UpdateWithoutPlayer` runs then), so cooldown freezes for the death animation with no
special-casing needed — the same implicit rule the main weapon's `shootTimer` already relies on.

Weapon and cooldown state are Player-owned and orthogonal to the enemy pool and Pod lifecycle:
practice-mode respawns and chaos-mode deletion requests (success or failure) never read or write
Player weapon state, and no new code path should connect them.
