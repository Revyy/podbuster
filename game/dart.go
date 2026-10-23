package game

// dartBrakeTicks is how long a Dart sits braking after a charge.
const dartBrakeTicks = 20

// dartDriftSpeed caps a Dart's drift while winding up.
const dartDriftSpeed = 0.8

type dartPhase int

const (
	dartWindup dartPhase = iota
	dartCharge
	dartBrake
)

// dartCycle tracks where a Dart is in wind-up, charge, brake.
type dartCycle struct {
	phase dartPhase
	timer int
}

// dartStats are the per-tier charge numbers.
type dartStats struct {
	chargeSpeed float64
	overshoot   float64
	windup      int
}

var dartTable = [...]dartStats{
	TierLight:    {chargeSpeed: 3.0, overshoot: 80, windup: 50},
	TierStandard: {chargeSpeed: 3.5, overshoot: 120, windup: 40},
	TierHeavy:    {chargeSpeed: 4.0, overshoot: 160, windup: 30},
}

func (t Tier) dartStats() dartStats {
	return tierEntry(dartTable[:], t)
}

// updateDart runs one tick of the Dart cycle: wind up (drifting, eye glowing),
// lock the player's position and charge through it, then brake.
func (e *Enemy) updateDart(targetX, targetY, steerX, steerY float64) {
	stats := e.Tier.dartStats()
	switch e.dart.phase {
	case dartWindup:
		drift := min(e.CruiseSpeed(), dartDriftSpeed)
		e.step(targetX, targetY, steerX, steerY, drift)
		e.dart.timer++
		e.Telegraph = min(1, float64(e.dart.timer)/float64(stats.windup))
		// An enemy still outside the screen has no runway, so it keeps
		// winding up until it has drifted in.
		if e.dart.timer >= stats.windup && e.insideEdge() {
			e.startLunge(targetX, targetY, stats.chargeSpeed, stats.overshoot)
			e.Telegraph = 0
			e.dart = dartCycle{phase: dartCharge}
		}
	case dartCharge:
		e.step(targetX, targetY, steerX, steerY, 0)
		if !e.lunge.active {
			e.dart = dartCycle{phase: dartBrake}
		}
	case dartBrake:
		e.step(targetX, targetY, steerX, steerY, 0)
		e.dart.timer++
		if e.dart.timer >= dartBrakeTicks {
			e.dart = dartCycle{}
		}
	}
}

// insideEdge reports whether the enemy's centre is within the inset screen.
func (e *Enemy) insideEdge() bool {
	inset := e.Size() / 2
	return e.X >= inset && e.X <= screenWidth-inset && e.Y >= inset && e.Y <= screenHeight-inset
}
