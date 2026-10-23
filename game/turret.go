package game

import "math"

const (
	turretFirstOffsetMax = 60 // first volley waits one interval plus 0..this
	turretRingTelegraph  = 60
	turretAimedTelegraph = 15
	turretAimedDelay     = 60 // Heavy: ticks from ring to aimed Bolt
	turretRotation       = math.Pi / 6
	// turretSpawnInset keeps a Turret's spawn well inside the screen: it never
	// moves, so it cannot enter from the edge as other enemies do.
	turretSpawnInset = 60.0
	// turretSpawnClearance keeps new Turrets off the player's starting spot.
	turretSpawnClearance = 120.0
)

// RNG is the source of randomness a Scene draws from. *rand.Rand satisfies it.
type RNG interface {
	IntN(n int) int
	Float64() float64
}

type turretPattern struct {
	bolts    int
	interval int
	rotates  bool
	aimed    bool
}

func turretPatternOf(t Tier) turretPattern {
	switch t {
	case TierLight:
		return turretPattern{bolts: 4, interval: 376}
	case TierHeavy:
		return turretPattern{bolts: 6, interval: 300, rotates: true, aimed: true}
	default:
		return turretPattern{bolts: 6, interval: 300, rotates: true}
	}
}

// turretCycle is a Turret's firing state. The zero value is unarmed: the next
// tick arms it with a fresh first-volley delay, which is also how a practice
// respawn or a revert from Terminating restarts the cycle.
type turretCycle struct {
	armed   bool
	timer   int // ticks until the next ring
	aimedIn int // ticks until the aimed Bolt, 0 when none is pending
	volleys int
}

// updateTurrets advances every active Turret's cycle and fires due Bolts.
// Terminating and Exploding Turrets do nothing.
func (s *Scene) updateTurrets(playerX, playerY float64) {
	for _, e := range s.enemies {
		if e.Archetype != ArchetypeTurret || e.State != EnemyActive {
			continue
		}
		s.stepTurret(e, playerX, playerY)
	}
}

func (s *Scene) stepTurret(e *Enemy, playerX, playerY float64) {
	pat := turretPatternOf(e.Tier)
	c := &e.turret
	if !c.armed {
		*c = turretCycle{armed: true, timer: pat.interval + s.rng.IntN(turretFirstOffsetMax+1)}
	}

	if c.aimedIn > 0 {
		c.aimedIn--
		if c.aimedIn == 0 {
			angle := math.Atan2(playerY-e.Y, playerX-e.X)
			s.bolts = append(s.bolts, NewBolt(e.X, e.Y, angle, e.Archetype))
		}
	}

	c.timer--
	if c.timer <= 0 {
		angle0 := 0.0
		if pat.rotates {
			angle0 = float64(c.volleys) * turretRotation
		}
		for i := 0; i < pat.bolts; i++ {
			angle := angle0 + 2*math.Pi*float64(i)/float64(pat.bolts)
			s.bolts = append(s.bolts, NewBolt(e.X, e.Y, angle, e.Archetype))
		}
		c.volleys++
		c.timer = pat.interval
		if pat.aimed {
			c.aimedIn = turretAimedDelay
		}
	}

	e.Telegraph = math.Max(
		windUp(c.timer, turretRingTelegraph),
		windUp(c.aimedIn, turretAimedTelegraph),
	)
}

// windUp is 0..1 progress through the last window ticks before an event,
// where remaining is the ticks left until it. A pending value of 0 means none.
func windUp(remaining, window int) float64 {
	if remaining <= 0 || remaining > window {
		return 0
	}
	return float64(window-remaining+1) / float64(window)
}

// turretSpawnPosition picks an on-screen spot for a Turret, clear of the
// player's start and of other Turrets.
func (s *Scene) turretSpawnPosition() (float64, float64) {
	var x, y float64
	for range 20 {
		x = turretSpawnInset + s.rng.Float64()*(screenWidth-2*turretSpawnInset)
		y = turretSpawnInset + s.rng.Float64()*(screenHeight-2*turretSpawnInset)
		if math.Hypot(x-screenWidth/2, y-screenHeight/2) < turretSpawnClearance {
			continue
		}
		clear := true
		for _, o := range s.enemies {
			if o.Archetype == ArchetypeTurret && math.Hypot(x-o.X, y-o.Y) < pairSeparation(o, o) {
				clear = false
				break
			}
		}
		if clear {
			break
		}
	}
	return x, y
}
