# Escalating archetypes by tier

Type: grilling
Status: resolved
Blocked by: 05

## Question

Beyond the shared tier levers (Hits, size, speed, contact damage), what does each archetype gain per tier, Light / Standard / Heavy? Examples:
- Turret: fixed radial → aimed → faster.
- Sentinel: orbit tightness and dash frequency.
- Dart: charge speed and overshoot.
- Probe: likely the shared levers only.

Does any archetype need its own tier behaviour at all, or is the shared scaling enough for it?

The Turret's Standard pattern and Bolt rules are fixed by "Specifying enemy fire" (6 Bolts, every 120 ticks, rotating 30°, 1 Health per hit at every tier). Its recommendation: only the Turret fires.

## Answer

These are starting values, to be tuned. Speeds are in px/tick, durations in ticks, and the player's max speed is 4.

- **Base stats per archetype**: each archetype has its own base stats, and tier scales them. This revises the "same for every archetype" table in [Defining tier banding](04-defining-tier-banding.md):
  - **Hits**: Probe, Turret and Sentinel take 2/3/5; Dart takes 1/2/3 (Light/Standard/Heavy).
  - **Size and contact damage**: shared, as in tier banding.
  - **Speed lever** (1.1×/1×/0.85×): scales **cruising only**, meaning Probe homing, Sentinel orbit and Dart drift. It doesn't affect dash or charge speeds, which are set per archetype and tier below. The stationary Turret ignores it.
- **Who fires**: only the Turret fires Bolts, at every tier.
- **Probe**: shared levers only.
- **Turret**: Bolt speed and cost are the same at every tier.
  - Light: a 4-way ring every 150 ticks with no rotation.
  - Standard: as in [Specifying enemy fire](05-specifying-enemy-fire.md) (6-way, every 120 ticks, rotating 30°, 30-tick telegraph).
  - Heavy: the Standard ring, plus one aimed Bolt at the player 60 ticks after each ring, with its own 15-tick telegraph.
- **Sentinel**:
  - **Standard**: cruises to a 180 px orbit around the player at 0.8 tangential, in a direction picked at random at spawn. Every 180 ticks (first cycle offset by a random 0–60) it gives a 30-tick telegraph (the ring spins up and brightens). It then locks the player's position and dashes at 2.7 in a straight line to 40 px past that point, then rejoins its orbit. Several Sentinels share a radius and rely on Separation.
  - **Per tier**: Light has a 200 px radius and dashes every 240 ticks; Standard 180 px every 180; Heavy 150 px every 120. Dash speed is 2.7 at every tier.
- **Dart**:
  - **Standard cycle**: a 40-tick wind-up (tentacles flared, eye locked, drifting at most 0.8), then it locks the player's position and charges at 3.5 straight through that point, overshooting by 120 px. It then brakes for 20 ticks and repeats. A charge that would leave the screen stops at the edge, inset by half the Dart's size, because enemies don't wrap.
  - **Per tier** (charge speed / overshoot / wind-up): Light 3.0 / 80 / 50; Standard 3.5 / 120 / 40; Heavy 4.0 / 160 / 30. Charges never exceed the player's max speed, so a Dart can always be outrun.
- **Lunges** (Sentinel dashes and Dart charges):
  - Contact during a lunge follows the normal contact rules (tier contact damage, knockback, shared Invulnerability) and ends the lunge immediately. Knockback then takes over, followed by the brake or return.
  - Lunges skip soft avoidance steering but keep the hard no-overlap floor. If the floor blocks a lunge, the lunge ends early.
- **Timers**: a practice respawn starts its timers fresh, with a new random offset and orbit direction, and keeps its tier. A chaos-mode enemy that reverts from Terminating resumes cruising with its cycle restarted, so it never fires or lunges the instant it revives.

## Comments

- 2026-10-01: The Sentinel's orbit speed (1.2 → 0.8) and dash speed (4 → 2.7) were lowered by a third after playtesting in [Prototyping archetype silhouettes](07-prototyping-archetype-silhouettes.md).
