# Design weapon status in the HUD

Type: prototype
Status: resolved
Blocked by: 01, 02

## Question

How should the HUD name the equipped main and secondary weapons, teach the Space and X inputs, and
show secondary cooldown readiness without crowding the existing health, enemy, kill, and chaos-mode
information?

## Answer

A bottom-left corner loadout panel: a single boxed panel around (8, 500) sized 170×46, stacked two
lines (`SPACE  Blaster` / `X  Rockets`) naming each slot and its input, with a thin horizontal
cooldown bar under the secondary line that fills left-to-right as the volley recharges and
switches color (amber while charging, green when ready). It sits clear of the existing top-left
health bar, top-right Enemies/Kills column, and top-center chaos banner, so no existing HUD
element needs to move or resize.

Rejected: a bottom-center two-slot ability dock (read too much like a MOBA/ARPG ability bar for
this game's minimalist HUD) and a text-only addition to the top-right Enemies/Kills column (too
easy to miss, and crowded that column further).

Prototype asset: three variants built and run via `go run . --mock`, cycled with `[`/`]`, on
throwaway branch `prototype/weapon-hud-status` (commits `ab48d53`, `348028f`). Not merged to
main; kept only on that branch as the primary source for this decision.
