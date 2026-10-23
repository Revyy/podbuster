package game

import (
	"math"
	"math/rand/v2"
)

const (
	spawnMargin  = 50.0 // pixels outside screen edge where enemies spawn
	screenWidth  = 800.0
	screenHeight = 600.0
	// respawnDelayTicks is how long a killed enemy stays away in practice
	// mode before returning. Its Pod is still running, so the enemy comes back.
	respawnDelayTicks = 180
	// separationIterations caps the relaxation sweeps spent prising enemies
	// apart each tick. Soft avoidance keeps a crowd mostly clear, so the first
	// sweep usually finds nothing and the rest are skipped.
	separationIterations = 64
	// separationEpsilon is the slack that makes the sweep terminate. Sweeps
	// approach the target distance asymptotically, so corrections smaller than
	// this count as settled.
	separationEpsilon = 0.01
)

// pendingRespawn is an enemy killed in practice mode, waiting to return.
type pendingRespawn struct {
	podName   string
	namespace string
	archetype Archetype
	tier      Tier
	symptoms  Symptoms
	timer     int
}

// Scene manages the enemy pool and projectiles. It owns the game event channel
// and routes kill events back to the bridge.
type Scene struct {
	enemies           []*Enemy
	projectiles       []*Projectile
	bolts             []*Bolt
	bursts            []ImpactBurst
	projectileImpacts []ProjectileImpact
	rng               RNG
	respawns          []pendingRespawn
	kills             int
	chaosMode         bool
	sessionID         uint64
	lastError         string
	gameEvents        <-chan GameEvent
	killEvents        chan<- KillEvent
}

