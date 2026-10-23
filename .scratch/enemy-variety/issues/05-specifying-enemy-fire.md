# Specifying enemy fire

Type: grilling
Status: resolved
Blocked by:

## Question

How do enemy shots work, given the Turret (and possibly harder tiers of other archetypes) fires at the player? Decide:
- the glossary term, since **Projectile** is currently defined as player-weapon-only
- what a hit costs the player and how it interacts with Health and Invulnerability
- whether shots are blocked by other enemies or by player projectiles
- shot speed, lifetime, and visual language relative to player projectiles
- what the Turret's base fixed radial pattern looks like (count, rate, rotation)
- whether Terminating or Exploding enemies' in-flight shots persist

## Answer

- **Term**: enemy shots are **Bolts** (added to `CONTEXT.md`). **Projectile** stays player-only. Health, Invulnerability and Terminating now mention Bolts.
- **Cost**: 1 Health at every tier; tiers escalate the pattern instead. Bolts share the contact Invulnerability window both ways, and knock the player back about half as hard as contact. Feedback is the existing ship flash plus a small burst in the Bolt's hue at the impact point. No new HUD.
- **Invulnerable player**: a Bolt passes through and stays live, so Invulnerability can't be used to clear lanes.
- **Blocking**: Bolts pass through enemies and can't be shot down. Dodging is the only counterplay.
- **Flight**: 2.4 px/tick, slower than the player's max of 4 (tuned down from 3 after the manual check in ticket 22: several Turrets made 3 too hard to navigate). Bolts expire at the screen edge and don't wrap.
- **Look**: a round glowing pellet about 6 px in the firing archetype's hue, distinct from player shots. The hitbox is slightly smaller than the visual.
- **Persistence**: in-flight Bolts don't depend on the enemy that fired them and survive its Terminating and Exploding states. Terminating enemies stop firing. Session end, scene reset and retry clear all Bolts.
- **Dying / Game Over**: no new volleys. In-flight Bolts keep moving but can't hit, then expire.
- **Who fires**: Bolts are defined generically. Recommendation for "Escalating archetypes by tier": only the Turret fires, so archetypes stay distinct.
- **Turret base pattern (Standard)**:
  - Volley: 6 Bolts evenly spaced (60°), every 120 ticks. Each volley rotates 30°.
  - Telegraph: the eye/barrel glows brighter for 30 ticks before each volley.
  - First volley: one full interval plus a random 0–60 tick offset after spawn or practice respawn, so Turrets don't fire in lockstep. The offset comes from an injectable RNG on the scene, so tests stay deterministic.
  - No range check: the Turret fires regardless of where the player is. Bolts don't wrap, so the seam is naturally safe.
- **Turret contact**: the Turret is immovable. Contact pushes only the player. Separation still applies, but other enemies give way and the Turret never moves.
- These values are starting points to tune.
