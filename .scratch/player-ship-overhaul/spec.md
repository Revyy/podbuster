# Spec: Player ship overhaul

Status: ready-for-agent

Charted by the wayfinder map (`map.md`). Decision tickets 01–04 hold each decision in full, and this spec summarises them. Where the two disagree, the decision ticket wins.

## Problem Statement

The enemies now have detailed vector art, but the player ship is still three white lines and its shots are a 4×4 square and a short streak. Next to the droids the ship looks unfinished, and the player can't read their own speed or state from it. Play has also become hectic with Turrets, Sentinels and Darts, and the player has no way to escape a crowd or a Bolt ring except to outrun it.

## Solution

- **Ship:** the player flies the **Interceptor**, a vector-drawn fighter with swept wings and twin engine nozzles. Its blue engine glow grows with speed and flares under thrust.
- **Shots:** the Blaster fires **twin tracers**. Rockets look like **classic missiles**, with fire and smoke from the back. Shots that hit an enemy leave short **impact sparks**.
- **Barrel roll:** Z triggers a new ability, the **Barrel roll**. It is a 45-tick squash-flip that shows the ship's belly. While it lasts, enemies and Bolts pass through harmlessly, and it has a 3 s cooldown shown on the HUD.
- **Hits:** taking a hit no longer makes the ship blink out. It shakes, sparks, smokes and pulses red.
- **Death:** the ship breaks into spinning pieces with a fireball.

## User Stories

1. As a player, I want my ship to look like a real spaceship, so that it matches the detailed enemies.
2. As a player, I want the ship to point clearly along its heading, so that I always know where I'll fire.
3. As a player, I want a blue engine glow that grows with my speed, so that I can read my momentum at a glance.
4. As a player, I want the glow to flare while I hold thrust, so that thrusting feels different from coasting.
5. As a player, I want a faint glow even when stopped, so that the ship always looks powered.
6. As a player, I want Blaster shots to look like twin tracers, so that firing feels punchy.
7. As a player, I want Rockets to look like real missiles with fire and smoke, so that the heavy weapon feels heavy.
8. As a player, I want my shots to look nothing like enemy Bolts, so that I never confuse mine with theirs.
9. As a player, I want a brief muzzle flash when I fire, so that each shot registers.
10. As a player, I want a spark or fireball where my shot hits, so that hits feel satisfying.
11. As a player, I want to press Z to barrel roll, so that I can escape a hectic moment.
12. As a player, I want enemies and Bolts to pass through me harmlessly during a roll, so that it reliably saves me.
13. As a player, I want the roll not to move my ship, so that it's purely defensive and predictable.
14. As a player, I want to steer, thrust and fire while rolling, so that the roll never costs me control.
15. As a player, I want a 3 s cooldown, so that the roll is a decision, not a spam button.
16. As a player, I want a clear roll animation, so that I know when I'm safe.
17. As a player, I want a "Z Roll" cooldown bar next to my weapons, so that I know when it's ready.
18. As a player, I want the roll ready at the start of every session and retry, so that I'm never penalised.
19. As a player, I want no rolling while my ship is dying, so that death is final.
20. As a player, I want my ship to stay visible after a hit, so that I don't lose track of it in a crowd.
21. As a player, I want a hit to look like damage (shake, sparks, smoke, red pulse), so that I know I was hurt and am temporarily safe.
22. As a player, I want the roll animation to override the hit look, so that the two don't fight.
23. As a player, I want my ship to break apart when destroyed, so that death is dramatic.
24. As a developer, I want roll immunity and timing covered by scene tests, so that the ability stays reliable.

## Implementation Decisions

- **Ship** ([Prototyping the ship and engine glow](issues/01-prototyping-the-ship-and-engine-glow.md)):
  - Port the Interceptor from branch `prototype/player-ship`, without wingtip lights.
  - It is drawn rotated to the heading, needing a rotating pen alongside the droids' upright `pen`.
  - The 20×20 hitbox is unchanged and the art overhangs it slightly.
  - Engine glow: two layered blue plumes whose length, width and brightness come from the speed ratio, with a thrust flare and a slight flicker.
  - `Player` records whether thrust is held, rather than drawing code reading the keyboard.
