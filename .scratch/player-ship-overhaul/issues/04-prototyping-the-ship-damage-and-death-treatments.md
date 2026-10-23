# Prototyping the ship's damage and death treatments

Type: prototype
Status: resolved
Blocked by:

## Question

How should the Interceptor look when it takes damage (Invulnerability) and while Dying and exploding?

Today the triangle simply blinks during Invulnerability (`playerVisible`), and Dying draws an expanding ring with no ship. Prototype treatments that suit the Interceptor chosen in "Prototyping the ship and engine glow". Options include a hull tint or flicker instead of a full blink, the engine glow sputtering on hit, and a break-apart or burst for the death explosion. Decide what happens to the engine glow in each case. Keep Invulnerability clearly readable as "you just got hurt", so it stays distinct from the Barrel roll.

## Answer

Hit treatment: **C (Damaged)**. Death treatment: **A (Break apart)**. Prototype: branch `prototype/ship-damage` (`game/prototype_hurt.go`, which builds on the ship, shots and roll prototypes; run `go run -tags prototype . --mock`, then K to hurt yourself and L to die).

- **Hit (during Invulnerability):** the ship is **never hidden**; the old blink is gone.
  - For the first 10 ticks after the hit, the ship shakes up to about 1.6 px, easing out.
  - A red-tinted copy of the hull (about 45% alpha) is drawn over it on alternate 6-tick phases, so the outline pulses red for the whole window.
  - For the first 45 ticks, six yellow-to-orange sparks spray from points on the hull.
  - Four dark smoke puffs trail behind the engines, fading as Invulnerability runs out.
  - The engine glow sputters, dropping out at random on about 30% of frames, for the first 45 ticks.
  - The Barrel roll still wins: while rolling, the plain roll animation is drawn with no hit treatment.
- **Death (the existing 60-tick Dying window, replacing the ring):** the Interceptor splits into four pieces: nose, rear hull, and each wing.
  - Each piece flies out from the centre along its own offset, easing out, travels 22–40 px plus some of the ship's last velocity, and spins at its own rate.
  - Pieces start orange-hot, cool to their hull colours, then char dark and fade out over the last part of the window.
  - An orange and yellow fireball covers the first 35%, and eight yellow embers scatter outward.
  - The engine glow is gone from the first tick.
- Implementation note: the tint and fade effects were done by drawing the ship into an offscreen image and scaling its colours. The real version may do the same, or tint the hull colours directly.