// NewScene creates a Scene wired to the given channels.
func NewScene(gameEvents <-chan GameEvent, killEvents chan<- KillEvent) *Scene {
	return &Scene{
		gameEvents: gameEvents,
		killEvents: killEvents,
		rng:        rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

// AddProjectiles appends player-fired projectiles to the scene.
func (s *Scene) AddProjectiles(projectiles []*Projectile) {
	s.projectiles = append(s.projectiles, projectiles...)
}

// Enemies returns the current enemy slice (may contain terminating or
// exploding enemies during the tick).
func (s *Scene) Enemies() []*Enemy {
	return s.enemies
}

// Projectiles returns the current projectile slice.
func (s *Scene) Projectiles() []*Projectile {
	return s.projectiles
}

// Kills returns the number of enemies killed this session. In chaos mode this
// is also the number of Pods deleted.
func (s *Scene) Kills() int {
	return s.kills
}

// LastError returns the most recent error reported by the bridge, or "".
func (s *Scene) LastError() string {
	return s.lastError
}

// ClearError forgets any error previously reported by the bridge.
func (s *Scene) ClearError() {
	s.lastError = ""
}

// BeginSession prepares the scene for a new session in the given mode. Kills
// are stamped with sessionID so the bridge can discard any that outlive it.
func (s *Scene) BeginSession(sessionID uint64, chaosMode bool) {
	s.sessionID = sessionID
	s.chaosMode = chaosMode
	s.Reset()
	s.ClearError()
}

// Reset empties the enemy pool, projectiles, pending respawns and kill count.
// Called when a session ends; the pool is meaningless without a live watch.
func (s *Scene) Reset() {
	s.enemies = nil
	s.projectiles = nil
	s.ClearBolts()
	s.respawns = nil
	s.kills = 0
}

// DrainStaleEvents discards events buffered from a session that has ended, so
// they cannot repopulate the pool after it was cleared.
func (s *Scene) DrainStaleEvents() {
	for {
		select {
		case ev := <-s.gameEvents:
			if ev.Type == EventError {
				s.lastError = ev.Message
			}
		default:
			return
		}
	}
}

// Update drains bridge events, moves enemies toward the player,
// moves projectiles, and resolves collisions and contact.
func (s *Scene) Update(player *Player) {
	s.drainEvents()
	s.updateRespawns()
	s.tickSymptoms()
	s.updateEnemies(player.X, player.Y)
	s.updateTurrets(player.X, player.Y)
	s.separate()
	s.updateProjectiles()
	s.updateBolts()
	s.resolveCollisions()
	s.resolveContact(player)
	s.resolveBolts(player)
	s.prune()
}

// UpdateWithoutPlayer advances the world while the player's ship is absent,
// during the death animation. No contact is resolved and nothing is fired.
func (s *Scene) UpdateWithoutPlayer(playerX, playerY float64) {
	s.drainEvents()
	s.updateRespawns()
	s.tickSymptoms()
	s.updateEnemies(playerX, playerY)
	s.separate()
	s.updateProjectiles()
	s.updateBolts()
	s.resolveCollisions()
	s.prune()
}

func (s *Scene) drainEvents() {
	for {
		select {
		case ev := <-s.gameEvents:
			switch ev.Type {
			case EventSpawnEnemy:
				s.spawnEnemy(ev.PodName, ev.Namespace, ev.Archetype, ev.Tier).Symptoms = ev.Symptoms
			case EventSymptomsChanged:
				s.updateSymptoms(ev.PodName, ev.Namespace, ev.Symptoms)
			case EventDespawnEnemy:
				s.despawnEnemy(ev.PodName, ev.Namespace)
			case EventPodTerminating:
				s.terminateEnemy(ev.PodName, ev.Namespace)
			case EventDeletionFailed:
				s.revertEnemy(ev.PodName, ev.Namespace)
			case EventError:
				s.lastError = ev.Message
			}
		default:
			return
		}
	}
}

func (s *Scene) spawnEnemy(podName, namespace string, archetype Archetype, tier Tier) *Enemy {
	x, y := randomSpawnPosition()
	if archetype == ArchetypeTurret {
		x, y = s.turretSpawnPosition()
	}
	e := NewEnemyOf(podName, namespace, archetype, tier, x, y)
	if archetype == ArchetypeSentinel {
		e.initSentinel(s.rng)
	}
	s.enemies = append(s.enemies, e)
	return e
}

func (s *Scene) despawnEnemy(podName, namespace string) {
	for _, e := range s.enemies {
		if e.PodName == podName && e.Namespace == namespace {
			if e.State != EnemyExploding {
				e.BeginExploding(enemyAnimationTicks)
			}
		}
	}
	// A Pod that is really gone must not come back as an enemy.
	s.cancelRespawn(podName, namespace)
}

func (s *Scene) terminateEnemy(podName, namespace string) {
	for _, e := range s.enemies {
		if e.PodName == podName && e.Namespace == namespace && e.State == EnemyActive {
			e.BeginTerminating(0)
		}
	}
}

func (s *Scene) revertEnemy(podName, namespace string) {
	for _, e := range s.enemies {
		if e.PodName == podName && e.Namespace == namespace && e.State == EnemyTerminating {
			e.RevertToActive()
		}
	}
}

func (s *Scene) cancelRespawn(podName, namespace string) {
	remaining := s.respawns[:0]
	for _, r := range s.respawns {
		if r.podName == podName && r.namespace == namespace {
			continue
		}
		remaining = append(remaining, r)
	}
	s.respawns = remaining
}

// updateRespawns returns practice-mode enemies to the pool once their delay
// has elapsed.
func (s *Scene) updateRespawns() {
	if len(s.respawns) == 0 {
		return
	}
	pending := s.respawns[:0]
	for _, r := range s.respawns {
		r.timer--
		if r.timer <= 0 {
			s.spawnEnemy(r.podName, r.namespace, r.archetype, r.tier).Symptoms = r.symptoms
			continue
		}
		pending = append(pending, r)
	}
	s.respawns = pending
}

func (s *Scene) updateEnemies(playerX, playerY float64) {
	for i, e := range s.enemies {
		if e.State == EnemyActive && e.Archetype == ArchetypeTurret {
			e.HoldStation()
			continue
		}
		if e.State == EnemyActive {
			sx, sy := s.avoidanceVector(i)
			e.Update(playerX, playerY, sx, sy)
			continue
		}
		e.AdvanceState()
	}
}

// avoidanceVector returns the unit direction the enemy at index i should lean
// to stay clear of its neighbours, or (0, 0) when it has room. Contributions
// are weighted by closeness, so a neighbour about to touch dominates one at
// the edge of range.
func (s *Scene) avoidanceVector(i int) (float64, float64) {
	e := s.enemies[i]
	var ax, ay float64
	for j, other := range s.enemies {
		if i == j || other.State != EnemyActive {
			continue
		}
		dx := e.X - other.X
		dy := e.Y - other.Y
		distSq := dx*dx + dy*dy
		avoidRange := pairSeparation(e, other) * enemyAvoidRangeFactor
		if distSq >= avoidRange*avoidRange {
			continue
		}
		nx, ny, dist := separationNormal(dx, dy, distSq)
		ax += nx * (1 - dist/avoidRange)
		ay += ny * (1 - dist/avoidRange)
	}
	mag := math.Hypot(ax, ay)
	if mag < 0.0001 {
		return 0, 0
	}
	return ax / mag, ay / mag
}

func (s *Scene) updateProjectiles() {
	for _, projectile := range s.projectiles {
		if projectile.Active {
			projectile.Update()
			// Deactivate projectiles that leave the screen.
			if projectile.X < 0 || projectile.X > screenWidth || projectile.Y < 0 || projectile.Y > screenHeight {
				projectile.Active = false
			}
		}
	}
	s.updateProjectileImpacts()
}

func (s *Scene) resolveCollisions() {
	for _, projectile := range s.projectiles {
		if !projectile.Active {
			continue
		}
		for _, e := range s.enemies {
			if e.State != EnemyActive {
				continue
			}
			if !projectile.Bounds().Overlaps(e.Bounds()) {
				continue
			}
			projectile.Active = false
			s.recordProjectileImpact(projectile)
			if kill := e.Hit(projectile.HitCost); kill != nil {
				if !s.chaosMode {
					e.BeginTerminating(enemyAnimationTicks)
				}
				s.recordKill(*kill, e.Archetype, e.Tier, e.Symptoms)
			}
			break
		}
	}
}

func (s *Scene) recordKill(kill KillEvent, archetype Archetype, tier Tier, symptoms Symptoms) {
	s.kills++
	kill.SessionID = s.sessionID
	select {
	case s.killEvents <- kill:
	default:
	}

	// In practice mode the Pod survives its enemy's kill, so the enemy returns.
	// In chaos mode the Pod is really deleted and the watch's despawn is final.
	if !s.chaosMode {
		s.respawns = append(s.respawns, pendingRespawn{
			podName:   kill.PodName,
			namespace: kill.Namespace,
			archetype: archetype,
			tier:      tier,
			symptoms:  symptoms,
			timer:     respawnDelayTicks,
		})
	}
}

// resolveContact costs the player health when an enemy overlaps it and pushes
// both apart. Contact never costs the enemy hits.
func (s *Scene) resolveContact(player *Player) {
	if player.IsRolling() {
		return
	}
	playerBounds := player.Bounds()
	for _, e := range s.enemies {
		if e.State != EnemyActive {
			continue
		}
		if !playerBounds.Overlaps(e.Bounds()) {
			continue
		}
		// Any contact ends a lunge, even while the player is invulnerable and
		// takes no damage: a lunge must not keep pressing through the ship.
		e.endLunge()
		nx, ny := contactNormal(player.X, player.Y, e.X, e.Y)
		if player.IsInvulnerable() {
			return
		}
		if player.TakeContactDamage(e.ContactDamage(), -nx, -ny) {
			e.Knockback(nx, ny)
		}
		return
	}
}

// separate enforces Separation: after this pass no two active enemies are
// closer than their pairSeparation. Overlapping enemies would otherwise behave as a
// single target, because a projectile stops at the first enemy it touches and
// contact resolves against only one enemy per tick.
//
// It corrects positions only. Knockback stays the sole author of enemy
// velocity, so a crowd cannot fling its members at speeds nothing else in the
// game produces.
//
// One sweep only resolves an isolated pair exactly: prising A off B can push A
// back into C. The sweep is repeated until nothing moves, which in a settled
// crowd is a single no-op pass.
func (s *Scene) separate() {
	for range separationIterations {
		if !s.separateOnce() {
			return
		}
	}
}

// separateOnce runs one relaxation sweep and reports whether it moved anything.
func (s *Scene) separateOnce() bool {
	moved := false
	for i, a := range s.enemies {
		if a.State != EnemyActive {
			continue
		}
		for _, b := range s.enemies[i+1:] {
			if b.State != EnemyActive {
				continue
			}
			// A Turret never gives way, so its neighbour takes the whole correction.
			// Two Turrets share the correction: they never overlap either.
			wa, wb := 0.5, 0.5
			switch {
			case a.Archetype == ArchetypeTurret && b.Archetype == ArchetypeTurret:
			case a.Archetype == ArchetypeTurret:
				wa, wb = 0, 1
			case b.Archetype == ArchetypeTurret:
				wa, wb = 1, 0
			}
			dx := b.X - a.X
			dy := b.Y - a.Y
			distSq := dx*dx + dy*dy
			separation := pairSeparation(a, b)
			if distSq >= separation*separation {
				continue
			}
			nx, ny, dist := separationNormal(dx, dy, distSq)
			// Overshoot slightly. Sweeps converge on the target distance from
			// below, so aiming exactly at it would leave every pair a hair
			// too close forever.
			push := separation + separationEpsilon - dist
			// Split the correction evenly unless one side is immovable.
			a.X -= nx * push * wa
			a.Y -= ny * push * wa
			b.X += nx * push * wb
			b.Y += ny * push * wb
			if push/2 > separationEpsilon {
				moved = true
			}
			// The no-overlap floor blocking a lunge ends it early.
			a.endLunge()
			b.endLunge()
		}
	}
	return moved
}

// pairSeparation is the distance two enemies keep from each other: the
// Standard separation scaled by the mean of their sizes.
func pairSeparation(a, b *Enemy) float64 {
	return EnemySeparation * (a.Scale() + b.Scale()) / 2
}

// separationNormal returns the unit vector from one enemy toward the other
// along with their distance. Coincident enemies fall back to a fixed direction
// at zero distance, so a pile that lands on exactly one point still comes
// apart instead of stalling on a zero-length vector.
func separationNormal(dx, dy, distSq float64) (float64, float64, float64) {
	if distSq < 0.0001 {
		return 1, 0, 0
	}
	dist := math.Sqrt(distSq)
	return dx / dist, dy / dist, dist
}

// contactNormal returns the unit vector pointing from the player toward the
// enemy. Coincident positions fall back to a fixed direction so the impulse is
// never zero.
func contactNormal(playerX, playerY, enemyX, enemyY float64) (float64, float64) {
	dx := enemyX - playerX
	dy := enemyY - playerY
	dist := math.Hypot(dx, dy)
	if dist < 0.0001 {
		return 1, 0
	}
	return dx / dist, dy / dist
}

func (s *Scene) prune() {
	remaining := s.enemies[:0]
	for _, e := range s.enemies {
		if e.State != EnemyExploding || e.StateTimer > 0 {
			remaining = append(remaining, e)
		}
	}
	s.enemies = remaining

	activeProjectiles := s.projectiles[:0]
	for _, projectile := range s.projectiles {
		if projectile.Active {
			activeProjectiles = append(activeProjectiles, projectile)
		}
	}
	s.projectiles = activeProjectiles
}

// randomSpawnPosition returns a position just outside the screen edges.
func randomSpawnPosition() (float64, float64) {
	edge := rand.IntN(4)
	switch edge {
	case 0: // top
		return rand.Float64() * screenWidth, -spawnMargin
	case 1: // bottom
		return rand.Float64() * screenWidth, screenHeight + spawnMargin
	case 2: // left
		return -spawnMargin, rand.Float64() * screenHeight
	default: // right
		return screenWidth + spawnMargin, rand.Float64() * screenHeight
	}
}
