package game

import (
	"math"
	"testing"
)

func TestEnemyTierStats(t *testing.T) {
	tests := []struct {
		archetype Archetype
		tier      Tier
		hits      int
		size      float64
		speed     float64
		contact   int
	}{
		{ArchetypeProbe, TierLight, 2, 0.85, 1.1, 1},
		{ArchetypeProbe, TierStandard, 3, 1, 1, 1},
		{ArchetypeProbe, TierHeavy, 5, 1.3, 0.85, 2},
		{ArchetypeTurret, TierHeavy, 5, 1.3, 0.85, 2},
		{ArchetypeSentinel, TierLight, 2, 0.85, 1.1, 1},
		{ArchetypeDart, TierLight, 1, 0.85, 1.1, 1},
		{ArchetypeDart, TierStandard, 2, 1, 1, 1},
		{ArchetypeDart, TierHeavy, 3, 1.3, 0.85, 2},
	}
	for _, tt := range tests {
		e := NewEnemyOf("p", "ns", tt.archetype, tt.tier, 400, 300)
		if e.Hits != tt.hits {
			t.Errorf("%v/%v hits = %d, want %d", tt.archetype, tt.tier, e.Hits, tt.hits)
		}
		if math.Abs(e.Bounds().W-EnemySize*tt.size) > 1e-9 || e.Bounds().W != e.Bounds().H {
			t.Errorf("%v/%v hitbox = %+v, want side %v", tt.archetype, tt.tier, e.Bounds(), EnemySize*tt.size)
		}
		if math.Abs(e.CruiseSpeed()-EnemySpeed*tt.speed) > 1e-9 {
			t.Errorf("%v/%v cruise = %v, want %v", tt.archetype, tt.tier, e.CruiseSpeed(), EnemySpeed*tt.speed)
		}
		if e.ContactDamage() != tt.contact {
			t.Errorf("%v/%v contact = %d, want %d", tt.archetype, tt.tier, e.ContactDamage(), tt.contact)
		}
	}
}

func TestEnemyCruisesAtTierSpeed(t *testing.T) {
	e := NewEnemyOf("p", "ns", ArchetypeProbe, TierHeavy, 400, 300)
	e.Update(800, 300, 0, 0)
	if math.Abs(e.X-400-EnemySpeed*0.85) > 1e-9 {
		t.Fatalf("expected a heavy enemy to move %v, moved %v", EnemySpeed*0.85, e.X-400)
	}
}

func TestSpawnEventCarriesArchetypeAndTier(t *testing.T) {
	events := make(chan GameEvent, 1)
	scene := NewScene(events, make(chan KillEvent, 1))
	events <- GameEvent{Type: EventSpawnEnemy, PodName: "db-0", Namespace: "ns", Archetype: ArchetypeTurret, Tier: TierHeavy}
	scene.drainEvents()
	if len(scene.enemies) != 1 {
		t.Fatalf("expected one enemy, got %d", len(scene.enemies))
	}
	e := scene.enemies[0]
	if e.Archetype != ArchetypeTurret || e.Tier != TierHeavy || e.Hits != 5 {
		t.Fatalf("unexpected enemy: %+v", e)
	}
}

func TestHeavyContactCostsTwoHealth(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 1))
	player := NewPlayer(400, 300)
	scene.enemies = append(scene.enemies, NewEnemyOf("p", "ns", ArchetypeProbe, TierHeavy, 405, 300))
	scene.resolveContact(player)
	if player.Health != PlayerMaxHealth-2 {
		t.Fatalf("expected health %d, got %d", PlayerMaxHealth-2, player.Health)
	}
}

func TestHitboxScalesWithTier(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 1))
	heavy := NewEnemyOf("p", "ns", ArchetypeProbe, TierHeavy, 400, 300)
	light := NewEnemyOf("q", "ns", ArchetypeProbe, TierLight, 100, 300)
	scene.enemies = []*Enemy{heavy, light}
	// 14px from centre is outside a Standard box (12) but inside a Heavy one (15.6).
	scene.projectiles = []*Projectile{{X: 414, Y: 300, Active: true, HitCost: 1}}
	scene.resolveCollisions()
	if heavy.Hits != 4 {
		t.Fatalf("expected the heavy enemy to be hit, hits=%d", heavy.Hits)
	}
	// 11px is inside a Standard box but outside a Light one (10.2).
	scene.projectiles = []*Projectile{{X: 111, Y: 300, Active: true, HitCost: 1}}
	scene.resolveCollisions()
	if light.Hits != 2 {
		t.Fatalf("expected the light enemy to be missed, hits=%d", light.Hits)
	}
}

func TestSeparationUsesEachPairsSizes(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 1))
	heavyA := NewEnemyOf("a", "ns", ArchetypeProbe, TierHeavy, 400, 300)
	heavyB := NewEnemyOf("b", "ns", ArchetypeProbe, TierHeavy, 402, 300)
	lightA := NewEnemyOf("c", "ns", ArchetypeProbe, TierLight, 100, 300)
	lightB := NewEnemyOf("d", "ns", ArchetypeProbe, TierLight, 102, 300)
	scene.enemies = []*Enemy{heavyA, heavyB, lightA, lightB}
	scene.separate()

	if got := heavyB.X - heavyA.X; got < EnemySeparation*1.3-separationEpsilon {
		t.Fatalf("expected heavy pair at least %v apart, got %v", EnemySeparation*1.3, got)
	}
	got := lightB.X - lightA.X
	if got < EnemySeparation*0.85-separationEpsilon || got > EnemySeparation*0.85+1 {
		t.Fatalf("expected light pair about %v apart, got %v", EnemySeparation*0.85, got)
	}
}

func TestPracticeRespawnKeepsArchetypeAndTier(t *testing.T) {
	events := make(chan GameEvent, 1)
	scene := NewScene(events, make(chan KillEvent, 8))
	scene.BeginSession(1, false)
	events <- GameEvent{Type: EventSpawnEnemy, PodName: "job-1", Namespace: "ns", Archetype: ArchetypeDart, Tier: TierHeavy}
	scene.drainEvents()

	e := scene.enemies[0]
	e.X, e.Y = 400, 300
	scene.projectiles = []*Projectile{{X: 400, Y: 300, Active: true, HitCost: 3}}
	scene.resolveCollisions()
	scene.enemies = nil

	for range respawnDelayTicks {
		scene.updateRespawns()
	}
	if len(scene.enemies) != 1 {
		t.Fatalf("expected a respawn, got %d enemies", len(scene.enemies))
	}
	r := scene.enemies[0]
	if r.Archetype != ArchetypeDart || r.Tier != TierHeavy || r.Hits != 3 {
		t.Fatalf("respawn lost its identity: %+v", r)
	}
}
