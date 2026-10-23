# Tune the rocket volley

Type: prototype
Status: resolved
Blocked by: None

## Question

What rocket speed, spacing across the ship, projectile size, cooldown duration, and visual treatment
make the three parallel rockets readable and useful without replacing the main weapon?

## Answer

The secondary weapon stays intentionally narrower than the main weapon. X fires a three-rocket
volley with a center shot plus two offset shots at roughly ±8° to the ship's heading, starting a
little left and right of the nose so the pattern reads as a single burst rather than a wide fan.
Each rocket is a compact 10×10 projectile with a slightly faster velocity than the standard bullet
(9.0 px/tick instead of 8.0 px/tick), which gives the volley enough reach to matter without making
it feel like a replacement for the primary fire. The volley uses a 180-tick cooldown, so the player
gets a burst of three shots every 3 s, while the main weapon keeps the existing 15-tick Space
shot cadence. The rockets should render as short orange-auburn bolts with a tiny bright exhaust
trail and a minimal muzzle flash; they stay solid, readable, and easy to track without any
homing/splash/pierce behavior.

This keeps the secondary weapon useful as a burst-damage support tool, while the current main
weapon remains the default attack and the main feel of the game is not displaced.
