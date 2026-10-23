package game

import (
	"math"
	"testing"
)

type fixedRNG struct{ n int }

func (f fixedRNG) IntN(int) int     { return f.n }
func (f fixedRNG) Float64() float64 { return 0.5 }

func turretScene(tier Tier, offset int) (*Scene, *Enemy, *Player) {
	scene := NewScene(make(chan GameEvent, 8), make(chan KillEvent, 8))
	scene.BeginSession(1, false)
	scene.rng = fixedRNG{offset}
	turret := NewEnemyOf("db-0", "default", ArchetypeTurret, tier, 100, 100)
	scene.enemies = []*Enemy{turret}
	return scene, turret, NewPlayer(700, 500)
}

func step(scene *Scene, player *Player, n int) {
	for range n {
		scene.Update(player)
	}
}

func TestTurret_FirstVolleyAfterIntervalPlusOffset(t *testing.T) {
	scene, _, player := turretScene(TierStandard, 20)

	step(scene, player, 319)
	if len(scene.Bolts()) != 0 {
		t.Fatalf("expected no Bolts before interval+offset, got %d", len(scene.Bolts()))
	}
	step(scene, player, 1)
	if len(scene.Bolts()) != 6 {
		t.Fatalf("expected a 6-way ring at tick 320, got %d Bolts", len(scene.Bolts()))
	}
}

func TestTurret_LightFiresFixedFourWayRing(t *testing.T) {
	scene, _, player := turretScene(TierLight, 0)

	step(scene, player, 376)
	if len(scene.Bolts()) != 4 {
		t.Fatalf("expected 4 Bolts, got %d", len(scene.Bolts()))
	}
	assertAngles(t, scene.Bolts(), 0, 4)

	scene.bolts = nil
	step(scene, player, 376)
	assertAngles(t, scene.Bolts(), 0, 4)
}

func TestTurret_StandardRingRotatesThirtyDegreesPerVolley(t *testing.T) {
	scene, _, player := turretScene(TierStandard, 0)

	step(scene, player, 300)
	assertAngles(t, scene.Bolts(), 0, 6)
	scene.bolts = nil
	step(scene, player, 300)
	assertAngles(t, scene.Bolts(), math.Pi/6, 6)
}

func TestTurret_HeavyAddsAimedBoltSixtyTicksAfterRing(t *testing.T) {
	scene, _, player := turretScene(TierHeavy, 0)

	step(scene, player, 300)
	if len(scene.Bolts()) != 6 {
		t.Fatalf("expected the ring, got %d Bolts", len(scene.Bolts()))
	}
	scene.bolts = nil
	step(scene, player, 59)
	if len(scene.Bolts()) != 0 {
		t.Fatal("aimed Bolt fired early")
	}
	step(scene, player, 1)
	if len(scene.Bolts()) != 1 {
		t.Fatalf("expected one aimed Bolt, got %d", len(scene.Bolts()))
	}
	b := scene.Bolts()[0]
	want := math.Atan2(player.Y-100, player.X-100)
	if got := math.Atan2(b.VY, b.VX); math.Abs(got-want) > 1e-9 {
		t.Fatalf("aimed Bolt angle %v, want %v", got, want)
	}
}

func TestTurret_TelegraphsBeforeRingAndAimedBolt(t *testing.T) {
	scene, turret, player := turretScene(TierHeavy, 0)

	step(scene, player, 239)
	if turret.Telegraph != 0 {
		t.Fatalf("expected no telegraph before the last 60 ticks, got %v", turret.Telegraph)
	}
	step(scene, player, 1)
	if turret.Telegraph <= 0 {
		t.Fatal("expected the ring telegraph to start 60 ticks out")
	}
	step(scene, player, 60)
	scene.bolts = nil
	step(scene, player, 44)
	if turret.Telegraph != 0 {
		t.Fatalf("expected no telegraph before the aimed window, got %v", turret.Telegraph)
	}
	step(scene, player, 1)
	if turret.Telegraph <= 0 {
		t.Fatal("expected the aimed telegraph to start 15 ticks out")
	}
}

func TestTurret_NeverMovesAndContactPushesOnlyPlayer(t *testing.T) {
	scene, turret, _ := turretScene(TierStandard, 0)
	player := NewPlayer(turret.X+10, turret.Y)

	step(scene, player, 30)

	if turret.X != 100 || turret.Y != 100 || turret.VX != 0 || turret.VY != 0 {
		t.Fatalf("turret moved to (%v,%v) v(%v,%v)", turret.X, turret.Y, turret.VX, turret.VY)
	}
	if player.Health == PlayerMaxHealth {
		t.Fatal("expected contact to damage the player")
	}
}

