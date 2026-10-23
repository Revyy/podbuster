# 01: Draw an aiming crosshair ahead of the ship

**What to build:** While the player is flying a session, a small faint hollow
cyan reticle sits a fixed distance ahead of the ship along its heading. It tracks
the nose continuously as the player turns, so the player can line up a shot
before firing instead of learning where the gun points from where the bullets
went. It reports the ship's aim and nothing else: no target snapping, no lead
compensation, no reaction to what is underneath it.

Because bullets are deactivated at the screen edge while the ship wraps, the
reticle must not wrap. When the player points at a nearby edge, the reticle
simply slides off that edge and is clipped, so it never marks a spot no bullet
from that shot could reach.

The reticle exists exactly when the player can shoot: present in the playing
state in both practice and chaos mode, and immediately present again after a
retry; absent on the title screen, the mode-select screen, during the death
explosion, and on the game-over screen. It stays solid while the ship flashes
after contact damage, so the blink keeps meaning "invulnerable" and only that.

Spec: `.scratch/aiming-crosshair/spec.md`

**Blocked by:** None (can start immediately)

**Status:** resolved

- [x] The player exposes an aim-point method taking no arguments, returning the
      reticle's world position measured from the ship's centre along its heading
- [x] The aim distance is a single named package constant declared with the other
      player tuning constants, around 120 px
- [x] The aim point is not screen-wrapped: for a ship near an edge pointing
      outward it returns a coordinate beyond the screen bounds
- [x] A unit test tables cardinal headings and asserts the expected aim point
      within a float tolerance
- [x] A unit test asserts the distance from ship to aim point equals the aim
      constant across a spread of headings
- [x] A unit test pins the non-wrap behaviour as a regression guard against a
      future "consistency fix" that wraps it like the ship
- [x] A crosshair draw helper renders a single small hollow stroked circle
      (~6 px radius) at the aim point
- [x] The crosshair colour is a new palette entry beside the existing colour
      variables: bullet cyan at reduced alpha (~140)
- [x] The crosshair is drawn only when the existing world-drawing player flag is
      set, with no new game-state check introduced
- [x] The crosshair is not gated on the ship's blink visibility, so it stays
      solid during invulnerability
- [x] No toggle is added: no keybind, no CLI flag, no persisted preference
- [x] All changes stay within the game package; no cluster or bridge imports,
      no new event types or channels
- [x] `go test ./...` passes
- [x] Landed as one conventional commit, e.g.
      `feat(game): add aiming crosshair ahead of the ship`

<!-- Implementing skills set Status to resolved and add a Completion section after the final commit. -->

## Completion

Completed: 2026-08-28
Commit: 5adab0a
Notes: Added a tested, non-wrapping cyan aiming crosshair ahead of the player during active play.
