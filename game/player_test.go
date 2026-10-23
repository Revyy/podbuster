package game

import (
	"math"
	"testing"
)

func TestPlayerStep_TurningChangesHeading(t *testing.T) {
	player := NewPlayer(100, 100)
	initialHeading := player.Heading

	player.updateWithControls(playerControls{turnRight: true})

	if player.Heading <= initialHeading {
		t.Fatalf("expected heading to increase, got %v from %v", player.Heading, initialHeading)
	}
}

func TestPlayerStep_AccelerationClampsToMaxSpeed(t *testing.T) {
	player := NewPlayer(100, 100)

	for range 200 {
		player.updateWithControls(playerControls{thrust: true})
	}

	speed := math.Hypot(player.VX, player.VY)
	if speed > playerMaxSpeed+1e-9 {
		t.Fatalf("expected speed <= %v, got %v", playerMaxSpeed, speed)
	}
	if speed <= 0 {
		t.Fatal("expected player to be moving after thrust")
	}
}

func TestPlayerStep_BrakeSlowsPlayerToStop(t *testing.T) {
	player := NewPlayer(100, 100)

	for range 25 {
		player.updateWithControls(playerControls{thrust: true})
	}

	for range 40 {
		player.updateWithControls(playerControls{brake: true})
	}

	speed := math.Hypot(player.VX, player.VY)
	if speed > 1e-9 {
		t.Fatalf("expected speed to be near zero, got %v", speed)
	}
}

func TestPlayerStep_CanSpinAndAccelerateOppositeDirection(t *testing.T) {
	player := NewPlayer(100, 100)

	for range 25 {
		player.updateWithControls(playerControls{thrust: true})
	}
	initialVerticalVelocity := player.VY
	if initialVerticalVelocity >= 0 {
		t.Fatalf("expected initial velocity to be upward (negative Y), got %v", initialVerticalVelocity)
	}

	turnSteps := int(math.Ceil(math.Pi / playerTurnSpeed))
	for range turnSteps {
		player.updateWithControls(playerControls{turnRight: true})
	}
	for range 40 {
		player.updateWithControls(playerControls{thrust: true})
	}

	if player.VY <= 0 {
		t.Fatalf("expected vertical velocity to flip positive after turning and thrusting, got %v", player.VY)
	}
}

func TestPlayerStep_ThrustAccelerationIsTunedSlower(t *testing.T) {
	player := NewPlayer(100, 100)

	player.updateWithControls(playerControls{thrust: true})

	speed := math.Hypot(player.VX, player.VY)
	if math.Abs(speed-0.1666666667) > 0.0001 {
		t.Fatalf("expected tuned thrust speed delta near 0.1667, got %v", speed)
	}
}

func TestPlayerStep_BrakeDecelerationIsTunedSlower(t *testing.T) {
	player := NewPlayer(100, 100)
	player.VX = 3
	player.VY = 0

	player.updateWithControls(playerControls{brake: true})

	speed := math.Hypot(player.VX, player.VY)
	if math.Abs(speed-2.8) > 0.0001 {
		t.Fatalf("expected tuned brake speed near 2.8 after one tick, got %v", speed)
	}
}

func TestPlayerStep_RollStartsOnlyWhenCooldownIsReady(t *testing.T) {
	player := NewPlayer(100, 100)

	player.updateWithControls(playerControls{roll: true})
	if player.RollTicks != barrelRollDurationTicks {
		t.Fatalf("expected roll to start at %d ticks, got %d", barrelRollDurationTicks, player.RollTicks)
	}
	if player.RollCooldown != barrelRollCooldownTicks {
		t.Fatalf("expected cooldown to start at %d ticks, got %d", barrelRollCooldownTicks, player.RollCooldown)
	}

	player.updateWithControls(playerControls{})
	player.updateWithControls(playerControls{roll: true})
	if player.RollTicks != barrelRollDurationTicks-2 {
		t.Fatalf("expected cooldown to block a new roll, got %d ticks", player.RollTicks)
	}
	if player.RollCooldown != barrelRollCooldownTicks-2 {
		t.Fatalf("expected cooldown to keep counting down, got %d", player.RollCooldown)
	}
}

func TestPlayerStep_RollAndCooldownTickDownIndependentlyOfInvulnerability(t *testing.T) {
	player := NewPlayer(100, 100)
	player.InvulnerableTimer = 3

	player.updateWithControls(playerControls{roll: true})
	for range barrelRollDurationTicks {
		player.updateWithControls(playerControls{})
	}

	if player.RollTicks != 0 {
		t.Fatalf("expected roll to finish after %d ticks, got %d", barrelRollDurationTicks, player.RollTicks)
	}
	if player.RollCooldown != barrelRollCooldownTicks-barrelRollDurationTicks {
		t.Fatalf("expected cooldown to keep running after the roll, got %d", player.RollCooldown)
	}
	if player.InvulnerableTimer != 0 {
		t.Fatalf("expected invulnerability to keep counting down under the roll, got %d", player.InvulnerableTimer)
	}

	for range barrelRollCooldownTicks - barrelRollDurationTicks {
		player.updateWithControls(playerControls{})
	}
	if player.RollCooldown != 0 {
		t.Fatalf("expected cooldown to finish after %d ticks, got %d", barrelRollCooldownTicks, player.RollCooldown)
	}
}

