package game

import "testing"

func TestExplosionProgressUsesFullAnimationRange(t *testing.T) {
	start := explosionProgress(enemyAnimationTicks)
	end := explosionProgress(0)
	if start <= 0 || start > 0.05 {
		t.Fatalf("expected explosion to start near zero but visible, got %v", start)
	}
	if end != 1 {
		t.Fatalf("expected explosion to finish at 1, got %v", end)
	}
}

func TestDroidLookReflectsEnemyState(t *testing.T) {
	e := NewEnemyOf("pod-a", "ns", ArchetypeTurret, TierHeavy, 100, 100)
	e.DamageFlashTimer = 5
	e.Telegraph = 1.7
	l := droidLookOf(e, 7, 200, 100)

	if !l.damage || l.terminating || l.explode != 0 {
		t.Fatalf("expected damage-only look, got %+v", l)
	}
	if l.charge != 1 {
		t.Fatalf("expected telegraph clamped to 1, got %v", l.charge)
	}
	if l.aimX != 1 || l.aimY != 0 {
		t.Fatalf("expected aim at the player, got %v,%v", l.aimX, l.aimY)
	}

	e.BeginTerminating(0)
	if l := droidLookOf(e, 7, 200, 100); !l.terminating || l.damage {
		t.Fatalf("expected terminating look, got %+v", l)
	}
	e.BeginExploding(enemyAnimationTicks / 2)
	if l := droidLookOf(e, 7, 200, 100); l.explode <= 0 {
		t.Fatalf("expected explode progress, got %+v", l)
	}
}

func TestEnemyHeadingFollowsMovement(t *testing.T) {
	e := NewEnemy("pod-a", "ns", 0, 0)
	for range 60 {
		e.Update(1000, 0, 0, 0)
	}
	if e.HeadingX < 0.5 || e.HeadingY != 0 {
		t.Fatalf("expected heading to approach cruise speed along +X, got %v,%v", e.HeadingX, e.HeadingY)
	}
}

func TestEnemySeedIsStablePerPod(t *testing.T) {
	a := NewEnemy("pod-a", "ns", 0, 0)
	b := NewEnemy("pod-a", "ns", 5, 5)
	c := NewEnemy("pod-b", "ns", 0, 0)
	if a.seed != b.seed || a.seed == c.seed {
		t.Fatalf("seed should depend on pod name only: %v %v %v", a.seed, b.seed, c.seed)
	}
}
