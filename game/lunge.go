package game

import "math"

// lunge is a straight-line rush at a locked point: a Dart charge or a Sentinel
// dash. While active it replaces homing, ignores soft avoidance and knockback,
// and ends when its distance is spent, on contact with the player, or when the
// hard no-overlap floor pushes the enemy.
type lunge struct {
	active    bool
	dirX      float64
	dirY      float64
	speed     float64
	remaining float64
}

// startLunge locks (targetX, targetY) and rushes through it, overshooting by
// overshoot pixels. The run is cut short at the screen edge, inset by half the
// enemy's size, because enemies don't wrap.
func (e *Enemy) startLunge(targetX, targetY, speed, overshoot float64) {
	dx, dy := targetX-e.X, targetY-e.Y
	dist := math.Hypot(dx, dy)
	if dist < 0.0001 {
		dx, dy, dist = 1, 0, 1
	}
	dirX, dirY := dx/dist, dy/dist
	length := math.Min(dist+overshoot, e.edgeRunway(dirX, dirY))
	e.lunge = lunge{active: length > 0, dirX: dirX, dirY: dirY, speed: speed, remaining: length}
}

// edgeRunway is how far the enemy can travel along (dirX, dirY) before its
// centre reaches the inset screen edge. An enemy already outside has none.
func (e *Enemy) edgeRunway(dirX, dirY float64) float64 {
	inset := e.Size() / 2
	run := math.Inf(1)
	axis := func(pos, dir, max float64) {
		switch {
		case dir > 1e-9:
			run = math.Min(run, (max-inset-pos)/dir)
		case dir < -1e-9:
			run = math.Min(run, (inset-pos)/dir)
		}
	}
	axis(e.X, dirX, screenWidth)
	axis(e.Y, dirY, screenHeight)
	return math.Max(0, run)
}

// advanceLunge moves one tick along the lunge and ends it when spent.
func (e *Enemy) advanceLunge() {
	step := math.Min(e.lunge.speed, e.lunge.remaining)
	e.X += e.lunge.dirX * step
	e.Y += e.lunge.dirY * step
	e.lunge.remaining -= step
	if e.lunge.remaining <= 0 {
		e.lunge.active = false
	}
}

// endLunge stops any lunge in progress.
func (e *Enemy) endLunge() { e.lunge.active = false }

// Lunging reports whether the enemy is mid-lunge.
func (e *Enemy) Lunging() bool { return e.lunge.active }
