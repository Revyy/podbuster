# Aiming crosshair

Status: ready-for-agent

## Problem Statement

The ship is a wireframe triangle that rotates continuously, and the player must
infer where a shot will go from the orientation of a small nose vertex. At a
glance — especially while drifting sideways under momentum, since the ship's
heading and its velocity are independent — it is hard to tell whether the nose is
lined up with an enemy. Players end up firing speculatively and learning where the
gun points from where the bullets went, which is slow feedback in a game where
enemies close on the player.

## Solution

Draw a crosshair a fixed distance ahead of the ship, along its heading. It is a
simple hollow circle that tracks the nose as the player turns, giving a
continuous, pre-fire readout of the shot line. It is present whenever the player
is in control of the ship and can shoot, and absent otherwise.

The crosshair reports the ship's aim and nothing else. It does not know about
enemies, does not snap to targets, and does not lead moving targets — it is a
sight, not a targeting computer.

## User Stories

1. As a player, I want a marker ahead of my ship, so that I can see where my gun
   is pointing without firing a test shot.
2. As a player, I want the marker to follow my heading as I turn, so that I get
   continuous feedback while lining up a shot.
3. As a player drifting sideways under momentum, I want the marker to follow my
   heading rather than my velocity, so that it tells me where bullets go rather
   than where the ship is going.
4. As a player, I want the marker at a distance that spans several enemy widths,
   so that small heading errors are visible before I commit to a shot.
5. As a player, I want the marker close enough to stay on screen during normal
   play, so that it is usable rather than a curiosity that only appears when I
   point at the middle of the map.
6. As a player pointing at a nearby screen edge, I want the marker to simply
   disappear off that edge, so that it never promises a shot my bullets cannot
   make — bullets die at the screen edge and do not wrap, even though my ship
   does.
7. As a player, I want the marker to be visually tied to my bullets, so that I
   read it as "where the shots go" rather than as another game object.
8. As a player, I want the marker drawn faintly, so that it never obscures an
   enemy, a Pod name, a hit counter, or an incoming bullet.
9. As a player, I want the marker to be small and hollow, so that I can see the
   enemy underneath it when my aim is correct.
10. As a player on the title screen, I do not want a crosshair, so that the
    pre-game screens stay uncluttered.
11. As a player on the mode-select screen, I do not want a crosshair, so that my
    attention stays on the practice/chaos decision.
12. As a player whose ship has just been destroyed, I do not want a crosshair
    during the death explosion, so that the HUD does not suggest I can still
    shoot.
13. As a player on the game-over screen, I do not want a crosshair, so that the
    summary is not overlaid with a live aiming aid.
14. As a player who has just taken contact damage, I want the crosshair to stay
    solid while my ship flashes, so that the blink continues to mean "I am
    invulnerable" and only that.
15. As a player, I want the crosshair to appear immediately on retry (`R`) with
    the reset ship, so that the aid is consistent across lives.
16. As a player, I want the crosshair present in both practice and chaos mode, so
    that my aim does not change depending on whether Pods are really deleted.
17. As a player, I do not want a new keybind to learn, so that the control scheme
    stays as small as it is today.
18. As a developer, I want the crosshair position derived from a single tested
    function, so that the geometry cannot silently drift from what the bullets
    actually do.
19. As a developer, I want the aim distance expressed as one named constant, so
    that tuning it later is a one-line change.
20. As a developer, I want the crosshair colour declared alongside the other
    palette entries, so that the game's colours stay in one place.
21. As a developer, I want a test that pins the non-wrapping behaviour, so that a
    future change cannot "helpfully" make the crosshair wrap like the ship and
    quietly start lying to the player.
22. As a maintainer, I want no new package, layer, or cross-layer channel for
    this, so that the three-layer architecture is untouched by a cosmetic HUD
    addition.

## Implementation Decisions

- **Layer**: entirely inside `game/`. No `cluster/` or `bridge/` involvement, no
  new event types, no channel traffic. This is local rendering derived from
  existing player state.
- **Seam**: a new exported method on `Player` that returns the crosshair's world
  position, of the shape `AimPoint() (x, y float64)`. No arguments: the distance
  is a package-level constant declared with the other player tuning constants
  (`playerSize`, `bulletSpeed`, and friends), not a parameter, because there is
  exactly one caller and one distance.