func TestPlayerStep_RollKeepsSteeringThrustAndFiring(t *testing.T) {
	rolling := NewPlayer(100, 100)
	plain := NewPlayer(100, 100)

	rolling.updateWithControls(playerControls{roll: true})
	controls := playerControls{
		turnRight:      true,
		thrust:         true,
		shootMain:      true,
		shootSecondary: true,
	}
	rollingFired := rolling.updateWithControls(controls)
	plainFired := plain.updateWithControls(controls)

	if len(rollingFired) != len(plainFired) {
		t.Fatalf("expected roll not to block firing, got %d projectiles want %d", len(rollingFired), len(plainFired))
	}
	if rolling.Heading != plain.Heading {
		t.Fatalf("expected roll not to block steering, got heading %v want %v", rolling.Heading, plain.Heading)
	}
	if rolling.VX != plain.VX || rolling.VY != plain.VY {
		t.Fatalf("expected roll not to change thrust, got velocity (%v, %v) want (%v, %v)", rolling.VX, rolling.VY, plain.VX, plain.VY)
	}
	if rolling.X != plain.X || rolling.Y != plain.Y {
		t.Fatalf("expected roll not to add movement, got position (%v, %v) want (%v, %v)", rolling.X, rolling.Y, plain.X, plain.Y)
	}
}

func TestPlayerStep_ThrustHeldFollowsControls(t *testing.T) {
	player := NewPlayer(100, 100)

	player.updateWithControls(playerControls{thrust: true})
	if !player.thrustHeld {
		t.Fatal("expected thrust flag while thrust is held")
	}

	player.updateWithControls(playerControls{})
	if player.thrustHeld {
		t.Fatal("expected thrust flag to clear when thrust is released")
	}
}

func TestPlayer_AimPointFollowsHeading(t *testing.T) {
	tests := []struct {
		name            string
		heading         float64
		useDefault      bool
		expectedOffsetX float64
		expectedOffsetY float64
	}{
		{name: "default heading", useDefault: true, expectedOffsetY: -playerAimDistance},
		{name: "right", heading: 0, expectedOffsetX: playerAimDistance},
		{name: "down", heading: math.Pi / 2, expectedOffsetY: playerAimDistance},
		{name: "left", heading: math.Pi, expectedOffsetX: -playerAimDistance},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			player := NewPlayer(400, 300)
			if !test.useDefault {
				player.Heading = test.heading
			}

			x, y := player.AimPoint()
			if math.Abs((x-player.X)-test.expectedOffsetX) > 1e-9 ||
				math.Abs((y-player.Y)-test.expectedOffsetY) > 1e-9 {
				t.Fatalf("expected aim offset (%v, %v), got (%v, %v)",
					test.expectedOffsetX, test.expectedOffsetY, x-player.X, y-player.Y)
			}
		})
	}
}

func TestPlayer_AimPointStaysAtAimDistance(t *testing.T) {
	player := NewPlayer(400, 300)

	for _, heading := range []float64{-2.7, -1.1, 0.2, 1.8, 3.0} {
		player.Heading = heading
		x, y := player.AimPoint()
		distance := math.Hypot(x-player.X, y-player.Y)
		if math.Abs(distance-playerAimDistance) > 1e-9 {
			t.Fatalf("heading %v: expected distance %v, got %v", heading, playerAimDistance, distance)
		}
	}
}

func TestPlayer_AimPointDoesNotWrapAtScreenEdge(t *testing.T) {
	player := NewPlayer(screenWidth-10, screenHeight/2)
	player.Heading = 0

	x, _ := player.AimPoint()

	if x <= screenWidth {
		t.Fatalf("expected aim point beyond right edge, got x=%v", x)
	}
}

func TestPlayer_ResetMakesWeaponsReadyToFire(t *testing.T) {
	player := NewPlayer(100, 100)

	// Exhaust both weapons by firing them.
	player.MainWeapon.Fire(100, 100, 0)
	player.SecondaryWeapon.Fire(100, 100, 0)

	// Verify weapons are in cooldown.
	if player.MainWeapon.CanFire() {
		t.Fatal("expected MainWeapon to be in cooldown after firing")
	}
	if player.SecondaryWeapon.CanFire() {
		t.Fatal("expected SecondaryWeapon to be in cooldown after firing")
	}

	// Reset the player.
	player.Reset(200, 200)

	// Verify both weapons are ready to fire.
	if !player.MainWeapon.CanFire() {
		t.Fatal("expected MainWeapon to be ready after Reset")
	}
	if !player.SecondaryWeapon.CanFire() {
		t.Fatal("expected SecondaryWeapon to be ready after Reset")
	}
}

func TestPlayer_ResetPreservesWeaponEquip(t *testing.T) {
	player := NewPlayer(100, 100)

	mainBefore := player.MainWeapon
	secondaryBefore := player.SecondaryWeapon

	player.Reset(200, 200)

	if player.MainWeapon != mainBefore {
		t.Fatal("expected MainWeapon to remain the same object after Reset")
	}
	if player.SecondaryWeapon != secondaryBefore {
		t.Fatal("expected SecondaryWeapon to remain the same object after Reset")
	}
}

func TestPlayer_ResetClearsRollState(t *testing.T) {
	player := NewPlayer(100, 100)
	player.RollTicks = 12
	player.RollCooldown = 99
	player.thrustHeld = true

	player.Reset(200, 200)

	if player.RollTicks != 0 || player.RollCooldown != 0 {
		t.Fatalf("expected reset to clear roll state, got ticks=%d cooldown=%d", player.RollTicks, player.RollCooldown)
	}
	if player.thrustHeld {
		t.Fatal("expected reset to clear the thrust flag")
	}
}
