package game

import "testing"

func newTestGame(t *testing.T) (*Game, chan SessionEvent) {
	t.Helper()
	sessionEvents := make(chan SessionEvent, 8)
	g := NewGame(
		Config{Namespace: "default"},
		make(chan GameEvent, 8),
		make(chan KillEvent, 8),
		sessionEvents,
	)
	return g, sessionEvents
}

func TestGame_StartsOnTitleScreen(t *testing.T) {
	g, sessionEvents := newTestGame(t)

	if g.State() != StateTitle {
		t.Fatalf("expected the game to start on the title screen, got %v", g.State())
	}
	// No cluster data may be read before the player starts a game.
	if len(sessionEvents) != 0 {
		t.Fatalf("expected no session event on startup, got %d", len(sessionEvents))
	}
}

func TestGame_StartSessionAnnouncesNamespaceAndMode(t *testing.T) {
	g, sessionEvents := newTestGame(t)

	g.startSession(true)

	if g.State() != StatePlaying {
		t.Fatalf("expected StatePlaying, got %v", g.State())
	}
	ev := <-sessionEvents
	if ev.Type != EventSessionStart {
		t.Fatalf("expected a session start event, got %v", ev.Type)
	}
	if ev.Namespace != "default" {
		t.Fatalf("expected namespace %q, got %q", "default", ev.Namespace)
	}
	if !ev.ChaosMode {
		t.Fatal("expected chaos mode to travel with the session start")
	}
}

func TestGame_EndSessionClearsPoolAndReturnsToTitle(t *testing.T) {
	g, sessionEvents := newTestGame(t)
	g.startSession(false)
	<-sessionEvents // start

	g.scene.enemies = append(g.scene.enemies, NewEnemy("pod-a", "default", 100, 100))
	g.endSession()

	if g.State() != StateTitle {
		t.Fatalf("expected StateTitle after ending the session, got %v", g.State())
	}
	if ev := <-sessionEvents; ev.Type != EventSessionEnd {
		t.Fatalf("expected a session end event, got %v", ev.Type)
	}
	if len(g.scene.Enemies()) != 0 {
		t.Fatalf("expected the enemy pool cleared, got %d enemies", len(g.scene.Enemies()))
	}
}

func TestGame_ZeroHealthEntersDyingThenGameOver(t *testing.T) {
	g, sessionEvents := newTestGame(t)
	g.startSession(false)
	<-sessionEvents

	g.player.Health = 0
	g.updatePlaying()

	if g.State() != StateDying {
		t.Fatalf("expected StateDying at zero health, got %v", g.State())
	}

	for range deathTicks {
		g.updateDying()
	}

	if g.State() != StateGameOver {
		t.Fatalf("expected StateGameOver after the death beat, got %v", g.State())
	}
}

// Retry restores the ship but leaves the enemy pool alone: the Pods never died.
func TestGame_RetryResetsPlayerButKeepsEnemyPool(t *testing.T) {
	g, sessionEvents := newTestGame(t)
	g.startSession(false)
	<-sessionEvents

	g.scene.enemies = append(g.scene.enemies, NewEnemy("pod-a", "default", 100, 100))
	g.player.Health = 0
	g.state = StateGameOver

	// Mirror what the R key does.
	g.player.Reset(screenWidth/2, screenHeight/2)
	g.state = StatePlaying

	if g.player.Health != PlayerMaxHealth {
		t.Fatalf("expected full health on retry, got %d", g.player.Health)
	}
	if len(g.scene.Enemies()) != 1 {
		t.Fatalf("expected the enemy pool preserved on retry, got %d", len(g.scene.Enemies()))
	}
	if len(sessionEvents) != 0 {
		t.Fatalf("expected retry not to restart the session, got %d events", len(sessionEvents))
	}
}

// A watch that failed must end the session and send the player back to choose.
func TestGame_WatchErrorReturnsToModeSelect(t *testing.T) {
	sessionEvents := make(chan SessionEvent, 8)
	gameEvents := make(chan GameEvent, 8)
	g := NewGame(Config{Namespace: "default"}, gameEvents, make(chan KillEvent, 8), sessionEvents)

	g.startSession(true)
	<-sessionEvents // start

	gameEvents <- GameEvent{Type: EventError, Namespace: "default", Message: "connection refused"}
	g.updatePlaying()

	if g.State() != StateModeSelect {
		t.Fatalf("expected StateModeSelect after a watch failure, got %v", g.State())
	}
	if ev := <-sessionEvents; ev.Type != EventSessionEnd {
		t.Fatalf("expected the failed session to be ended, got %v", ev.Type)
	}
	if g.scene.LastError() != "connection refused" {
		t.Fatalf("expected the error preserved for display, got %q", g.scene.LastError())
	}
}

