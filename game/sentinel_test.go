package game

import (
	"math"
	"testing"
)

// sentinelScene places a Sentinel on its orbit around a stationary player.
// The RNG offset is the first-dash delay beyond one interval; dirPick 1 orbits
// the opposite way.
func sentinelScene(tier Tier, offset, dirPick int) (*Scene, *Enemy, *Player) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.rng = sentinelRNG{offset: offset, dir: dirPick}
	p := &Player{X: 400, Y: 300, Health: 5, InvulnerableTimer: 1 << 30}
	scene.spawnEnemy("api-0", "default", ArchetypeSentinel, tier)
	s := scene.enemies[0]
	s.X, s.Y = p.X+tier.sentinelStats().radius, p.Y
	return scene, s, p
}

// sentinelRNG answers the direction pick (IntN(2)) and the offset pick separately.
type sentinelRNG struct{ offset, dir int }

func (r sentinelRNG) IntN(n int) int {
	if n == 2 {
		return r.dir
	}
	return min(r.offset, n-1)
}
func (sentinelRNG) Float64() float64 { return 0.5 }

func dist(e *Enemy, p *Player) float64 { return math.Hypot(e.X-p.X, e.Y-p.Y) }

func TestSentinel_OrbitsAtTierRadiusAtCruiseSpeed(t *testing.T) {
	for _, tier := range []Tier{TierLight, TierStandard, TierHeavy} {
		scene, s, p := sentinelScene(tier, 0, 0)
		r := tier.sentinelStats().radius
		var travelled float64
		// Stay within the telegraph-free part of the first interval.
		for i := 0; i < 100; i++ {
			x, y := s.X, s.Y
			scene.Update(p)
			travelled += math.Hypot(s.X-x, s.Y-y)
			if d := dist(s, p); math.Abs(d-r) > 1 {
				t.Fatalf("tier %v tick %d: radius %v, want ~%v", tier, i, d, r)
			}
		}
		want := 100 * s.CruiseSpeed()
		if math.Abs(travelled-want) > want*0.05 {
			t.Fatalf("tier %v travelled %v, want ~%v", tier, travelled, want)
		}
	}
}

func TestSentinel_OrbitDirectionFromRNG(t *testing.T) {
	cross := func(dirPick int) float64 {
		scene, s, p := sentinelScene(TierStandard, 0, dirPick)
		rx, ry := s.X-p.X, s.Y-p.Y
		scene.Update(p)
		return rx*(s.Y-p.Y) - ry*(s.X-p.X)
	}
	if a, b := cross(0), cross(1); a*b >= 0 {
		t.Fatalf("expected opposite orbit directions, got %v and %v", a, b)
	}
}

func TestSentinel_CruiseScalesWithTier(t *testing.T) {
	_, light, _ := sentinelScene(TierLight, 0, 0)
	_, heavy, _ := sentinelScene(TierHeavy, 0, 0)
	if light.CruiseSpeed() <= heavy.CruiseSpeed() {
		t.Fatalf("light %v should cruise faster than heavy %v", light.CruiseSpeed(), heavy.CruiseSpeed())
	}
}

func TestSentinel_DashCadenceAndTelegraph(t *testing.T) {
	for _, tier := range []Tier{TierLight, TierStandard, TierHeavy} {
		for _, offset := range []int{0, 37} {
			scene, s, p := sentinelScene(tier, offset, 0)
			interval := tier.sentinelStats().interval
			first := interval + offset
			var starts []int
			for tk := 1; tk <= first+interval+5; tk++ {
				was := s.Lunging()
				scene.Update(p)
				if !was && s.Lunging() {
					starts = append(starts, tk)
				}
				if tk == first-sentinelTelegraphTicks-1 && s.Telegraph != 0 {
					t.Fatalf("tier %v: telegraph before window", tier)
				}
				if tk == first-1 && s.Telegraph < 0.9 {
					t.Fatalf("tier %v: ring not near full before dash: %v", tier, s.Telegraph)
				}
			}
			if len(starts) != 2 || starts[0] != first || starts[1] != first+interval {
				t.Fatalf("tier %v offset %d: dash ticks %v, want [%d %d]", tier, offset, starts, first, first+interval)
			}
		}
	}
}

