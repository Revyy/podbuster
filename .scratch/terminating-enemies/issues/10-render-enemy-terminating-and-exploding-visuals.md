# Render enemy Terminating and Exploding visuals

Status: resolved
Blocked by: 09

## Context

See the map (`../map.md`) and ticket 02 for full design context.

## Scope

`game/draw.go` (wherever enemies are currently drawn, alongside the existing hit-flash/contact-flash
handling):

- `EnemyTerminating`: render the enemy sprite dimmed/desaturated with a slow pulse — distinct from
  the existing fast `FlashTimer`/`DamageFlashTimer` contact and hit flashes. A reasonable approach:
  scale the fill color's brightness with a slow sine/triangle wave driven by the state timer or a
  free-running counter, since a `Terminating` enemy in chaos mode may have no fixed timer to drive
  the pulse period from — use a fixed pulse period (e.g. ~40-60 ticks per cycle) independent of how
  long the enemy stays in this state.
- `EnemyExploding`: draw a one-shot expanding ring reusing `drawExplosion`'s approach (radius grows,
  alpha fades, over the 24-tick duration), scaled down to enemy size rather than the player's larger
  radius/duration. Factor out any shared logic between the player's `drawExplosion` and this new
  enemy version if it keeps both readable, but don't force a shared abstraction if the two remain
  simple enough separately.
- No HUD changes are implied by this ticket.

## Out of scope

- Any state/timing logic — that's ticket 09, already done.

## Acceptance criteria

- A `Terminating` enemy visibly dims/pulses distinctly from existing contact/hit flashes.
- An `Exploding` enemy shows a one-shot expanding ring that completes within its 24-tick duration
  and does not render once removed.
- No visual regression for `EnemyActive` enemies (existing flash/damage-flash rendering unchanged).
- `go build ./...` passes; a manual run via `go run . --mock` shows the sequence (this can be
  confirmed in ticket 11's playtest if not practical to verify headlessly here).

## Answer

Terminating enemies now render as dimmed, desaturated sprites with a fixed 50-tick pulse driven by
the game render tick. Exploding enemies render a compact expanding, fading ring over the existing
24-tick lifecycle duration and skip the normal sprite. Active enemy rendering and existing hit and
contact flashes remain unchanged. Rendering helpers have focused coverage for pulse, desaturation,
and explosion progression.

## Completion

Completed: 2026-09-09
Commit: 79e4789
Notes: Added terminating pulse and scaled enemy explosion visuals with rendering-focused tests.
