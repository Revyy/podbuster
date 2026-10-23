package game

import "math"

const (
	// EnemySize is the width and height of the enemy bounding box in pixels.
	EnemySize = 24.0
	// EnemySpeed is pixels moved per tick toward the player.
	EnemySpeed = 0.8
	// EnemyHits is the number of projectile impacts needed to kill a Standard
	// enemy.
	EnemyHits = 3
	// EnemySeparation is the distance enemies keep from one another. At 1.4x
	// the enemy box it leaves a visible gap, so a crowd stays a set of
	// individually shootable targets rather than one pile.
	EnemySeparation = EnemySize * 1.4
	// enemyAvoidRangeFactor times a pair's separation is how far an enemy looks
	// for that neighbour to steer around. At twice the separation, avoidance starts early and gently, leaving the
	// hard separation pass a rarely-triggered floor rather than the main
	// mechanism.
	enemyAvoidRangeFactor = 2.0
	// enemyAvoidWeight is how strongly avoidance bends the homing direction.
	// Below 1 it can never fully overpower the urge to reach the player.
	enemyAvoidWeight = 0.6
	// enemyKnockbackSpeed is the impulse an enemy receives from contact.
	enemyKnockbackSpeed = 6.0
	// enemyFriction decays knockback velocity each tick, handing control back
	// to the homing behaviour over roughly a second.
	enemyFriction = 0.92
	// enemyFlashTicks is how long an enemy flashes after contact.
	enemyFlashTicks = 30
	// enemyDamageFlashTicks is how long the enemy shows a projectile-impact reaction.
	enemyDamageFlashTicks = 12
	// enemyAnimationTicks is the fixed practice-mode and explosion duration.
	enemyAnimationTicks = 24
)

// tierStats are the levers every archetype shares, scaled by Tier.
type tierStats struct {
	hits          int
	dartHits      int
	size          float64
	speed         float64
	contactDamage int
}

var tierTable = [...]tierStats{
	TierLight:    {hits: 2, dartHits: 1, size: 0.85, speed: 1.1, contactDamage: 1},
	TierStandard: {hits: EnemyHits, dartHits: 2, size: 1, speed: 1, contactDamage: 1},
	TierHeavy:    {hits: 5, dartHits: 3, size: 1.3, speed: 0.85, contactDamage: 2},
}

func (t Tier) stats() tierStats {
	return tierEntry(tierTable[:], t)
}

// tierEntry indexes a per-tier table, falling back to the Standard row for an
// out-of-range tier.
func tierEntry[T any](table []T, t Tier) T {
	if int(t) < 0 || int(t) >= len(table) {
		return table[TierStandard]
	}
	return table[t]
}

// StartingHits is the number of projectile impacts a fresh enemy of this
// archetype and tier takes to kill.
func StartingHits(a Archetype, t Tier) int {
	if a == ArchetypeDart {
		return t.stats().dartHits
	}
	return t.stats().hits
}

// EnemyState describes the lifecycle of an enemy backed by a Pod.
type EnemyState int

const (
	EnemyActive EnemyState = iota
	EnemyTerminating
	EnemyExploding
)

// Enemy represents a Pod enemy in the game world. One enemy stands for exactly
// one live Pod.
type Enemy struct {
	X, Y       float64
	VX, VY     float64
	Hits       int
	Archetype  Archetype
	Tier       Tier
	PodName    string
	Namespace  string
	State      EnemyState
	StateTimer int
	// FlashTimer counts down the visual reaction to contact.
	FlashTimer int
	// DamageFlashTimer counts down the visual reaction to a projectile impact.
	DamageFlashTimer int
	// Telegraph is the 0..1 wind-up of an attack about to happen. Behaviour
	// tickets drive it; the Turret eye, Sentinel ring and Dart eye glow with it.
	Telegraph float64
	// Heading is a smoothed per-tick velocity used only to pose the droid:
	// tentacles trail it and the Dart faces along it.
	HeadingX, HeadingY float64
	seed               float64
	// turret is a Turret's firing cycle; see turret.go.
	turret turretCycle
	// lunge is the straight-line rush shared by Dart charges and Sentinel dashes.
	lunge lunge
	// dart is the Dart's wind-up/charge/brake cycle; other archetypes ignore it.
	dart dartCycle
	// sentinel is the Sentinel's orbit direction and dash countdown.
	sentinel sentinelCycle
	// Symptoms are the Pod's live cosmetic signals; they never affect behaviour.
	Symptoms    Symptoms
	rebootTimer int
}

// headingBlend is how much of each tick's movement enters the smoothed heading.
const headingBlend = 0.15

// NewEnemy creates an active Standard Probe for the given pod at position (x, y).
func NewEnemy(podName, namespace string, x, y float64) *Enemy {
	return NewEnemyOf(podName, namespace, ArchetypeProbe, TierStandard, x, y)
}

// NewEnemyOf creates an active enemy of the given archetype and tier.
func NewEnemyOf(podName, namespace string, archetype Archetype, tier Tier, x, y float64) *Enemy {
	var seed float64
	for _, r := range podName {
		seed += float64(r)
	}
	return &Enemy{
		seed:      seed,
		PodName:   podName,
		Namespace: namespace,
		X:         x,
		Y:         y,
		Hits:      StartingHits(archetype, tier),
		Archetype: archetype,
		Tier:      tier,
		State:     EnemyActive,
	}
}

// Scale is the enemy's size relative to a Standard enemy.
func (e *Enemy) Scale() float64 { return e.Tier.stats().size }

// Size is the width and height of the enemy's bounding box.
func (e *Enemy) Size() float64 { return EnemySize * e.Scale() }

// CruiseSpeed is the pixels moved per tick while homing.
func (e *Enemy) CruiseSpeed() float64 { return EnemySpeed * e.Tier.stats().speed }

