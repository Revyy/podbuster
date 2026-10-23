package game

import (
	"math"
	"testing"
)

func TestCooldownFillRatio_ReportsSecondaryRechargeProgress(t *testing.T) {
	tests := []struct {
		name      string
		remaining int
		total     int
		want      float64
	}{
		{name: "empty after firing", remaining: rocketCooldown, total: rocketCooldown, want: 0},
		{name: "half recharged", remaining: rocketCooldown / 2, total: rocketCooldown, want: 0.5},
		{name: "ready", remaining: 0, total: rocketCooldown, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cooldownFillRatio(tt.remaining, tt.total); got != tt.want {
				t.Fatalf("cooldownFillRatio(%d, %d) = %v, want %v", tt.remaining, tt.total, got, tt.want)
			}
		})
	}
}

func TestRocketWeapon_StatusReportsNameAndCooldown(t *testing.T) {
	weapon := NewRocketWeapon()
	var status WeaponStatus = weapon

	if got := status.Name(); got != "Rockets" {
		t.Fatalf("weapon name = %q, want %q", got, "Rockets")
	}
	if got := status.CooldownTotal(); got != rocketCooldown {
		t.Fatalf("cooldown total = %d, want %d", got, rocketCooldown)
	}
	if got := status.CooldownRemaining(); got != 0 {
		t.Fatalf("cooldown remaining before firing = %d, want 0", got)
	}

	weapon.Fire(100, 100, -math.Pi/2)
	if got := status.CooldownRemaining(); got != rocketCooldown {
		t.Fatalf("cooldown remaining after firing = %d, want %d", got, rocketCooldown)
	}
}

func TestRocketWeapon_FiresThreeShotsInSpreadPattern(t *testing.T) {
	player := NewPlayer(400, 300)
	player.Heading = -math.Pi / 2
	player.SecondaryWeapon = NewRocketWeapon()

	fired := player.updateWithControls(playerControls{shootSecondary: true})
	if len(fired) != 3 {
		t.Fatalf("expected 3 rockets, got %d", len(fired))
	}

	wantOffsets := []float64{0, -rocketSpread, rocketSpread}
	gotOrigins := map[float64]bool{}
	for i, shot := range fired {
		if shot.HitCost != 2 {
			t.Fatalf("expected rocket hit cost 2, got %d", shot.HitCost)
		}
		if shot.Kind != projectileKindRocket {
			t.Fatalf("expected shot %d to be kind rocket, got %v", i, shot.Kind)
		}
		if math.Abs(math.Hypot(shot.VX, shot.VY)-rocketSpeed) > 0.01 {
			t.Fatalf("expected rocket speed %v, got %v", rocketSpeed, math.Hypot(shot.VX, shot.VY))
		}

		angle := math.Atan2(shot.VY, shot.VX)
		wantAngle := player.Heading + wantOffsets[i]
		delta := math.Abs(math.Atan2(math.Sin(angle-wantAngle), math.Cos(angle-wantAngle)))
		if delta > 0.01 {
			t.Fatalf("expected shot %d at heading offset %.4f rad, got offset %.4f rad", i, wantOffsets[i], angle-player.Heading)
		}

		// The two side shots must originate away from the ship's centerline
		// so the burst reads as three parallel tracks, not one point source.
		gotOrigins[math.Round(shot.X*100)/100] = true
	}
	if len(gotOrigins) != 3 {
		t.Fatalf("expected 3 distinct rocket origins (center + left/right offset), got %d", len(gotOrigins))
	}

	if player.SecondaryWeapon.CanFire() {
		t.Fatal("expected secondary weapon to enter cooldown after firing")
	}
}

func TestRocketWeapon_FireUsesTwoHitCostAndTenBySizeAndClampsOverkill(t *testing.T) {
	weapon := NewRocketWeapon()
	rockets := weapon.Fire(400, 300, -math.Pi/2)
	if len(rockets) != 3 {
		t.Fatalf("expected 3 rockets fired, got %d", len(rockets))
	}

	scene := NewScene(make(chan GameEvent), make(chan KillEvent, 8))
	enemy := NewEnemy("pod-a", "default", rockets[0].X, rockets[0].Y)
	scene.enemies = []*Enemy{enemy}
	scene.projectiles = []*Projectile{rockets[0]}

	if rockets[0].Size != 10.0 {
		t.Fatalf("expected a rocket to be a 10x10 projectile, got size %v", rockets[0].Size)
	}
	bounds := rockets[0].Bounds()
	if bounds.W != 10.0 || bounds.H != 10.0 {
		t.Fatalf("expected rocket collision bounds 10x10, got %vx%v", bounds.W, bounds.H)
	}

	scene.resolveCollisions()

	if enemy.Hits != EnemyHits-2 {
		t.Fatalf("expected a rocket to cost exactly 2 hits, got %d", enemy.Hits)
	}
	if rockets[0].Active {
		t.Fatal("expected the rocket to deactivate after resolving a hit")
	}

	overkill := NewEnemy("pod-b", "default", 300, 200)
	overkill.Hit(10)
	if overkill.Hits != 0 {
		t.Fatalf("expected overkill to clamp at zero, got %d", overkill.Hits)
	}
}

func TestRocketWeapon_CooldownBlocksRefireUntilExpiry(t *testing.T) {
	weapon := NewRocketWeapon()
	weapon.Fire(100, 100, -math.Pi/2)

	for tick := range rocketCooldown {
		if weapon.CanFire() {
			t.Fatalf("expected the volley to still be cooling down at tick %d", tick)
		}
		if len(weapon.Fire(100, 100, -math.Pi/2)) != 0 {
			t.Fatalf("expected firing before cooldown expiry (tick %d) to produce nothing", tick)
		}
		weapon.Update()
	}

	if !weapon.CanFire() {
		t.Fatal("expected the rocket volley to be ready again after the cooldown expires")
	}
	if len(weapon.Fire(100, 100, -math.Pi/2)) != 3 {
		t.Fatal("expected a fresh volley once the cooldown has fully elapsed")
	}
}

func TestPlayer_SecondaryFiringDoesNotAffectMainWeaponCooldown(t *testing.T) {
	player := NewPlayer(400, 300)
	player.SecondaryWeapon = NewRocketWeapon()

	mainFired := player.updateWithControls(playerControls{shootMain: true, shootSecondary: true})
	if len(mainFired) != 4 { // 1 basic shot + 3 rockets
		t.Fatalf("expected both weapons to fire on the same tick, got %d projectiles", len(mainFired))
	}
	if player.MainWeapon.CanFire() {
		t.Fatal("expected the main weapon to enter its own cooldown after firing")
	}

	// Advance past the main weapon's 15-tick cooldown but well short of the
	// secondary's 180-tick cooldown, and confirm they recover independently.
	for range shootCooldown {
		player.updateWithControls(playerControls{})
	}
	if !player.MainWeapon.CanFire() {
		t.Fatal("expected the main weapon to be ready again after its own cooldown")
	}
	if player.SecondaryWeapon.CanFire() {
		t.Fatal("expected the secondary weapon to still be cooling down independently of the main weapon")
	}
}