- **Origin**: measured from the ship's centre (`X`, `Y`), not from the bullet
  spawn point at the nose. The nose offset is half the ship size and is not
  perceptible at the chosen distance; measuring from the centre keeps the method
  a plain polar-to-cartesian conversion.
- **Direction**: the ship's `Heading`, matching how bullet velocity is derived at
  fire time. Bullets do not inherit ship velocity, so the heading ray is the
  literal shot path and no velocity term belongs in this calculation.
- **Distance**: approximately 120 px — around four enemy widths ahead, far enough
  to make heading errors legible and close enough to remain on screen through
  normal play.
- **Screen edges**: `AimPoint` deliberately does **not** apply the screen-wrap
  used for the ship's own position. The returned point may lie outside the screen
  bounds and is then clipped naturally by the renderer. This mirrors bullets,
  which are deactivated at the screen edge rather than wrapped. Clamping the
  point to the edge was rejected: it would place the reticle at a distance that
  no longer represents the shot.
- **Rendering**: a `drawCrosshair` helper in the rendering file, called from the
  world-drawing routine. It draws a single small hollow circle (~6 px radius,
  stroked) at the point returned by `AimPoint`.
- **Visibility**: gated on the existing `withPlayer` flag already threaded through
  world drawing, which is true only for the playing state. No new state check is
  introduced. It is not gated on player visibility: the crosshair stays solid
  while the ship blinks during invulnerability, because the blink is the
  invulnerability signal and should not be duplicated.
- **Colour**: a new palette entry alongside the existing colour variables, using
  the bullet cyan at reduced alpha (~140). Cyan ties the reticle to the projectile
  it predicts; the alpha keeps it subordinate to actual bullets.
- **No toggle**: no keybind, no CLI flag, no persisted preference. Chaos mode is
  a flag-free in-game choice by ADR 0001; adding a flag for a cosmetic aid would
  cut against that grain, and an always-on aid needs no discoverability text.
- **Commit**: a single conventional commit, e.g.
  `feat(game): add aiming crosshair ahead of the ship`.

## Testing Decisions

A good test here asserts observable behaviour — the position the game will draw —
and not how it is computed. It should not reach into rendering, take screenshots,
or assert on drawing calls; `game/` has no rendering tests today and this feature
does not justify introducing that machinery.

- **Module under test**: `Player`, via the new `AimPoint` method. This is the
  single seam; drawing remains untested, matching the existing treatment of the
  rest of the rendering code.
- **Prior art**: the existing player unit tests, which construct a player
  directly and assert on numeric state after driving it with synthetic controls.
  The new tests follow the same shape: construct, set heading, assert on returned
  coordinates with a float tolerance.
- **Cases**:
  - A table over cardinal headings (up, right, down, and the default heading) that
    the returned point is the expected offset from the ship's position.
  - An invariant that the distance from the ship to the returned point always
    equals the aim-distance constant, for a spread of headings.
  - A non-wrap case: a ship positioned near the right edge and pointing right
    returns an x greater than the screen width, rather than a wrapped small x.
    This is the regression guard for the deliberate divergence from the ship's
    own wrapping behaviour.
- No cluster or bridge involvement, so no mock provider is needed; per project
  convention any future `game/` test data would come from the mock provider only.

## Out of Scope

- Target acquisition of any kind: raycasting along the heading, snapping the
  reticle onto the first enemy it would hit, or highlighting that enemy.
- Lead indicators that compensate for ship or enemy motion.
- Drawing the full shot path (a dotted or faded line from nose to reticle).
- Making the crosshair toggleable by key, flag, or config, and any persistence of
  such a preference.
- Changing bullet behaviour, including making bullets wrap the screen or inherit
  ship velocity.
- Reticle feedback states: changing colour or shape when an enemy is under the
  crosshair, when the gun is on cooldown, or when the player is invulnerable.
- Crosshairs on any non-playing screen.
- An ADR: nothing here constrains future work.
- Documentation changes to `AGENTS.md` or `CONTEXT.md`; the existing description
  of the rendering file as "all rendering" remains accurate.

## Further Notes

The shot-path honesty point is the crux of this feature and the reason for the
non-wrap decision. The ship wraps around the screen; bullets do not — they are
deactivated the moment they cross a boundary. A crosshair that wrapped would draw
a reticle on the far side of the screen at a spot no bullet from that shot can
ever reach. Preserving that asymmetry is what makes the aid trustworthy, and the
non-wrap test exists specifically because "make the crosshair wrap like the ship"
looks like a consistency fix to anyone who has not read this note.
