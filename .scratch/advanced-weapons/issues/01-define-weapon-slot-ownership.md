# Define weapon slot ownership

Type: grilling
Status: resolved
Blocked by: None

## Question

What game-side model should own the main and secondary weapon slots, firing inputs, cooldown state,
and projectile creation so future weapons can replace either slot without introducing power-up,
inventory, or cluster concepts now?

## Answer

The Player owns exactly two Weapon slots: main and secondary. It routes Space to the main slot and
X to the secondary slot during its input update. Each equipped weapon owns its firing behavior and
cooldown state, and creates one or more projectiles from the Player's current pose when ready.

The Scene owns every active projectile after firing: movement, off-screen removal, collision
testing, applying impacts, and recording kills. The Player must not retain a weapon-specific
shoot timer or construct a concrete projectile directly. Weapons depend only on game-package
combat types and the firing pose; neither weapons nor the Player gain inventory, acquisition,
cluster, bridge, or session responsibilities.

This makes replacing either equipped weapon a Player-local composition change while leaving
world simulation and Pod consequences in their existing owners.
