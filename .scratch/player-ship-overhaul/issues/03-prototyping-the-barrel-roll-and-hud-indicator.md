# Prototyping the Barrel roll and HUD indicator

Type: prototype
Status: resolved
Blocked by: 01

## Question

How should the Barrel roll animate on the new ship, how should its cooldown show on the HUD, and do the charted timings feel right?

Prototype the roll on the ship chosen in "Prototyping the ship and engine glow". It's top-down, so the roll has to read through things like hull width squash, belly/top shading flip and engine glow behaviour. Use Z to trigger it, a 240-tick cooldown and about a 30-tick roll with immunity throughout, and tune these if play says otherwise. During the roll the ship has no Contact or Bolt damage and no push. Add a roll cooldown indicator next to the existing weapon status HUD. Confirm the roll needs no Invulnerability flashing on top to read as safe.

## Answer

Animation **A (Squash flip)**, HUD **1 (Weapon panel row)**, and the timings retuned in play to a **45-tick roll** and a **180-tick (3 s) cooldown**. The roll animation wins over the Invulnerability blink. Prototype: branch `prototype/barrel-roll` (`game/prototype_roll.go`, which builds on the ship and shots prototypes; run `go run -tags prototype . --mock`).

- **Animation:** over roll progress t (0→1), the hull's width is scaled by |cos(2πt)|, with a floor of 0.12 so it never vanishes. When cos < 0 the **belly** shows: the same silhouette in darker gunmetal with panel lines and no cockpit.
  - The whole ship scales up by 1 + 0.18·sin(πt) to suggest lift.
  - A white highlight sweeps the top of the hull as it banks.
  - The engine plumes keep drawing throughout.
  - There are no trails, afterimages or shimmer: variants B and C were rejected.
- **Immunity:** while rolling, Contact and Bolt resolution are skipped entirely. There's no damage, no knockback, the enemy's lunge isn't ended, and Bolts stay live and pass through. The prototype did this with a `RollTicks` gate at the top of `resolveContact` and `resolveBolts`.
- **Invulnerability overlap:** while the ship is rolling it is always drawn with the roll animation and never hidden by the blink. The Invulnerability timer keeps counting down underneath, and blinking resumes after the roll if any time is left.
- **HUD:** a separate small panel directly below the weapon status panel (prototype: x 8, y 552, 170×26) with the label `Z      Roll` and a 4 px cooldown bar. It uses the existing charging and ready colours and `cooldownFillRatio`, matching the Rockets row.
- **Rules carried over from charting:** Z starts the roll when the cooldown is 0; the cooldown starts when the roll begins. Retry and a new session start ready, and there's no rolling while Dying.