- **Shots** ([Prototyping Blaster shots and Rockets](issues/02-prototyping-blaster-shots-and-rockets.md)):
  - Twin tracers and the Classic missile come from branch `prototype/player-shots`. Rendering only: hitboxes, speeds, spread and hit costs are unchanged.
  - Rocket smoke puffs stay put in the world and are derived from the projectile's age.
- **Impact sparks:** the Scene records an effect when a player projectile hits an enemy, much as Bolt bursts are recorded, and draw renders it for about 18 ticks. Effects don't touch combat. Retry, reset and session end clear them.
- **Barrel roll** ([Prototyping the Barrel roll and HUD indicator](issues/03-prototyping-the-barrel-roll-and-hud-indicator.md), and charting decisions in `map.md` Notes):
  - Z starts a 45-tick roll when the 180-tick cooldown is 0, and the cooldown starts at the same moment.
  - While rolling, Contact and Bolt resolution are skipped entirely: no damage, no knockback, enemy lunges are not ended, and Bolts stay live.
  - Steering, thrust and firing are unaffected.
  - `Player.Reset` clears the roll and cooldown. Dying doesn't roll.
  - Rolling is not Invulnerability: it neither starts nor extends the Invulnerability timer, which keeps counting down underneath.
- **Roll animation:**
  - The hull width scales by |cos 2πt|, with a 0.12 floor. The darker belly shows in the back half of the flip.
  - The ship scales up by 1 + 0.18·sin πt, a highlight sweeps the hull, and the engine plumes stay on.
  - The roll is drawn every frame, overriding the hit treatment.
- **Roll HUD:** a small panel below the weapon status panel showing `Z      Roll` and a 4 px cooldown bar. It uses the existing charging and ready colours and `cooldownFillRatio`.
- **Hit and death** ([Prototyping the ship's damage and death treatments](issues/04-prototyping-the-ship-damage-and-death-treatments.md)):
  - The Invulnerability blink is removed and replaced with: a 10-tick easing shake, a red-tinted hull pulse in 6-tick phases, 45 ticks of sparks and engine sputter, and smoke fading with the timer.
  - Death keeps the 60-tick Dying window. The ring is replaced by four pieces (nose, rear hull, two wings) flying apart and spinning, glowing hot then charring and fading. A fireball covers the first 35%, embers scatter, and the engine glow is gone.
- **Layering:** `game/` only. No new events and no bridge or cluster changes.

## Testing Decisions

- Test external behaviour at existing seams only. Don't test drawing code.
- **Player seam** (`updateWithControls` with fixed controls):
  - roll start only when the cooldown is ready;
  - roll and cooldown durations;
  - steering, thrust and firing during a roll;
  - `Reset` clearing the roll;
  - the thrust flag following the control.
- **Scene seam:**
  - a rolling player takes no Contact or Bolt damage and gets no knockback;
  - a Bolt passes through and stays live;
  - a lunge isn't ended by a roll overlap;
  - Invulnerability is untouched by rolling;
  - impact effects are recorded on projectile hits and cleared on reset.
- **Game seam:** no roll during Dying, and retry starts with the roll ready.
- Prior art: `player_test.go`, `bolt_test.go`, `combat_test.go` and `dart_test.go`.
- Every ticket ends with `go test ./...` green.

## Out of Scope

- New weapons, homing Rockets, or any weapon balance change.
- Sound.
- HUD redesign beyond the roll row.
- A player-ship or Barrel roll page in the Field guide.

## Further Notes

- All numbers are starting values to tune.
- Update AGENTS.md: the lunge rule (a roll overlap is not contact), the Barrel roll, the input keys, and the removal of the blink.
- `CONTEXT.md` already defines **Barrel roll**.

## Implementation

Completed: 2026-10-06
Branch: feat/player-ship-overhaul (local, no remote/PR)
Tickets resolved: 05-11 (12 pending manual playtest)
Notes: No tuning changes.
