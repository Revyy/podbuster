package game

import "math"

const (
	// sentinelTelegraphTicks is how long the ring spins up before a dash.
	sentinelTelegraphTicks = 30
	// sentinelFirstOffsetMax bounds the random extra delay before the first dash.
	sentinelFirstOffsetMax = 60
	sentinelDashSpeed      = 2.7
	sentinelDashOvershoot  = 40.0
	// sentinelRadialGain turns the pixel gap to the orbit into a radial pull
	// that is mixed with the tangent, capped so a far Sentinel mostly closes in.
	sentinelRadialGain = 0.1
	sentinelRadialCap  = 3.0
	// sentinelAimDistance is how far along its heading the Sentinel homes, far
	// enough that step never treats the aim point as already reached.
	sentinelAimDistance = 100.0
)

// sentinelStats are the per-tier orbit numbers.
type sentinelStats struct {
	radius   float64
	interval int
}

var sentinelTable = [...]sentinelStats{
	TierLight:    {radius: 200, interval: 240},
	TierStandard: {radius: 180, interval: 180},
	TierHeavy:    {radius: 150, interval: 120},
}

func (t Tier) sentinelStats() sentinelStats {
	return tierEntry(sentinelTable[:], t)
}

// sentinelCycle is a Sentinel's orbit direction and dash countdown. The zero
// value is unarmed: the next tick arms it with a full interval and no random
// offset, which is how a revert from Terminating restarts the cycle.
type sentinelCycle struct {
	armed bool
	dir   float64 // +1 or -1 around the player
	timer int     // ticks until the next dash starts
}

// initSentinel picks the orbit direction and first-dash offset from rng.
func (e *Enemy) initSentinel(rng RNG) {
	dir := 1.0
	if rng.IntN(2) == 1 {
		dir = -1
	}
	e.sentinel = sentinelCycle{
		armed: true,
		dir:   dir,
		timer: e.Tier.sentinelStats().interval + rng.IntN(sentinelFirstOffsetMax+1),
	}
}

// updateSentinel runs one tick of the Sentinel: orbit the player, telegraph,
// lock the player's position and dash through it, then rejoin the orbit. The
// countdown runs through the dash, so the interval is dash start to dash start.
func (e *Enemy) updateSentinel(targetX, targetY, steerX, steerY float64) {
	stats := e.Tier.sentinelStats()
	c := &e.sentinel
	if !c.armed {
		*c = sentinelCycle{armed: true, dir: orDefault(c.dir, 1), timer: stats.interval}
	}
	if e.lunge.active {
		e.step(targetX, targetY, steerX, steerY, 0)
		c.timer--
		return
	}

	dirX, dirY := e.orbitHeading(targetX, targetY, stats.radius)
	e.step(e.X+dirX*sentinelAimDistance, e.Y+dirY*sentinelAimDistance, steerX, steerY, e.CruiseSpeed())

	c.timer--
	if c.timer > sentinelTelegraphTicks {
		e.Telegraph = 0
		return
	}
	e.Telegraph = min(1, float64(sentinelTelegraphTicks-c.timer+1)/sentinelTelegraphTicks)
	// An enemy still outside the screen has no runway, so it keeps
	// telegraphing until it has circled in.
	if c.timer <= 0 && e.insideEdge() {
		e.startLunge(targetX, targetY, sentinelDashSpeed, sentinelDashOvershoot)
		e.Telegraph = 0
		c.timer = stats.interval
	}
}

func orDefault(v, def float64) float64 {
	if v == 0 {
		return def
	}
	return v
}

// orbitHeading is the unit direction that circles the player at the given
// radius: the tangent in the Sentinel's orbit direction, bent in or out in
// proportion to how far it is from the orbit.
func (e *Enemy) orbitHeading(targetX, targetY, radius float64) (float64, float64) {
	rx, ry := e.X-targetX, e.Y-targetY
	d := math.Hypot(rx, ry)
	if d < 0.0001 {
		return 1, 0
	}
	nx, ny := rx/d, ry/d
	tx, ty := -ny*e.sentinel.dir, nx*e.sentinel.dir
	pull := math.Max(-sentinelRadialCap, math.Min(sentinelRadialCap, (radius-d)*sentinelRadialGain))
	hx, hy := tx+nx*pull, ty+ny*pull
	mag := math.Hypot(hx, hy)
	return hx / mag, hy / mag
}
