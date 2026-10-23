# Prototyping Blaster shots and Rockets

Type: prototype
Status: resolved
Blocked by:

## Question

What should Blaster shots and Rockets look like?

Prototype new visuals for both player Projectiles, changing rendering only. Hitboxes, speeds, spread and hit costs stay the same. Blaster shots replace today's 4×4 square. Rockets should look like real missiles, with a body, nose and fins, and with fire bursting out of the back. Their muzzle flash can be restyled. Both must stay clearly distinct from enemy Bolts.

## Answer

Blaster: **C (Twin tracers)**. Rockets: **A (Classic missile)**. Impact sparks stay in scope. Prototype: branch `prototype/player-shots` (`game/prototype_shots.go`, which builds on `prototype/player-ship`; run `go run -tags prototype . --mock`).

- **Blaster, Twin tracers:** two thin parallel tracers about 2.5 px either side of the flight line, converging slightly toward the nose.
  - Each has a translucent blue glow (width 3), a cyan line (width 1.3) and a white-blue tip dot.
  - The tail grows from launch up to about 10 px.
  - A small cyan muzzle flash with a cross glint shows for the first 3 ticks.
  - It's still **one projectile with the unchanged 8 px hitbox**. The twin look is only drawn that way.
- **Rocket, Classic missile:** a white-grey body about 10 px long and 3.2 px wide with a grey band, a red nose cone, and two dark-red swept tail fins.
  - A flickering fire cone about 7 px long comes out of the back, layered red → orange → yellow with a white-hot core at the nozzle.
  - It leaves grey smoke puffs that hang in place in the world and grow and fade over about 18 ticks.
  - An orange and white muzzle burst shows for the first 4 ticks.
  - The projectile position is the nose tip, as today.
- **Impact sparks (new, kept in scope):** a shot that disappears on hitting an enemy leaves a short effect of about 18 ticks.
  - Blaster: a small cyan flash with five white sparks spraying backward.
  - Rocket: an orange and yellow fireball with an expanding grey smoke ring.
  - The prototype inferred impacts by watching for vanished projectiles. The real version should record an impact effect in the Scene when a projectile hits, much as Bolt bursts are recorded, rather than inferring it in draw.
- Both shots read clearly apart from the round enemy Bolts.