func TestSentinel_DashLocksPositionAndOvershoots(t *testing.T) {
	scene, s, p := sentinelScene(TierStandard, 0, 0)
	for !s.Lunging() {
		scene.Update(p)
	}
	startX, startY := s.X, s.Y
	lockX, lockY := p.X, p.Y
	p.X, p.Y = 100, 500 // moving afterwards must not redirect the dash
	var maxSpeed float64
	for s.Lunging() {
		x, y := s.X, s.Y
		scene.Update(p)
		maxSpeed = math.Max(maxSpeed, math.Hypot(s.X-x, s.Y-y))
	}
	if maxSpeed < 2.69 || maxSpeed > 2.7+1e-9 {
		t.Fatalf("dash speed %v, want 2.7", maxSpeed)
	}
	// The Sentinel only ever travels the locked line, ending 40 px past it.
	want := math.Hypot(lockX-startX, lockY-startY) + 40
	if got := math.Hypot(s.X-startX, s.Y-startY); math.Abs(got-want) > 1e-6 {
		t.Fatalf("dash length %v, want %v", got, want)
	}
	cross := (lockX-startX)*(s.Y-startY) - (lockY-startY)*(s.X-startX)
	if math.Abs(cross) > 1e-6 {
		t.Fatalf("dash left the locked line, cross %v", cross)
	}
}

func TestSentinel_RejoinsOrbitAfterDash(t *testing.T) {
	scene, s, p := sentinelScene(TierStandard, 0, 0)
	for !s.Lunging() {
		scene.Update(p)
	}
	for s.Lunging() {
		scene.Update(p)
	}
	after := dist(s, p)
	tick(scene, p, 40)
	if dist(s, p) <= after+10 {
		t.Fatalf("expected to climb back toward the orbit: %v then %v", after, dist(s, p))
	}
}

func TestSentinel_ContactEndsDash(t *testing.T) {
	scene, s, p := sentinelScene(TierStandard, 0, 0)
	for !s.Lunging() {
		scene.Update(p)
	}
	p.InvulnerableTimer = 0
	for i := 0; i < 200 && p.Health == 5; i++ {
		scene.Update(p)
	}
	if p.Health >= 5 {
		t.Fatal("dash never reached the player")
	}
	if s.Lunging() {
		t.Fatal("contact should end the dash")
	}
}

func TestSentinel_EdgeStopsDash(t *testing.T) {
	scene, s, p := sentinelScene(TierStandard, 0, 0)
	p.X, p.Y = 700, 300
	s.X, s.Y = 700-180, 300
	for !s.Lunging() {
		scene.Update(p)
	}
	p.X = 790
	for s.Lunging() {
		scene.Update(p)
	}
	if s.X > screenWidth-s.Size()/2+1e-9 {
		t.Fatalf("dash left the screen: x=%v", s.X)
	}
}

func TestSentinel_RevertRestartsTimers(t *testing.T) {
	scene, s, p := sentinelScene(TierStandard, 0, 1)
	tick(scene, p, 160)
	dir := s.sentinel.dir
	s.BeginTerminating(0)
	s.RevertToActive()
	if s.Telegraph != 0 || s.Lunging() {
		t.Fatal("revert should clear telegraph and lunge")
	}
	for i := 0; i < 180-1; i++ {
		scene.Update(p)
		if s.Lunging() {
			t.Fatalf("dashed at tick %d after revert", i+1)
		}
	}
	scene.Update(p)
	if !s.Lunging() {
		t.Fatal("expected a dash one full interval after revert")
	}
	if s.sentinel.dir != dir {
		t.Fatal("revert should keep the orbit direction")
	}
}

func TestSentinel_PracticeRespawnStartsFresh(t *testing.T) {
	scene, _, _ := sentinelScene(TierHeavy, 11, 1)
	scene.respawns = []pendingRespawn{{podName: "b-0", namespace: "default", archetype: ArchetypeSentinel, tier: TierHeavy, timer: 1}}
	scene.updateRespawns()
	e := scene.enemies[len(scene.enemies)-1]
	if e.Tier != TierHeavy || e.sentinel.timer != 120+11 || e.sentinel.dir != -1 {
		t.Fatalf("unexpected respawn cycle %+v tier %v", e.sentinel, e.Tier)
	}
}
