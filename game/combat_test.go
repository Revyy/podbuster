package game

import (
	"math"
	"testing"
)

func TestEnemy_HitCausesDamageFlashAndLowersHits(t *testing.T) {
	enemy := NewEnemy("pod-a", "default", 400, 300)

	enemy.Hit(1)

	if enemy.Hits != EnemyHits-1 {
		t.Fatalf("expected %d hits after a projectile hit, got %d", EnemyHits-1, enemy.Hits)
	}
	if enemy.DamageFlashTimer == 0 {
		t.Fatal("expected a projectile hit to trigger a damage flash")
	}
	if enemy.FlashTimer != 0 {
		t.Fatal("expected contact flash to stay separate from projectile damage feedback")
	}
}

func TestEnemy_ContactFlashDoesNotChangeRemainingHullDisplay(t *testing.T) {
	enemy := NewEnemy("pod-a", "default", 400, 300)
	enemy.Hit(1)
	enemy.Knockback(1, 0)

	if enemy.Hits != EnemyHits-1 {
		t.Fatalf("expected contact to leave the hit counter at %d, got %d", EnemyHits-1, enemy.Hits)
	}
	if enemy.DamageFlashTimer == 0 {
		t.Fatal("expected the enemy to remain visibly damaged after the hit")
	}
}

func TestEnemy_HitClampsAtZero(t *testing.T) {
	enemy := NewEnemy("pod-a", "default", 400, 300)

	if kill := enemy.Hit(10); kill == nil {
		t.Fatal("expected an overkill hit to kill the enemy")
	}
	if enemy.Hits != 0 {
		t.Fatalf("expected remaining hits to clamp at zero, got %d", enemy.Hits)
	}
	if enemy.State != EnemyTerminating {
		t.Fatal("expected the enemy to enter termination after overkill")
	}
	if kill := enemy.Hit(1); kill != nil {
		t.Fatal("expected a terminating enemy not to emit a second kill")
	}
}

func TestPracticeMode_KillTransitionsThroughTerminationAndExplosion(t *testing.T) {
	events := make(chan GameEvent, 8)
	scene := NewScene(events, make(chan KillEvent, 8))
	scene.BeginSession(1, false)
	enemy := NewEnemy("pod-a", "default", 400, 300)
	scene.enemies = []*Enemy{enemy}
	scene.projectiles = []*Projectile{{X: 400, Y: 300, Active: true, HitCost: EnemyHits}}

	scene.resolveCollisions()
	if enemy.State != EnemyTerminating {
		t.Fatalf("expected terminating state, got %v", enemy.State)
	}

	for range enemyAnimationTicks {
		scene.updateEnemies(400, 300)
	}
	if enemy.State != EnemyExploding {
		t.Fatalf("expected exploding state after termination timer, got %v", enemy.State)
	}

	for range enemyAnimationTicks {
		scene.updateEnemies(400, 300)
	}
	scene.prune()
	if len(scene.enemies) != 0 {
		t.Fatalf("expected expired explosion to be removed, got %d enemies", len(scene.enemies))
	}
}

func TestScene_TerminationEventsControlEnemyLifecycle(t *testing.T) {
	events := make(chan GameEvent, 8)
	scene := NewScene(events, make(chan KillEvent, 8))
	scene.BeginSession(1, true)
	enemy := NewEnemy("pod-a", "default", 400, 300)
	scene.enemies = []*Enemy{enemy}

	events <- GameEvent{Type: EventPodTerminating, PodName: "pod-a", Namespace: "default"}
	scene.drainEvents()
	if enemy.State != EnemyTerminating {
		t.Fatalf("expected external termination to freeze enemy, got %v", enemy.State)
	}
	if scene.Kills() != 0 {
		t.Fatalf("expected external termination not to count as a kill, got %d", scene.Kills())
	}

	events <- GameEvent{Type: EventDeletionFailed, PodName: "pod-a", Namespace: "default"}
	scene.drainEvents()
	if enemy.State != EnemyActive {
		t.Fatalf("expected failed deletion to restore active state, got %v", enemy.State)
	}

	events <- GameEvent{Type: EventDespawnEnemy, PodName: "pod-a", Namespace: "default"}
	scene.drainEvents()
	if enemy.State != EnemyExploding || enemy.StateTimer != enemyAnimationTicks {
		t.Fatalf("expected confirmed deletion to start explosion, got state=%v timer=%d", enemy.State, enemy.StateTimer)
	}
}

func TestScene_ResetDiscardsLifecycleStates(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.enemies = []*Enemy{
		NewEnemy("active", "default", 0, 0),
		{PodName: "terminating", State: EnemyTerminating},
		{PodName: "exploding", State: EnemyExploding, StateTimer: enemyAnimationTicks},
	}

	scene.Reset()

	if len(scene.enemies) != 0 {
		t.Fatalf("expected reset to discard all lifecycle states, got %d enemies", len(scene.enemies))
	}
}

