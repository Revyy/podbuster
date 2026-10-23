package game

import (
	"math"
	"testing"
)

func boltScene() (*Scene, *Player) {
	scene := NewScene(make(chan GameEvent, 8), make(chan KillEvent, 8))
	scene.BeginSession(1, false)
	return scene, NewPlayer(400, 300)
}

func TestBolt_MovesAtBoltSpeedAndExpiresAtEdge(t *testing.T) {
	scene, player := boltScene()
	scene.bolts = []*Bolt{NewBolt(790, 100, 0, ArchetypeTurret)}

	scene.Update(player)
	if b := scene.Bolts()[0]; math.Abs(b.X-(790+boltSpeed)) > 1e-9 {
		t.Fatalf("expected X %v, got %v", 790+boltSpeed, b.X)
	}
	step(scene, player, 5)
	if len(scene.Bolts()) != 0 {
		t.Fatal("expected the Bolt to expire at the screen edge")
	}
}

func TestBolt_HitCostsOneHealthWithHalfKnockbackAndBurst(t *testing.T) {
	scene, player := boltScene()
	scene.bolts = []*Bolt{NewBolt(390, 300, 0, ArchetypeTurret)}

	scene.Update(player)

	if player.Health != PlayerMaxHealth-1 {
		t.Fatalf("expected 1 Health lost, got %d", PlayerMaxHealth-player.Health)
	}
	if want := playerKnockbackSpeed * boltKnockbackScale; math.Abs(player.VX-want) > 1e-9 {
		t.Fatalf("expected knockback %v along the Bolt, got %v", want, player.VX)
	}
	if !player.IsInvulnerable() {
		t.Fatal("expected a Bolt hit to start Invulnerability")
	}
	if len(scene.Bolts()) != 0 || len(scene.Bursts()) != 1 {
		t.Fatalf("expected Bolt consumed and one burst, got %d/%d", len(scene.Bolts()), len(scene.Bursts()))
	}
}

func TestBolt_PassesThroughDuringInvulnerabilityAndStaysLive(t *testing.T) {
	scene, player := boltScene()
	player.InvulnerableTimer = invulnerableTicks
	scene.bolts = []*Bolt{NewBolt(400, 300, 0, ArchetypeTurret)}

	scene.Update(player)

	if player.Health != PlayerMaxHealth {
		t.Fatal("Invulnerable player took damage")
	}
	if len(scene.Bolts()) != 1 {
		t.Fatal("expected the Bolt to stay live")
	}
}

func TestBolt_SharesInvulnerabilityWithContact(t *testing.T) {
	scene, player := boltScene()
	scene.enemies = []*Enemy{NewEnemy("a", "default", 400, 300)}
	scene.bolts = []*Bolt{NewBolt(400, 300, 0, ArchetypeTurret)}

	scene.Update(player)

	if player.Health != PlayerMaxHealth-1 {
		t.Fatalf("expected contact and Bolt to share one window, lost %d", PlayerMaxHealth-player.Health)
	}
}

func TestBolt_PassesThroughEnemiesAndCannotBeShotDown(t *testing.T) {
	scene, player := boltScene()
	enemy := NewEnemy("a", "default", 100, 100)
	scene.enemies = []*Enemy{enemy}
	scene.bolts = []*Bolt{NewBolt(98, 100, 0, ArchetypeTurret)}
	scene.projectiles = []*Projectile{NewProjectile(101, 100, 0, 0, 1)}

	scene.Update(player)

	if len(scene.Bolts()) != 1 {
		t.Fatal("expected the Bolt to survive enemies and projectiles")
	}
	if enemy.Hits != EnemyHits-1 {
		t.Fatalf("Bolt harmed an enemy: hits %d", enemy.Hits)
	}
}

func TestBolt_CannotHitDuringDyingAndIsClearedOnReset(t *testing.T) {
	scene, player := boltScene()
	scene.bolts = []*Bolt{NewBolt(400, 300, 0, ArchetypeTurret)}

	scene.UpdateWithoutPlayer(player.X, player.Y)
	if player.Health != PlayerMaxHealth {
		t.Fatal("Bolt hit while the player was dying")
	}
	if len(scene.Bolts()) != 1 || scene.Bolts()[0].X != 400+boltSpeed {
		t.Fatal("expected the Bolt to keep moving")
	}

	scene.ClearBolts()
	if len(scene.Bolts()) != 0 {
		t.Fatal("expected ClearBolts to remove Bolts")
	}
	scene.bolts = []*Bolt{NewBolt(1, 1, 0, ArchetypeTurret)}
	scene.Reset()
	if len(scene.Bolts()) != 0 {
		t.Fatal("expected Reset to clear Bolts")
	}
}

func TestTurret_NoNewVolleysDuringDying(t *testing.T) {
	scene, _, player := turretScene(TierLight, 0)
	for range 400 {
		scene.UpdateWithoutPlayer(player.X, player.Y)
	}
	if len(scene.Bolts()) != 0 {
		t.Fatal("Turret fired while the player was dying")
	}
}
