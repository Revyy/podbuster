package game

import (
	"math"
	"testing"
)

// Enemies stand for distinct Pods, so they must stay distinct targets: a pile
// of overlapping enemies would absorb projectiles and contact as if it were one.
func TestSeparation_OverlappingEnemiesArePushedApart(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.enemies = []*Enemy{
		NewEnemy("pod-a", "default", 400, 300),
		NewEnemy("pod-b", "default", 404, 300),
	}

	scene.separate()

	got := math.Hypot(
		scene.enemies[0].X-scene.enemies[1].X,
		scene.enemies[0].Y-scene.enemies[1].Y,
	)
	if got < EnemySeparation {
		t.Fatalf("expected enemies at least %v apart, got %v", EnemySeparation, got)
	}
}

// A killed enemy no longer stands for a Pod, so the pool must not reserve
// space for it.
func TestSeparation_NonActiveEnemiesExertNoForce(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	dead := NewEnemy("pod-a", "default", 400, 300)
	dead.State = EnemyExploding
	live := NewEnemy("pod-b", "default", 402, 300)
	scene.enemies = []*Enemy{dead, live}

	scene.separate()

	if live.X != 402 || live.Y != 300 {
		t.Fatalf("expected a dead enemy to push nobody, live enemy moved to (%v, %v)", live.X, live.Y)
	}
	if dead.X != 400 || dead.Y != 300 {
		t.Fatalf("expected a dead enemy not to be pushed, got (%v, %v)", dead.X, dead.Y)
	}
}

// Neither enemy has priority: the correction is shared.
func TestSeparation_PairIsSplitEvenly(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.enemies = []*Enemy{
		NewEnemy("pod-a", "default", 396, 300),
		NewEnemy("pod-b", "default", 404, 300),
	}

	scene.separate()

	leftMoved := 396 - scene.enemies[0].X
	rightMoved := scene.enemies[1].X - 404
	if math.Abs(leftMoved-rightMoved) > 0.0001 {
		t.Fatalf("expected an even split, left moved %v and right moved %v", leftMoved, rightMoved)
	}
}

// Two Pods can spawn on the same edge in the same tick and land on the same
// point. A zero-length vector must not leave them fused forever.
func TestSeparation_CoincidentEnemiesStillComeApart(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.enemies = []*Enemy{
		NewEnemy("pod-a", "default", 400, 300),
		NewEnemy("pod-b", "default", 400, 300),
	}

	scene.separate()

	got := math.Hypot(
		scene.enemies[0].X-scene.enemies[1].X,
		scene.enemies[0].Y-scene.enemies[1].Y,
	)
	if got < EnemySeparation {
		t.Fatalf("expected coincident enemies at least %v apart, got %v", EnemySeparation, got)
	}
}

// Separation corrects positions; only contact knockback moves enemies under
// their own velocity.
func TestSeparation_DoesNotWriteVelocity(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.enemies = []*Enemy{
		NewEnemy("pod-a", "default", 400, 300),
		NewEnemy("pod-b", "default", 402, 300),
	}

	scene.separate()

	for _, e := range scene.enemies {
		if e.VX != 0 || e.VY != 0 {
			t.Fatalf("expected separation to leave velocity alone, got (%v, %v)", e.VX, e.VY)
		}
	}
}

// Avoidance bends where an enemy is heading, never how fast it gets there, so
// a crowded enemy can never outrun the player.
func TestAvoidance_SteeringDoesNotChangeSpeed(t *testing.T) {
	enemy := NewEnemy("pod-a", "default", 400, 300)

	enemy.Update(400, 100, 1, 0) // player straight up, steer hard right

	moved := math.Hypot(enemy.X-400, enemy.Y-300)
	if math.Abs(moved-EnemySpeed) > 0.0001 {
		t.Fatalf("expected the enemy to move exactly %v, got %v", EnemySpeed, moved)
	}
}

func TestAvoidance_EnemyVeersAwayFromANeighbour(t *testing.T) {
	straight := NewEnemy("pod-a", "default", 400, 300)
	veering := NewEnemy("pod-b", "default", 400, 300)

	straight.Update(400, 100, 0, 0) // player straight up, no neighbours
	veering.Update(400, 100, -1, 0) // same, but steering away from the right

	if veering.X >= straight.X {
		t.Fatalf("expected the steered enemy to veer left of %v, got %v", straight.X, veering.X)
	}
	if veering.Y >= 300 {
		t.Fatal("expected the steered enemy to still make progress toward the player")
	}
}