func TestScene_ResetClearsProjectileImpacts(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.projectileImpacts = []ProjectileImpact{{X: 200, Y: 200}}

	scene.Reset()

	if got := len(scene.ProjectileImpacts()); got != 0 {
		t.Fatalf("expected reset to clear projectile impacts, got %d", got)
	}
}

func TestScene_ResolveCollisionsUsesProjectileHitCostInFiringOrder(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	first := NewEnemy("pod-a", "default", 200, 200)
	second := NewEnemy("pod-b", "default", 210, 200)
	scene.enemies = []*Enemy{first, second}
	scene.projectiles = []*Projectile{
		{X: 200, Y: 200, Active: true, HitCost: 3},
		{X: 210, Y: 200, Active: true, HitCost: 1},
	}

	scene.resolveCollisions()

	if first.State != EnemyTerminating || second.State != EnemyActive {
		t.Fatalf("expected the first enemy to terminate and the second to remain active, got state=%v/%v", first.State, second.State)
	}
	if first.Hits != 0 || second.Hits != 2 {
		t.Fatalf("expected first enemy to be killed and second to lose one hit, got %d/%d", first.Hits, second.Hits)
	}
	if scene.projectiles[0].Active || scene.projectiles[1].Active {
		t.Fatal("expected both projectiles to deactivate after resolving their first overlap")
	}
}

func TestScene_ResolveCollisionsRecordsProjectileImpact(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	enemy := NewEnemy("pod-a", "default", 200, 200)
	scene.enemies = []*Enemy{enemy}
	scene.projectiles = []*Projectile{{
		X:       200,
		Y:       200,
		VX:      3,
		VY:      4,
		Active:  true,
		HitCost: 2,
		Kind:    projectileKindRocket,
		Size:    projectileSize,
	}}

	scene.resolveCollisions()

	if got := len(scene.ProjectileImpacts()); got != 1 {
		t.Fatalf("expected 1 projectile impact, got %d", got)
	}
	impact := scene.ProjectileImpacts()[0]
	if impact.X != 200 || impact.Y != 200 {
		t.Fatalf("expected impact at 200,200 got %v,%v", impact.X, impact.Y)
	}
	if impact.Kind != projectileKindRocket {
		t.Fatalf("expected rocket impact kind, got %v", impact.Kind)
	}
	if impact.Direction != math.Atan2(4, 3) {
		t.Fatalf("expected impact direction %v, got %v", math.Atan2(4, 3), impact.Direction)
	}
}

func TestScene_ProjectileLeavingScreenCreatesNoImpact(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.projectiles = []*Projectile{{
		X:      screenWidth - 1,
		Y:      200,
		VX:     2,
		VY:     0,
		Active: true,
		Size:   projectileSize,
	}}

	scene.updateProjectiles()
	scene.prune()

	if got := len(scene.ProjectileImpacts()); got != 0 {
		t.Fatalf("expected no projectile impact when leaving the screen, got %d", got)
	}
}

func TestScene_ProjectileImpactsAgeOut(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.projectileImpacts = []ProjectileImpact{{X: 200, Y: 200}}

	for range projectileImpactTicks {
		scene.updateProjectiles()
	}

	if got := len(scene.ProjectileImpacts()); got != 0 {
		t.Fatalf("expected projectile impacts to age out, got %d", got)
	}
}

func TestContact_CostsHealthAndKnocksBothApart(t *testing.T) {
	killEvents := make(chan KillEvent, 8)
	scene := NewScene(make(chan GameEvent), killEvents)
	player := NewPlayer(400, 300)

	enemy := NewEnemy("pod-a", "default", 405, 300)
	scene.enemies = append(scene.enemies, enemy)

	scene.resolveContact(player)

	if player.Health != PlayerMaxHealth-1 {
		t.Fatalf("expected health %d, got %d", PlayerMaxHealth-1, player.Health)
	}
	// The enemy is to the player's right, so it is pushed right and the
	// player left.
	if enemy.VX <= 0 {
		t.Fatalf("expected enemy knocked away in +X, got VX %v", enemy.VX)
	}
	if player.VX >= 0 {
		t.Fatalf("expected player knocked away in -X, got VX %v", player.VX)
	}
	if enemy.FlashTimer == 0 {
		t.Fatal("expected the enemy to flash after contact")
	}
	if !player.IsInvulnerable() {
		t.Fatal("expected the player to be invulnerable after contact")
	}
}

