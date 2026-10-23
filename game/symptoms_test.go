package game

import "testing"

func symptomScene() (*Scene, chan GameEvent) {
	events := make(chan GameEvent, 16)
	return NewScene(events, make(chan KillEvent, 8)), events
}

func spawnWith(scene *Scene, events chan GameEvent, name string, s Symptoms) *Enemy {
	events <- GameEvent{Type: EventSpawnEnemy, PodName: name, Namespace: "default", Symptoms: s}
	scene.drainEvents()
	return scene.enemies[len(scene.enemies)-1]
}

func TestPatchPlateBands(t *testing.T) {
	for restarts, want := range map[int]int{0: 0, 1: 1, 4: 1, 5: 2, 6: 2, 40: 2} {
		if got := patchPlates(restarts); got != want {
			t.Errorf("patchPlates(%d) = %d, want %d", restarts, got, want)
		}
	}
}

func TestEyeFlickerIsIrregularAndBounded(t *testing.T) {
	seen := map[float64]bool{}
	for tick := range uint64(300) {
		v := eyeFlicker(tick, 5)
		if v < 0.15 || v > 1 {
			t.Fatalf("flicker %v out of range at tick %d", v, tick)
		}
		seen[v] = true
	}
	if len(seen) < 4 {
		t.Fatalf("expected an irregular flicker, got %d distinct levels", len(seen))
	}
	sparks := 0
	for tick := range uint64(600) {
		if antennaSparks(tick, 5) {
			sparks++
		}
	}
	if sparks == 0 || sparks == 600 {
		t.Fatalf("expected the antenna to spark now and then, got %d/600", sparks)
	}
}

func TestSpawnSymptomsShowOnlyAsPatchesAndNoBlink(t *testing.T) {
	scene, events := symptomScene()
	e := spawnWith(scene, events, "a", Symptoms{Ready: true, Restarts: 6})
	if e.rebootTimer != 0 {
		t.Fatalf("restarts at spawn must not blink, timer=%d", e.rebootTimer)
	}
	if l := symptomLookOf(e, 10); l.plates != 2 || l.sparks || l.eyeLevel != 1 || l.eyeShut {
		t.Fatalf("unexpected look %+v", l)
	}
}

func TestNotReadyFlickersAndSparks(t *testing.T) {
	scene, events := symptomScene()
	e := spawnWith(scene, events, "a", Symptoms{Ready: false})
	dimmed, sparked := false, false
	for tick := range uint64(300) {
		l := symptomLookOf(e, tick)
		dimmed = dimmed || l.eyeLevel < 1
		sparked = sparked || l.sparks
	}
	if !dimmed || !sparked {
		t.Fatalf("expected flicker and sparks, dimmed=%v sparked=%v", dimmed, sparked)
	}
}

func TestRestartIncreaseBlinksForTwentyTicks(t *testing.T) {
	scene, events := symptomScene()
	e := spawnWith(scene, events, "a", Symptoms{Ready: true, Restarts: 1})
	events <- GameEvent{Type: EventSymptomsChanged, PodName: "a", Namespace: "default", Symptoms: Symptoms{Ready: true, Restarts: 2}}
	scene.drainEvents()
	if e.rebootTimer != rebootBlinkTicks {
		t.Fatalf("expected a blink, timer=%d", e.rebootTimer)
	}
	shut := false
	for range rebootBlinkTicks {
		scene.tickSymptoms()
		shut = shut || symptomLookOf(e, 0).eyeShut
	}
	if !shut || e.rebootTimer != 0 || symptomLookOf(e, 0).eyeShut {
		t.Fatalf("blink should shut the eye then end, shut=%v timer=%d", shut, e.rebootTimer)
	}

	// Readiness changes and equal counts never blink.
	e.SetSymptoms(Symptoms{Ready: false, Restarts: 2})
	if e.rebootTimer != 0 {
		t.Fatal("a readiness change must not blink")
	}
}

func TestTerminatingAndExplodingHideSymptoms(t *testing.T) {
	scene, events := symptomScene()
	e := spawnWith(scene, events, "a", Symptoms{Ready: false, Restarts: 9})
	e.rebootTimer = 5

	e.BeginTerminating(0)
	if l := symptomLookOf(e, 3); l != (symptomLook{eyeLevel: 1}) {
		t.Fatalf("terminating must hide Symptoms, got %+v", l)
	}
	e.BeginExploding(enemyAnimationTicks)
	if l := symptomLookOf(e, 3); l != (symptomLook{eyeLevel: 1}) {
		t.Fatalf("exploding must hide Symptoms, got %+v", l)
	}

	e.State = EnemyTerminating
	e.RevertToActive()
	if l := symptomLookOf(e, 3); l.plates != 2 {
		t.Fatalf("a reverted enemy shows its Symptoms again, got %+v", l)
	}
}

func TestSymptomsChangedWhileTerminatingStillRecorded(t *testing.T) {
	scene, events := symptomScene()
	e := spawnWith(scene, events, "a", Symptoms{Ready: true})
	e.BeginTerminating(0)
	events <- GameEvent{Type: EventSymptomsChanged, PodName: "a", Namespace: "default", Symptoms: Symptoms{Ready: false, Restarts: 5}}
	scene.drainEvents()
	e.RevertToActive()
	if l := symptomLookOf(e, 1); l.plates != 2 {
		t.Fatalf("reverted enemy must show current Symptoms, got %+v", l)
	}
}

func TestPendingRespawnTracksSymptomsWithoutBlink(t *testing.T) {
	scene, events := symptomScene()
	scene.BeginSession(1, false)
	scene.recordKill(KillEvent{PodName: "a", Namespace: "default"}, ArchetypeProbe, TierStandard, Symptoms{Ready: true})

	events <- GameEvent{Type: EventSymptomsChanged, PodName: "a", Namespace: "default", Symptoms: Symptoms{Ready: false, Restarts: 5}}
	scene.drainEvents()

	for range respawnDelayTicks {
		scene.updateRespawns()
	}
	if len(scene.enemies) != 1 {
		t.Fatalf("expected the respawn, got %d enemies", len(scene.enemies))
	}
	e := scene.enemies[0]
	if e.Symptoms != (Symptoms{Ready: false, Restarts: 5}) || e.rebootTimer != 0 {
		t.Fatalf("respawn should carry current Symptoms without a blink: %+v timer=%d", e.Symptoms, e.rebootTimer)
	}
}

func TestSymptomsNeverAffectBehaviourOrStats(t *testing.T) {
	healthy := NewEnemyOf("a", "default", ArchetypeProbe, TierStandard, 100, 100)
	sick := NewEnemyOf("a", "default", ArchetypeProbe, TierStandard, 100, 100)
	sick.SetSymptoms(Symptoms{Ready: false, Restarts: 9})
	for range 50 {
		healthy.Update(400, 300, 0, 0)
		sick.Update(400, 300, 0, 0)
	}
	if healthy.X != sick.X || healthy.Y != sick.Y || healthy.Hits != sick.Hits ||
		healthy.CruiseSpeed() != sick.CruiseSpeed() || healthy.ContactDamage() != sick.ContactDamage() {
		t.Fatal("Symptoms changed behaviour or stats")
	}
}
