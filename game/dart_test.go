package game

import (
	"math"
	"testing"
)

func dartScene(tier Tier, x, y float64) (*Scene, *Enemy, *Player) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	d := NewEnemyOf("job-a", "default", ArchetypeDart, tier, x, y)
	scene.enemies = []*Enemy{d}
	return scene, d, &Player{X: 700, Y: 300, Health: 5, InvulnerableTimer: 1 << 30}
}

func tick(s *Scene, p *Player, n int) {
	for range n {
		s.Update(p)
	}
}

func TestDart_WindupDriftsSlowlyAndGlows(t *testing.T) {
	for _, tier := range []Tier{TierLight, TierStandard, TierHeavy} {
		scene, d, p := dartScene(tier, 100, 300)
		wind := tier.dartStats().windup
		for i := 1; i < wind; i++ {
			x, y := d.X, d.Y
			scene.Update(p)
			if got := math.Hypot(d.X-x, d.Y-y); got > 0.8+1e-9 {
				t.Fatalf("tier %v tick %d: drift %v exceeds 0.8", tier, i, got)
			}
			if d.Lunging() {
				t.Fatalf("tier %v lunged during wind-up at tick %d", tier, i)
			}
		}
		if d.Telegraph <= 0.9 {
			t.Fatalf("tier %v expected eye glow near full, got %v", tier, d.Telegraph)
		}
		scene.Update(p)
		if !d.Lunging() {
			t.Fatalf("tier %v expected charge after %d ticks", tier, wind)
		}
	}
}

func TestDart_ChargesThroughLockedPositionAndOvershoots(t *testing.T) {
	scene, d, p := dartScene(TierStandard, 100, 300)
	p.X = 400
	tick(scene, p, 40)
	lockX, lockY := p.X, p.Y
	startX := d.X
	p.X, p.Y = 300, 100 // moving afterwards must not redirect the charge
	var maxSpeed float64
	for d.Lunging() || d.dart.phase == dartCharge {
		x, y := d.X, d.Y
		scene.Update(p)
		maxSpeed = math.Max(maxSpeed, math.Hypot(d.X-x, d.Y-y))
	}
	if maxSpeed < 3.49 || maxSpeed > 3.5+1e-9 {
		t.Fatalf("expected charge speed 3.5, got %v", maxSpeed)
	}
	want := lockX - startX + 120
	if math.Abs(d.X-(startX+want)) > 1e-6 || math.Abs(d.Y-lockY) > 1e-6 {
		t.Fatalf("expected to end at (%v,%v), got (%v,%v)", startX+want, lockY, d.X, d.Y)
	}
}

func TestDart_BrakesThenRepeats(t *testing.T) {
	scene, d, p := dartScene(TierStandard, 100, 300)
	tick(scene, p, 40)
	for d.dart.phase != dartBrake {
		scene.Update(p)
	}
	x, y := d.X, d.Y
	for i := range dartBrakeTicks {
		if d.dart.phase != dartBrake {
			t.Fatalf("left brake early at tick %d", i)
		}
		scene.Update(p)
	}
	if d.X != x || d.Y != y {
		t.Fatalf("expected Dart to hold still while braking")
	}
	if d.dart.phase != dartWindup || d.Telegraph != 0 {
		t.Fatalf("expected a new wind-up, phase %v telegraph %v", d.dart.phase, d.Telegraph)
	}
}

func TestDart_ChargeStopsAtInsetScreenEdge(t *testing.T) {
	scene, d, p := dartScene(TierHeavy, 600, 300)
	p.X, p.Y = 780, 300 // overshoot would leave the screen
	d.dart.timer = 100
	scene.Update(p)
	p.Y = 50 // the target is locked; keep the ship out of the charge path
	for i := 0; i < 100 && d.dart.phase != dartBrake; i++ {
		scene.Update(p)
	}
	if want := screenWidth - d.Size()/2; math.Abs(d.X-want) > 1e-6 {
		t.Fatalf("expected stop at x=%v, got %v", want, d.X)
	}
}

func TestDart_ContactEndsLungeWithNormalDamageAndKnockback(t *testing.T) {
	scene, d, p := dartScene(TierHeavy, 100, 300)
	p.X = 300
	p.InvulnerableTimer = 0
	tick(scene, p, 30)
	for i := 0; i < 200 && p.Health == 5; i++ {
		scene.Update(p)
	}
	if p.Health != 5-d.ContactDamage() {
		t.Fatalf("expected %d contact damage, health %d", d.ContactDamage(), p.Health)
	}
	if d.Lunging() {
		t.Fatalf("contact should end the lunge")
	}
	if d.VX == 0 || d.FlashTimer == 0 {
		t.Fatalf("expected knockback and flash")
	}
}

func TestDart_FloorBlockingEndsLungeEarly(t *testing.T) {
	scene, d, p := dartScene(TierStandard, 100, 300)
	p.X, p.Y = 700, 300
	tick(scene, p, 40)
	if !d.Lunging() {
		t.Fatal("expected lunge")
	}
	blocker := NewEnemy("pod-b", "default", d.X+60, d.Y)
	scene.enemies = append(scene.enemies, blocker)
	tick(scene, p, 40)
	if d.Lunging() {
		t.Fatal("expected the no-overlap floor to end the lunge")
	}
	if d.X > 500 {
		t.Fatalf("lunge should have ended early, reached %v", d.X)
	}
}

func TestDart_LungeSkipsSoftAvoidance(t *testing.T) {
	scene, d, p := dartScene(TierStandard, 100, 300)
	p.Y = 300
	tick(scene, p, 40)
	// A neighbour inside avoid range but outside the hard floor, off the path.
	scene.enemies = append(scene.enemies, NewEnemy("pod-b", "default", d.X+30, d.Y+EnemySeparation+5))
	y := d.Y
	scene.Update(p)
	if d.Y != y {
		t.Fatalf("lunge was steered by avoidance: y %v -> %v", y, d.Y)
	}
}

func TestDart_RevertRestartsCycle(t *testing.T) {
	scene, d, p := dartScene(TierStandard, 100, 300)
	tick(scene, p, 41)
	d.BeginTerminating(0)
	d.RevertToActive()
	if d.Lunging() || d.dart != (dartCycle{}) || d.Telegraph != 0 {
		t.Fatalf("expected a fresh cycle after revert: %+v", d.dart)
	}
}

func TestDart_FreshOnRespawnAndOffscreenWindsUp(t *testing.T) {
	d := NewEnemyOf("job-a", "default", ArchetypeDart, TierLight, -50, 300)
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.enemies = []*Enemy{d}
	p := &Player{X: 700, Y: 300, Health: 5}
	if d.dart != (dartCycle{}) {
		t.Fatal("new Dart should start fresh")
	}
	tick(scene, p, 60)
	if d.Lunging() && d.X < 0 {
		t.Fatal("offscreen Dart must not lunge")
	}
}

func TestDart_ContactEndsLungeEvenWhileInvulnerable(t *testing.T) {
	scene, d, p := dartScene(TierHeavy, 100, 300)
	p.X = 300
	tick(scene, p, 30)
	for i := 0; i < 200 && !d.Lunging(); i++ {
		scene.Update(p)
	}
	if !d.Lunging() {
		t.Fatal("expected a lunge")
	}
	d.X, d.Y = p.X, p.Y
	scene.Update(p)
	if p.Health != 5 {
		t.Fatalf("invulnerable player took damage: %d", p.Health)
	}
	if d.Lunging() {
		t.Fatal("contact should end the lunge even during Invulnerability")
	}
}