// Cooldown does not advance during StateDying because Player.Update is not called.
func TestGame_WeaponCooldownDoesNotAdvanceDuringDying(t *testing.T) {
	g, sessionEvents := newTestGame(t)
	g.startSession(false)
	<-sessionEvents

	// Fire a weapon to enter cooldown.
	g.player.MainWeapon.Fire(g.player.X, g.player.Y, g.player.Heading)
	mainStatus := g.player.MainWeapon.(WeaponStatus)
	cooldownBefore := mainStatus.CooldownRemaining()
	if cooldownBefore == 0 {
		t.Fatal("expected weapon to be in cooldown after firing")
	}

	// Kill the player and enter the death state.
	g.player.Health = 0
	g.updatePlaying() // This calls Player.Update() and will tick the cooldown once
	if g.State() != StateDying {
		t.Fatalf("expected StateDying, got %v", g.State())
	}

	mainStatus = g.player.MainWeapon.(WeaponStatus)
	cooldownAfterUpdate := mainStatus.CooldownRemaining()

	// Now let the death animation run and verify cooldown is frozen during dying.
	for range 10 {
		g.updateDying() // This does NOT call Player.Update(), so cooldown should not change
		mainStatus := g.player.MainWeapon.(WeaponStatus)
		if mainStatus.CooldownRemaining() != cooldownAfterUpdate {
			t.Fatalf("expected cooldown to remain %d during dying, but it changed to %d",
				cooldownAfterUpdate, mainStatus.CooldownRemaining())
		}
	}
}

func TestGame_DyingClearsAnActiveRoll(t *testing.T) {
	g, sessionEvents := newTestGame(t)
	g.startSession(false)
	<-sessionEvents

	g.player.RollTicks = barrelRollDurationTicks
	g.player.RollCooldown = barrelRollCooldownTicks
	g.player.Health = 0

	g.updatePlaying()

	if g.State() != StateDying {
		t.Fatalf("expected StateDying, got %v", g.State())
	}
	if g.player.RollTicks != 0 {
		t.Fatalf("expected dying to stop the roll, got %d ticks", g.player.RollTicks)
	}
	if g.player.RollCooldown != barrelRollCooldownTicks-1 {
		t.Fatalf("expected dying not to reset cooldown, got %d", g.player.RollCooldown)
	}
}

func TestGame_RetryStartsWithRollReady(t *testing.T) {
	g, sessionEvents := newTestGame(t)
	g.startSession(false)
	<-sessionEvents

	g.player.RollTicks = 7
	g.player.RollCooldown = 99
	g.state = StateGameOver

	g.player.Reset(screenWidth/2, screenHeight/2)
	g.state = StatePlaying

	if g.player.RollTicks != 0 || g.player.RollCooldown != 0 {
		t.Fatalf("expected retry to restore roll readiness, got ticks=%d cooldown=%d", g.player.RollTicks, g.player.RollCooldown)
	}
}

func TestGame_GameOverKeepsBoltsMovingUntilTheyExpire(t *testing.T) {
	g, sessionEvents := newTestGame(t)
	g.startSession(false)
	<-sessionEvents
	b := NewBolt(100, 100, 0, ArchetypeTurret)
	g.scene.bolts = []*Bolt{b}
	g.player.X, g.player.Y = 102, 100
	g.state = StateGameOver

	g.updateGameOver()
	if b.X != 100+boltSpeed {
		t.Fatalf("bolt did not advance in Game Over: x=%v", b.X)
	}
	if !b.Active || g.player.Health != PlayerMaxHealth {
		t.Fatal("bolt must not hit during Game Over")
	}
	for range 1000 {
		g.updateGameOver()
	}
	if len(g.scene.Bolts()) != 0 {
		t.Fatalf("expected the bolt to expire, %d left", len(g.scene.Bolts()))
	}
}

func TestGame_GameOverAgesProjectileImpacts(t *testing.T) {
	g, sessionEvents := newTestGame(t)
	g.startSession(false)
	<-sessionEvents
	g.scene.projectileImpacts = []ProjectileImpact{{X: 200, Y: 200}}
	g.state = StateGameOver

	g.updateGameOver()
	if got := g.scene.ProjectileImpacts()[0].Age; got != 1 {
		t.Fatalf("expected projectile impact age 1 in Game Over, got %d", got)
	}

	for range projectileImpactTicks - 1 {
		g.updateGameOver()
	}
	if got := len(g.scene.ProjectileImpacts()); got != 0 {
		t.Fatalf("expected projectile impacts to expire in Game Over, got %d", got)
	}
}