// Contact must never cost an enemy hits: only a projectile can cause a kill, and
// only a kill can delete a Pod.
func TestContact_NeverDamagesEnemyOrKills(t *testing.T) {
	killEvents := make(chan KillEvent, 8)
	scene := NewScene(make(chan GameEvent), killEvents)
	player := NewPlayer(400, 300)
	enemy := NewEnemy("pod-a", "default", 402, 300)
	scene.enemies = append(scene.enemies, enemy)

	for range 600 {
		scene.resolveContact(player)
		player.InvulnerableTimer = 0 // force every contact to land
	}

	if enemy.Hits != EnemyHits {
		t.Fatalf("expected enemy hits untouched at %d, got %d", EnemyHits, enemy.Hits)
	}
	if enemy.State != EnemyActive {
		t.Fatal("expected the enemy to survive repeated ramming")
	}
	if len(killEvents) != 0 {
		t.Fatalf("expected no kill events from contact, got %d", len(killEvents))
	}
}

func TestContact_InvulnerabilitySuppressesRepeatDamage(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	player := NewPlayer(400, 300)
	scene.enemies = append(scene.enemies, NewEnemy("pod-a", "default", 405, 300))

	scene.resolveContact(player)
	healthAfterFirst := player.Health

	// Overlapping for a whole second must cost only the one point.
	for range invulnerableTicks - 1 {
		player.updateWithControls(playerControls{})
		scene.resolveContact(player)
	}

	if player.Health != healthAfterFirst {
		t.Fatalf("expected health to stay at %d during invulnerability, got %d", healthAfterFirst, player.Health)
	}
}

func TestScene_RollSkipsContactAndBoltResolution(t *testing.T) {
	scene := NewScene(make(chan GameEvent, 8), make(chan KillEvent, 8))
	player := NewPlayer(400, 300)
	player.RollTicks = barrelRollDurationTicks
	scene.enemies = []*Enemy{NewEnemy("pod-a", "default", 405, 300)}
	scene.bolts = []*Bolt{NewBolt(400, 300, 0, ArchetypeTurret)}

	scene.Update(player)

	if player.Health != PlayerMaxHealth {
		t.Fatalf("expected rolling player not to take damage, got %d health", player.Health)
	}
	if player.VX != 0 || player.VY != 0 {
		t.Fatalf("expected rolling player not to be knocked back, got (%v, %v)", player.VX, player.VY)
	}
	if player.InvulnerableTimer != 0 {
		t.Fatalf("expected rolling not to start invulnerability, got %d", player.InvulnerableTimer)
	}
	if len(scene.Bolts()) != 1 {
		t.Fatalf("expected bolt to stay live through the roll, got %d", len(scene.Bolts()))
	}
	if len(scene.Bursts()) != 0 {
		t.Fatalf("expected no bolt burst during the roll, got %d", len(scene.Bursts()))
	}
}

func TestScene_RollOverlapDoesNotEndEnemyLunge(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	player := NewPlayer(400, 300)
	player.RollTicks = barrelRollDurationTicks
	enemy := NewEnemy("pod-a", "default", 405, 300)
	enemy.lunge.active = true
	scene.enemies = []*Enemy{enemy}

	scene.resolveContact(player)

	if !enemy.Lunging() {
		t.Fatal("expected a roll overlap not to end the lunge")
	}
}

func TestContact_DamageResumesAfterInvulnerabilityExpires(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	player := NewPlayer(400, 300)
	scene.enemies = append(scene.enemies, NewEnemy("pod-a", "default", 405, 300))

	scene.resolveContact(player)
	for range invulnerableTicks {
		player.updateWithControls(playerControls{})
	}

	if player.IsInvulnerable() {
		t.Fatal("expected invulnerability to have expired")
	}
	// Put the enemy back on top of the drifting player.
	scene.enemies[0].X, scene.enemies[0].Y = player.X+5, player.Y
	scene.resolveContact(player)

	if player.Health != PlayerMaxHealth-2 {
		t.Fatalf("expected health %d after a second contact, got %d", PlayerMaxHealth-2, player.Health)
	}
}

func TestKnockback_DecaysAndHomingResumes(t *testing.T) {
	enemy := NewEnemy("pod-a", "default", 500, 300)
	enemy.Knockback(1, 0)
	startVX := enemy.VX

	for range 120 {
		enemy.Update(400, 300, 0, 0) // player to the left, no neighbours
	}

	if math.Abs(enemy.VX) >= math.Abs(startVX) {
		t.Fatalf("expected knockback to decay from %v, got %v", startVX, enemy.VX)
	}
	if enemy.X >= 500 {
		t.Fatalf("expected the enemy to home back toward the player, got X %v", enemy.X)
	}
}

