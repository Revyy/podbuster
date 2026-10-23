package game

import (
	"math"
)

// Weapon owns a slot's firing state, cooldown, and projectile creation from a
// firing pose. The player owns two slots, but the world only sees the resulting
// projectiles after they are fired.
type Weapon interface {
	Update()
	Reset()
	CanFire() bool
	Fire(x, y, heading float64) []*Projectile
}

// WeaponStatus supplies the information shown in the weapon status HUD.
type WeaponStatus interface {
	Name() string
	CooldownRemaining() int
	CooldownTotal() int
}

const (
	rocketCooldown = 180
	rocketSpeed    = 9.0
	rocketSpread   = 8.0 * math.Pi / 180.0
	rocketNose     = 14.0
	rocketSide     = 5.0
	rocketSize     = 10.0
)

// cooldownWeapon shares the tick-down cooldown state and behavior every
// current Weapon implementation needs, so BasicWeapon and RocketWeapon don't
// each redefine identical Update/Reset/CanFire bodies.
type cooldownWeapon struct {
	cooldown      int
	cooldownTotal int
}

func (w *cooldownWeapon) Update() {
	if w.cooldown > 0 {
		w.cooldown--
	}
}

func (w *cooldownWeapon) Reset() {
	w.cooldown = 0
}

func (w *cooldownWeapon) CanFire() bool {
	return w.cooldown == 0
}

func (w *cooldownWeapon) CooldownRemaining() int {
	return w.cooldown
}

func (w *cooldownWeapon) CooldownTotal() int {
	return w.cooldownTotal
}

// BasicWeapon is the default player weapon: a single shot with a cooldown and a
// fixed impact cost.
type BasicWeapon struct {
	cooldownWeapon
	hitCost int
}

// RocketWeapon fires a three-round burst at a slow cadence with a heavier hit cost.
type RocketWeapon struct {
	cooldownWeapon
	hitCost int
}

// NewBasicWeapon creates a standard weapon that costs one enemy hit per shot.
func NewBasicWeapon() *BasicWeapon {
	return &BasicWeapon{
		cooldownWeapon: cooldownWeapon{cooldownTotal: shootCooldown},
		hitCost:        1,
	}
}

// NewRocketWeapon creates a burst-damage secondary weapon that fires three rockets.
func NewRocketWeapon() *RocketWeapon {
	return &RocketWeapon{
		cooldownWeapon: cooldownWeapon{cooldownTotal: rocketCooldown},
		hitCost:        2,
	}
}

func (w *BasicWeapon) Name() string {
	return "Blaster"
}

func (w *RocketWeapon) Name() string {
	return "Rockets"
}

func (w *BasicWeapon) Fire(x, y, heading float64) []*Projectile {
	if w.cooldown > 0 {
		return nil
	}
	w.cooldown = w.cooldownTotal
	originX := x + math.Cos(heading)*playerSize/2
	originY := y + math.Sin(heading)*playerSize/2
	vx := math.Cos(heading) * projectileSpeed
	vy := math.Sin(heading) * projectileSpeed
	return []*Projectile{NewProjectile(originX, originY, vx, vy, w.hitCost)}
}

func (w *RocketWeapon) Fire(x, y, heading float64) []*Projectile {
	if w.cooldown > 0 {
		return nil
	}
	w.cooldown = w.cooldownTotal

	baseX := x + math.Cos(heading)*rocketNose
	baseY := y + math.Sin(heading)*rocketNose
	perpX := math.Cos(heading + math.Pi/2)
	perpY := math.Sin(heading + math.Pi/2)
	spreads := []float64{0, -rocketSpread, rocketSpread}
	shots := make([]*Projectile, 0, len(spreads))
	for _, spread := range spreads {
		angle := heading + spread
		lateral := 0.0
		if spread != 0 {
			lateral = rocketSide * math.Copysign(1, spread)
		}
		originX := baseX + perpX*lateral
		originY := baseY + perpY*lateral
		vx := math.Cos(angle) * rocketSpeed
		vy := math.Sin(angle) * rocketSpeed
		shot := NewProjectile(originX, originY, vx, vy, w.hitCost)
		shot.Size = rocketSize
		shot.Kind = projectileKindRocket
		shots = append(shots, shot)
	}
	return shots
}