func TestTurret_SeparationMakesOthersGiveWay(t *testing.T) {
	scene, turret, player := turretScene(TierStandard, 0)
	probe := NewEnemy("web-0", "default", turret.X+5, turret.Y)
	scene.enemies = append(scene.enemies, probe)

	scene.Update(player)

	if turret.X != 100 || turret.Y != 100 {
		t.Fatalf("turret was pushed to (%v,%v)", turret.X, turret.Y)
	}
	if d := math.Hypot(probe.X-turret.X, probe.Y-turret.Y); d < pairSeparation(turret, probe) {
		t.Fatalf("probe still inside separation: %v", d)
	}
}

func TestTurret_TerminatingStopsFiringButBoltsPersist(t *testing.T) {
	scene, turret, player := turretScene(TierStandard, 0)
	step(scene, player, 300)
	inFlight := len(scene.Bolts())

	turret.BeginTerminating(0)
	step(scene, player, 310)

	if len(scene.Bolts()) > inFlight {
		t.Fatal("terminating Turret fired a new volley")
	}
	scene.bolts = nil
	scene.bolts = append(scene.bolts, NewBolt(400, 300, 0, ArchetypeTurret))
	turret.BeginExploding(enemyAnimationTicks)
	scene.Update(player)
	if len(scene.Bolts()) != 1 {
		t.Fatal("expected an in-flight Bolt to outlive its Turret")
	}
}

func TestTurret_RevertRestartsCycle(t *testing.T) {
	events := make(chan GameEvent, 8)
	scene, turret, player := turretScene(TierStandard, 0)
	scene.gameEvents = events
	step(scene, player, 100)
	turret.BeginTerminating(0)
	scene.Update(player)

	events <- GameEvent{Type: EventDeletionFailed, PodName: "db-0", Namespace: "default"}
	scene.bolts = nil
	step(scene, player, 299)
	if len(scene.Bolts()) != 0 {
		t.Fatal("reverted Turret fired before a full interval")
	}
	step(scene, player, 1)
	if len(scene.Bolts()) == 0 {
		t.Fatal("reverted Turret never fired")
	}
}

func TestTurret_SpawnsOnScreenAndRespawnStartsFresh(t *testing.T) {
	scene, _, player := turretScene(TierStandard, 0)
	scene.enemies = nil
	scene.spawnEnemy("db-1", "default", ArchetypeTurret, TierStandard)
	e := scene.enemies[0]
	if e.X < 0 || e.X > screenWidth || e.Y < 0 || e.Y > screenHeight {
		t.Fatalf("turret spawned off-screen at (%v,%v)", e.X, e.Y)
	}
	step(scene, player, 299)
	if len(scene.Bolts()) != 0 {
		t.Fatal("fresh Turret fired early")
	}
}

func TestTurret_FieldGuidePortraitGlowsOnTheRealCycle(t *testing.T) {
	for _, tier := range []Tier{TierLight, TierStandard, TierHeavy} {
		scene, turret, player := turretScene(tier, 0)
		interval := turretPatternOf(tier).interval
		step(scene, player, interval)
		for tick := interval + 1; tick <= 3*interval; tick++ {
			scene.Update(player)
			if got, want := guideTurretCharge(tier, tick), turret.Telegraph; got != want {
				t.Fatalf("tier %v tick %d: portrait glow %v, Turret Telegraph %v", tier, tick, got, want)
			}
		}
	}
}

func assertAngles(t *testing.T, bolts []*Bolt, start float64, n int) {
	t.Helper()
	if len(bolts) != n {
		t.Fatalf("expected %d Bolts, got %d", n, len(bolts))
	}
	for i, b := range bolts {
		want := start + 2*math.Pi*float64(i)/float64(n)
		got := math.Atan2(b.VY, b.VX)
		if math.Abs(math.Remainder(got-want, 2*math.Pi)) > 1e-9 {
			t.Fatalf("bolt %d angle %v, want %v", i, got, want)
		}
	}
}

func TestTurret_TurretsNeverOverlap(t *testing.T) {
	scene, a, player := turretScene(TierStandard, 0)
	b := NewEnemyOf("db-1", "default", ArchetypeTurret, TierHeavy, a.X+3, a.Y)
	scene.enemies = append(scene.enemies, b)

	scene.Update(player)

	if d := math.Hypot(a.X-b.X, a.Y-b.Y); d < pairSeparation(a, b) {
		t.Fatalf("turrets overlap: %v < %v", d, pairSeparation(a, b))
	}
	if a.VX != 0 || a.VY != 0 || b.VX != 0 || b.VY != 0 {
		t.Fatal("separation must not give turrets velocity")
	}
}