// ContactDamage is the Health a touch from this enemy costs the player.
func (e *Enemy) ContactDamage() int { return e.Tier.stats().contactDamage }

// Update moves the enemy one step toward (targetX, targetY), bent away from
// (steerX, steerY), on top of any knockback velocity it still carries. The
// steer vector is the Separation avoidance direction; passing (0, 0) homes
// straight at the target.
func (e *Enemy) Update(targetX, targetY, steerX, steerY float64) {
	if e.Archetype == ArchetypeDart {
		e.updateDart(targetX, targetY, steerX, steerY)
		return
	}
	if e.Archetype == ArchetypeSentinel {
		e.updateSentinel(targetX, targetY, steerX, steerY)
		return
	}
	e.step(targetX, targetY, steerX, steerY, e.CruiseSpeed())
}

// step is one tick of movement at the given homing speed. An active lunge
// replaces homing and ignores avoidance steering and knockback.
func (e *Enemy) step(targetX, targetY, steerX, steerY, speed float64) {
	if e.FlashTimer > 0 {
		e.FlashTimer--
	}
	if e.DamageFlashTimer > 0 {
		e.DamageFlashTimer--
	}

	prevX, prevY := e.X, e.Y
	defer func() {
		e.HeadingX += ((e.X - prevX) - e.HeadingX) * headingBlend
		e.HeadingY += ((e.Y - prevY) - e.HeadingY) * headingBlend
	}()

	if e.lunge.active {
		e.advanceLunge()
		return
	}

	e.X += e.VX
	e.Y += e.VY
	e.VX *= enemyFriction
	e.VY *= enemyFriction
	if math.Hypot(e.VX, e.VY) < 0.05 {
		e.VX, e.VY = 0, 0
	}

	dx := targetX - e.X
	dy := targetY - e.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	// Blend homing with avoidance, then renormalise: the enemy always travels
	// exactly its cruise speed, so crowding changes its heading and never its pace.
	hx := dx/dist + steerX*enemyAvoidWeight
	hy := dy/dist + steerY*enemyAvoidWeight
	mag := math.Hypot(hx, hy)
	if mag < 0.0001 {
		// Avoidance exactly cancelled homing; keep homing rather than stall.
		hx, hy, mag = dx/dist, dy/dist, 1
	}
	e.X += (hx / mag) * speed
	e.Y += (hy / mag) * speed
}

// Bounds returns the enemy's collision rectangle.
func (e *Enemy) Bounds() Rect {
	size := e.Size()
	return Rect{X: e.X - size/2, Y: e.Y - size/2, W: size, H: size}
}

// Hit reduces hits by cost and begins termination if none remain. The hit
// count is clamped at zero so a projectile cannot drive a surviving enemy
// below zero or trigger a second kill event.
func (e *Enemy) Hit(cost int) *KillEvent {
	if e.State != EnemyActive {
		return nil
	}
	if cost < 0 {
		cost = 0
	}
	e.Hits -= cost
	if e.Hits < 0 {
		e.Hits = 0
	}
	e.DamageFlashTimer = enemyDamageFlashTicks
	if e.Hits == 0 {
		e.State = EnemyTerminating
		e.StateTimer = 0
		return &KillEvent{PodName: e.PodName, Namespace: e.Namespace}
	}

	return nil
}

// BeginTerminating moves an enemy into its frozen termination state. A zero
// timer means it waits for confirmation from the provider.
func (e *Enemy) BeginTerminating(ticks int) {
	e.State = EnemyTerminating
	e.StateTimer = ticks
}

// BeginExploding starts the final removal animation.
func (e *Enemy) BeginExploding(ticks int) {
	e.State = EnemyExploding
	e.StateTimer = ticks
}

// RevertToActive restores an enemy after a failed deletion request.
func (e *Enemy) RevertToActive() {
	e.State = EnemyActive
	e.StateTimer = 0
	e.resetBehaviour()
}

// resetBehaviour restarts attack cycles so a revived enemy never lunges the
// instant it returns.
func (e *Enemy) resetBehaviour() {
	e.lunge = lunge{}
	e.dart = dartCycle{}
	e.sentinel = sentinelCycle{dir: e.sentinel.dir}
	e.turret = turretCycle{}
	e.Telegraph = 0
}

// AdvanceState advances non-active lifecycle timers and reports when an
// explosion has completed.
func (e *Enemy) AdvanceState() bool {
	if e.State == EnemyActive || e.StateTimer <= 0 {
		return e.State == EnemyExploding && e.StateTimer <= 0
	}
	e.StateTimer--
	if e.State == EnemyTerminating && e.StateTimer == 0 {
		e.BeginExploding(enemyAnimationTicks)
		return false
	}
	return e.State == EnemyExploding && e.StateTimer == 0
}

// Knockback pushes the enemy along the unit vector (nx, ny) and starts its
// contact flash. Contact never costs an enemy hits: only projectiles can kill,
// so only deliberate shooting can delete a Pod.
func (e *Enemy) Knockback(nx, ny float64) {
	if e.Archetype == ArchetypeTurret {
		// Immovable: contact pushes only the player.
		e.FlashTimer = enemyFlashTicks
		return
	}
	e.VX += nx * enemyKnockbackSpeed
	e.VY += ny * enemyKnockbackSpeed
	e.FlashTimer = enemyFlashTicks
}

// HoldStation ticks a stationary enemy's reaction timers without moving it.
func (e *Enemy) HoldStation() {
	if e.FlashTimer > 0 {
		e.FlashTimer--
	}
	if e.DamageFlashTimer > 0 {
		e.DamageFlashTimer--
	}
	e.VX, e.VY = 0, 0
	e.HeadingX, e.HeadingY = 0, 0
}
