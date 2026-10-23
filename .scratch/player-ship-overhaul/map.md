# Player ship overhaul

Label: wayfinder:map

## Destination

A spec, ready to hand off as implementation tickets, covering a new player ship look with a speed-driven engine glow, new Blaster shot and Rocket visuals, and the **Barrel roll** ability with its animation and HUD cooldown indicator.

## Notes

- Domain: game feel and visuals. Use the `CONTEXT.md` terms Barrel roll, Invulnerability, Projectile, Bolt and Contact. The weapons are Blaster and Rockets.
- Skills: `prototype` for visuals, and `grilling` + `domain-modeling` for decisions.
- Plan, don't do. Hand off once the spec is clear.
- Standing decisions from charting:
  - The ship stays vector-drawn like the droids (no sprites). It keeps its 20×20 hitbox, and the art may overhang it a little.
  - Engine glow: blue-ish, with a small idle glow that grows with speed plus a flare while thrusting.
  - Blaster shots and Rockets get new looks only (plus impact sparks, added by [Prototyping Blaster shots and Rockets](issues/02-prototyping-blaster-shots-and-rockets.md)). Rockets keep their name, spread and behaviour (no homing). They should read as real missiles with fire coming out of the back.
  - Barrel roll: Z key, with immunity throughout the roll. Timings were retuned to a 45-tick roll and a 180-tick cooldown by [Prototyping the Barrel roll and HUD indicator](issues/03-prototyping-the-barrel-roll-and-hud-indicator.md). It is purely defensive and doesn't move the ship. Enemies and Bolts pass through with no damage and no push. Steering, thrust and firing all work during the roll. Retry resets the cooldown, and there's no rolling while Dying. It is separate from Invulnerability and doesn't trigger its flashing.
  - The HUD gets a roll cooldown indicator.
  - `game/` only. No bridge or cluster changes.

## Decisions so far

- [Prototyping the ship and engine glow](issues/01-prototyping-the-ship-and-engine-glow.md): The ship is the vector "Interceptor", with swept wings, twin nozzles and no wingtip lights. Each nozzle has a blue plume that grows with speed and flares while thrusting.
- [Prototyping Blaster shots and Rockets](issues/02-prototyping-blaster-shots-and-rockets.md): The Blaster fires twin tracers (still one projectile and one hitbox). Rockets become classic missiles with a flame and smoke puffs, and hits leave short impact sparks.
- [Prototyping the Barrel roll and HUD indicator](issues/03-prototyping-the-barrel-roll-and-hud-indicator.md): The ship does a squash-flip roll that shows its belly, lasting 45 ticks with a 180-tick cooldown. During the roll Contact and Bolts are skipped and the Invulnerability blink is suppressed. A "Z Roll" row sits below the weapon panel.
- [Prototyping the ship's damage and death treatments](issues/04-prototyping-the-ship-damage-and-death-treatments.md): When hit, the ship stays visible and shakes, sparks, smokes and pulses red, with its engine sputtering, instead of blinking. On death it breaks into four spinning pieces with a fireball and embers.

## Handoff

Destination reached: [spec.md](spec.md), with implementation tickets 05–12 in `issues/`.

## Not yet specified

_(Nothing remains here. The roll-over-Invulnerability question was settled by "Prototyping the Barrel roll and HUD indicator".)_

## Out of scope

- New weapons, homing Rockets, or any weapon balance change.
- Sound.
- HUD redesign beyond the roll cooldown indicator.
- A player-ship or Barrel roll page in the Field guide.