// Avoidance is a heading change, not a shove.
func TestAvoidance_SteeringDoesNotWriteVelocity(t *testing.T) {
	enemy := NewEnemy("pod-a", "default", 400, 300)

	enemy.Update(400, 100, 1, 0)

	if enemy.VX != 0 || enemy.VY != 0 {
		t.Fatalf("expected steering to leave velocity alone, got (%v, %v)", enemy.VX, enemy.VY)
	}
}

// closestPair returns the distance between the two nearest active enemies.
func closestPair(t *testing.T, enemies []*Enemy) float64 {
	t.Helper()
	closest := math.Inf(1)
	for i, a := range enemies {
		for _, b := range enemies[i+1:] {
			if a.State != EnemyActive || b.State != EnemyActive {
				continue
			}
			if d := math.Hypot(a.X-b.X, a.Y-b.Y); d < closest {
				closest = d
			}
		}
	}
	return closest
}

// The bug this guards: every enemy homes on the same point at the same speed,
// so without Separation the pool converges and never comes apart. A pile of
// overlapping enemies reads as one target, because a projectile stops at the first
// enemy it touches and contact resolves against only one enemy per tick.
func TestSeparation_ConvergingEnemiesNeverOverlap(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 32))
	scene.BeginSession(1, false)
	for i, pos := range [][2]float64{
		{100, 100},
		{700, 100},
		{100, 500},
		{700, 500},
		{400, 50},
		{400, 550},
		{50, 300},
		{750, 300},
		{200, 200},
		{600, 200},
		{200, 400},
		{600, 400},
	} {
		scene.enemies = append(scene.enemies, NewEnemy(
			"pod-"+string(rune('a'+i)), "default", pos[0], pos[1],
		))
	}

	// Long enough that any convergence would have fully collapsed the pool.
	for range 3000 {
		scene.UpdateWithoutPlayer(400, 300)
		if got := closestPair(t, scene.enemies); got < EnemySeparation-separationEpsilon {
			t.Fatalf("expected enemies to stay %v apart, got %v", EnemySeparation, got)
		}
	}

	if len(scene.enemies) != 12 {
		t.Fatalf("expected all 12 enemies to survive, got %d", len(scene.enemies))
	}
}

// Separation must hold on the live play path too, not only during the death
// animation.
func TestSeparation_HoldsWhileHuntingThePlayer(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 32))
	scene.BeginSession(1, false)
	player := NewPlayer(400, 300)
	for i := range 8 {
		scene.enemies = append(scene.enemies, NewEnemy(
			"pod-"+string(rune('a'+i)), "default", float64(100+i*70), 50,
		))
	}

	for range 1000 {
		player.Health = PlayerMaxHealth // keep the ship alive under the swarm
		scene.Update(player)
		if got := closestPair(t, scene.enemies); got < EnemySeparation-separationEpsilon {
			t.Fatalf("expected enemies to stay %v apart, got %v", EnemySeparation, got)
		}
	}
}

// Separation is enforced by relaxation sweeps, capped per tick. A pile handed
// to the scene fully formed takes a few ticks to prise apart rather than one,
// so this pins how quickly the pool recovers.
func TestSeparation_RecoversFromAPileWithinAFewTicks(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 32))
	scene.BeginSession(1, false)
	for i := range 8 {
		scene.enemies = append(scene.enemies, NewEnemy(
			"pod-"+string(rune('a'+i)), "default", 400, float64(300+i*3),
		))
	}

	const recoveryTicks = 10
	for range recoveryTicks {
		scene.UpdateWithoutPlayer(400, 300)
	}

	if got := closestPair(t, scene.enemies); got < EnemySeparation-separationEpsilon {
		t.Fatalf("expected a pile to clear within %d ticks, closest pair still %v", recoveryTicks, got)
	}
}

// Enemies spawn just off-screen and can land on top of each other, so the
// invariant has to hold where the player cannot yet see it.
func TestSeparation_AppliesOffScreen(t *testing.T) {
	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	scene.enemies = []*Enemy{
		NewEnemy("pod-a", "default", -spawnMargin, 300),
		NewEnemy("pod-b", "default", -spawnMargin, 302),
	}

	scene.separate()

	got := math.Hypot(
		scene.enemies[0].X-scene.enemies[1].X,
		scene.enemies[0].Y-scene.enemies[1].Y,
	)
	if got < EnemySeparation {
		t.Fatalf("expected off-screen enemies at least %v apart, got %v", EnemySeparation, got)
	}
}