func TestWrap_PlayerReappearsOnOppositeEdge(t *testing.T) {
	player := NewPlayer(2, 300)
	player.Heading = math.Pi // face left
	player.VX = -10

	player.updateWithControls(playerControls{})

	if player.X < screenWidth/2 {
		t.Fatalf("expected the player to wrap to the right edge, got X %v", player.X)
	}

	player.Y = 2
	player.VX, player.VY = 0, -10
	player.updateWithControls(playerControls{})
	if player.Y < screenHeight/2 {
		t.Fatalf("expected the player to wrap to the bottom edge, got Y %v", player.Y)
	}
}

// In practice mode the Pod outlives its enemy's kill, so the enemy returns.
func TestPracticeMode_KilledEnemyRespawns(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.BeginSession(1, false)
	scene.recordKill(KillEvent{PodName: "pod-a", Namespace: "default"}, ArchetypeProbe, TierStandard, Symptoms{})

	if len(scene.enemies) != 0 {
		t.Fatalf("expected no enemy immediately after the kill, got %d", len(scene.enemies))
	}

	for range respawnDelayTicks {
		scene.updateRespawns()
	}

	if len(scene.enemies) != 1 {
		t.Fatalf("expected the enemy to respawn, got %d enemies", len(scene.enemies))
	}
	if scene.enemies[0].PodName != "pod-a" {
		t.Fatalf("expected the same pod to return, got %q", scene.enemies[0].PodName)
	}
	if scene.enemies[0].Hits != EnemyHits {
		t.Fatalf("expected a respawned enemy at full hits, got %d", scene.enemies[0].Hits)
	}
}

// In chaos mode the Pod is really gone, so its enemy must not come back.
func TestChaosMode_KilledEnemyDoesNotRespawn(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.BeginSession(1, true)
	scene.recordKill(KillEvent{PodName: "pod-a", Namespace: "default"}, ArchetypeProbe, TierStandard, Symptoms{})

	for range respawnDelayTicks * 2 {
		scene.updateRespawns()
	}

	if len(scene.enemies) != 0 {
		t.Fatalf("expected no respawn in chaos mode, got %d enemies", len(scene.enemies))
	}
}

// A despawn from the watch is authoritative: a pending respawn must be dropped.
func TestPracticeMode_DespawnCancelsPendingRespawn(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.BeginSession(1, false)
	scene.recordKill(KillEvent{PodName: "pod-a", Namespace: "default"}, ArchetypeProbe, TierStandard, Symptoms{})

	scene.despawnEnemy("pod-a", "default")

	for range respawnDelayTicks * 2 {
		scene.updateRespawns()
	}

	if len(scene.enemies) != 0 {
		t.Fatalf("expected a despawned pod not to respawn, got %d enemies", len(scene.enemies))
	}
}

func TestScene_KillsAreCounted(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.BeginSession(1, true)

	scene.recordKill(KillEvent{PodName: "pod-a", Namespace: "default"}, ArchetypeProbe, TierStandard, Symptoms{})
	scene.recordKill(KillEvent{PodName: "pod-b", Namespace: "default"}, ArchetypeProbe, TierStandard, Symptoms{})

	if scene.Kills() != 2 {
		t.Fatalf("expected 2 kills, got %d", scene.Kills())
	}
}

// Events buffered from an ended session must not repopulate a cleared pool.
func TestScene_DrainStaleEventsDiscardsSpawns(t *testing.T) {
	events := make(chan GameEvent, 8)
	scene := NewScene(events, make(chan KillEvent, 8))
	scene.BeginSession(1, false)

	events <- GameEvent{Type: EventSpawnEnemy, PodName: "pod-a", Namespace: "default"}
	scene.Reset()
	scene.DrainStaleEvents()

	if len(scene.enemies) != 0 {
		t.Fatalf("expected stale spawns to be discarded, got %d enemies", len(scene.enemies))
	}
}

func TestScene_ErrorEventIsSurfaced(t *testing.T) {
	events := make(chan GameEvent, 8)
	scene := NewScene(events, make(chan KillEvent, 8))

	events <- GameEvent{Type: EventError, Namespace: "default", Message: "connection refused"}
	scene.drainEvents()

	if scene.LastError() != "connection refused" {
		t.Fatalf("expected the error to be surfaced, got %q", scene.LastError())
	}
}

func TestPlayer_ResetRestoresStartingState(t *testing.T) {
	player := NewPlayer(400, 300)
	player.Health = 1
	player.VX, player.VY = 3, 3
	player.InvulnerableTimer = 30

	player.Reset(screenWidth/2, screenHeight/2)

	if player.Health != PlayerMaxHealth {
		t.Fatalf("expected full health after reset, got %d", player.Health)
	}
	if player.VX != 0 || player.VY != 0 {
		t.Fatalf("expected the ship stopped after reset, got (%v, %v)", player.VX, player.VY)
	}
	if player.IsInvulnerable() {
		t.Fatal("expected invulnerability cleared after reset")
	}
}
