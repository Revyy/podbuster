# Prototyping the Probe droid

Type: prototype
Status: resolved
Blocked by:

## Question

What should the baseline Probe enemy look like, and is Ebitengine vector drawing good enough or do we need sprites?

Prototype a probe-droid enemy styled on `Probe_droid.png` (repo root): dome and eye on top, a body, and dangling arms. It is drawn upright whatever its heading, sized around the 24×24 hitbox with some overhang allowed. Include an idle animation (e.g. arm sway, eye pulse) and droid treatments for Terminating, Exploding, projectile damage, and contact flash, replacing today's recolouring. Build it from parts that later archetypes and tiers could vary.

## Answer

Variant **C (Seeker)** is the baseline Probe. Vector drawing is good enough, so no sprite pipeline is needed.

- Look: an upright hovering orb split into hemispheres (lit dome on top, dark lower half with a trim band), a short antenna, and one large eye that tracks the player. Four tentacles hang from underneath and stream behind the direction of travel with a travelling wave. It overhangs the 24×24 hitbox a little.
- Idle: a gentle bob, the tentacle wave, and the eye following the player.
- Damage: tints toward the damage colour, shakes sideways, and the eye flashes solid white.
- Contact: the existing flash tint.
- Terminating: greyscale pulse; the droid sinks slightly, the tentacles hang straight down and still, and the eye closes to a slit.
- Exploding: the hemispheres split apart, the tentacles splay, the eye pops upward, and an expanding cyan ring fades out.
- For later archetypes and tiers, the hull colours, eye, tentacle count and length, and body shape are the obvious parts to vary.
- Rejected: A (Classic, filled reference droid) and B (Wireframe neon outline).

Prototype (primary source): branch `prototype/probe-droid`, file `game/prototype_probe.go` (function `drawProbeSeeker`). Run it with `go run -tags prototype . --mock`, press Tab to reach variant C, and G for the state gallery.
