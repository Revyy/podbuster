# Prototyping the ship and engine glow

Type: prototype
Status: resolved
Blocked by:

## Question

What should the player ship look like, and how should its engine glow respond to speed and thrust?

Prototype a vector-drawn ship that reads as a real spaceship and points along its heading. It replaces today's three-line triangle, keeps the 20×20 hitbox, and may overhang it a little. It should sit stylistically alongside the droid enemies (`game/droid.go` pen helpers). Add a blue-ish engine glow at the back: a small idle glow scaled by current speed, plus a visible flare while thrust is held. Compare a few variants in play against the busy enemy field.

## Answer

Variant **A (Interceptor)** is the player ship, with the engine glow as prototyped. Prototype: branch `prototype/player-ship` (`game/prototype_ship.go`, run `go run -tags prototype . --mock`).

- Look: drawn top-down and rotated to the heading (nose along +x), using a rotating pen that's separate from the droids' upright `pen`. It is roughly 26 long by 24 wide, so it overhangs the 20×20 hitbox a little at the wingtips.
  - Steel hull with a lighter spine, darker swept wings with steel inner panels, and two dark engine nozzles.
  - Dark cockpit canopy with a cyan glint.
  - **No wingtip running lights**: they were removed after review.
- Engine glow: two blue plumes, one per nozzle, each drawn as layered translucent triangles (outer blue, mid light blue, white-blue core) with a soft nozzle bloom.
  - Length is about 2.5 + 9 × speed ratio and width grows about 35% with speed. Brightness rises from 0.45 at rest to 0.8 at max speed, so there's a small glow even when idle.
  - Holding thrust adds about 7 px of length, makes the plume about 25% wider and sets full brightness.
  - A slight sine flicker runs at all times, and the two plumes flicker out of phase.
  - The user confirmed this feels right.
- Implementation note: `Player` doesn't store whether thrust is held, though the glow needs it. The prototype read the keyboard directly in draw. The real version should record the thrust control on `Player` during update.
