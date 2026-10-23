package game

import "math"

const (
	// boltSpeed is pixels per tick. It is slower than the player's max speed, so
	// a Bolt can always be outrun.
	boltSpeed = 2.4
	// boltSize is the visual diameter; the hitbox is slightly smaller.
	boltSize    = 6.0
	boltHitSize = 5.0
	// boltKnockbackScale is the share of contact knockback a Bolt impact applies.
	boltKnockbackScale = 0.5
	// boltDamage is the Health a Bolt costs at every tier.
	boltDamage = 1
	// burstTicks is how long an impact burst lasts.
	burstTicks = 14
)

// Bolt is a shot fired by an enemy. It costs the player Health, passes through
// enemies and cannot be shot down. It does not depend on the enemy that fired
// it, so it outlives that enemy.
type Bolt struct {
	X, Y   float64
	VX, VY float64
	Active bool
	// Archetype is the firing archetype and sets the Bolt's colour.
	Archetype Archetype
}

// NewBolt creates a Bolt at (x, y) travelling at boltSpeed along angle radians.
func NewBolt(x, y, angle float64, archetype Archetype) *Bolt {
	return &Bolt{
		X: x, Y: y,
		VX:        math.Cos(angle) * boltSpeed,
		VY:        math.Sin(angle) * boltSpeed,
		Active:    true,
		Archetype: archetype,
	}
}

// Update moves the Bolt and deactivates it once it leaves the screen. Bolts
// never wrap.
func (b *Bolt) Update() {
	b.X += b.VX
	b.Y += b.VY
	if b.X < 0 || b.X > screenWidth || b.Y < 0 || b.Y > screenHeight {
		b.Active = false
	}
}

// Bounds returns the Bolt's collision rectangle.
func (b *Bolt) Bounds() Rect {
	half := boltHitSize / 2
	return Rect{X: b.X - half, Y: b.Y - half, W: boltHitSize, H: boltHitSize}
}

// ImpactBurst is the short flash where a Bolt hit the player, in the firing
// archetype's colour.
type ImpactBurst struct {
	X, Y      float64
	Archetype Archetype
	Age       int
}

// Bolts returns the Bolts currently in flight.
func (s *Scene) Bolts() []*Bolt { return s.bolts }

// Bursts returns the impact bursts currently showing.
func (s *Scene) Bursts() []ImpactBurst { return s.bursts }

// ClearBolts removes every Bolt and short-lived impact effect. Retry, reset and
// session end call it so no transient combat state carries over.
func (s *Scene) ClearBolts() {
	s.bolts = nil
	s.bursts = nil
	s.clearProjectileImpacts()
}

// updateBolts moves Bolts, expires those off-screen and ages bursts.
func (s *Scene) updateBolts() {
	live := s.bolts[:0]
	for _, b := range s.bolts {
		b.Update()
		if b.Active {
			live = append(live, b)
		}
	}
	s.bolts = live

	bursts := s.bursts[:0]
	for _, b := range s.bursts {
		b.Age++
		if b.Age < burstTicks {
			bursts = append(bursts, b)
		}
	}
	s.bursts = bursts
}

// resolveBolts applies Bolt hits to the player. During Invulnerability a Bolt
// passes through and stays live, so it cannot be used to clear lanes.
func (s *Scene) resolveBolts(player *Player) {
	if player.IsRolling() {
		return
	}
	if player.IsInvulnerable() {
		return
	}
	playerBounds := player.Bounds()
	for _, b := range s.bolts {
		if !b.Active || !playerBounds.Overlaps(b.Bounds()) {
			continue
		}
		nx, ny := b.VX/boltSpeed, b.VY/boltSpeed
		if player.TakeBoltDamage(boltDamage, nx, ny) {
			b.Active = false
			s.bursts = append(s.bursts, ImpactBurst{X: b.X, Y: b.Y, Archetype: b.Archetype})
			s.dropSpentBolts()
		}
		return
	}
}

func (s *Scene) dropSpentBolts() {
	live := s.bolts[:0]
	for _, b := range s.bolts {
		if b.Active {
			live = append(live, b)
		}
	}
	s.bolts = live
}
