package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	playerSize              = 20.0
	playerTurnSpeed         = 0.048
	playerAccel             = 0.1666666667
	playerBrakeAccel        = 0.20
	playerMaxSpeed          = 4.0
	projectileSpeed         = 8.0
	playerAimDistance       = 120.0
	shootCooldown           = 15 // ticks between shots
	barrelRollDurationTicks = 45
	barrelRollCooldownTicks = 180

	// PlayerMaxHealth is the player's starting damage pool.
	PlayerMaxHealth = 10
	// invulnerableTicks is the window after contact during which the player
	// cannot take further contact damage.
	invulnerableTicks = 60
	// playerKnockbackSpeed is the impulse the player receives from contact.
	playerKnockbackSpeed = 4.0
)

// Player is the player-controlled spaceship.
type Player struct {
	X, Y    float64
	VX, VY  float64
	Heading float64
	Health  int
	// InvulnerableTimer counts down the window after contact in which further
	// contact does no damage.
	InvulnerableTimer int
	// RollTicks counts down an active Barrel roll.
	RollTicks int
	// RollCooldown counts down until the next Barrel roll can start.
	RollCooldown    int
	thrustHeld      bool
	MainWeapon      Weapon
	SecondaryWeapon Weapon
}

type playerControls struct {
	turnLeft       bool
	turnRight      bool
	thrust         bool
	brake          bool
	roll           bool
	shootMain      bool
	shootSecondary bool
}

// NewPlayer creates the player at position (x, y).
func NewPlayer(x, y float64) *Player {
	return &Player{
		X:               x,
		Y:               y,
		Heading:         -math.Pi / 2, // point upward
		Health:          PlayerMaxHealth,
		MainWeapon:      NewBasicWeapon(),
		SecondaryWeapon: NewRocketWeapon(),
	}
}

// Reset returns the player to its starting state at (x, y). Used on retry:
// the ship is restored, but the enemy pool is not, because the Pods behind it
// never died.
func (p *Player) Reset(x, y float64) {
	p.X, p.Y = x, y
	p.VX, p.VY = 0, 0
	p.Heading = -math.Pi / 2
	p.Health = PlayerMaxHealth
	p.InvulnerableTimer = 0
	p.RollTicks = 0
	p.RollCooldown = 0
	p.thrustHeld = false
	p.MainWeapon.Reset()
	p.SecondaryWeapon.Reset()
}

// Update reads keyboard input and moves the player.
// Returns any projectiles fired this tick.
func (p *Player) Update() []*Projectile {
	controls := playerControls{
		turnLeft:       ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA),
		turnRight:      ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD),
		thrust:         ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW),
		brake:          ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS),
		roll:           inpututil.IsKeyJustPressed(ebiten.KeyZ),
		shootMain:      inpututil.IsKeyJustPressed(ebiten.KeySpace),
		shootSecondary: inpututil.IsKeyJustPressed(ebiten.KeyX),
	}
	return p.updateWithControls(controls)
}

func (p *Player) updateWithControls(controls playerControls) []*Projectile {
	if p.InvulnerableTimer > 0 {
		p.InvulnerableTimer--
	}
	if p.RollTicks > 0 {
		p.RollTicks--
	}
	if p.RollCooldown > 0 {
		p.RollCooldown--
	}
	if controls.roll && p.RollCooldown == 0 {
		p.RollTicks = barrelRollDurationTicks
		p.RollCooldown = barrelRollCooldownTicks
	}
	p.thrustHeld = controls.thrust

	if controls.turnLeft {
		p.Heading -= playerTurnSpeed
	}
	if controls.turnRight {
		p.Heading += playerTurnSpeed
	}

	if controls.thrust {
		p.VX += math.Cos(p.Heading) * playerAccel
		p.VY += math.Sin(p.Heading) * playerAccel
	}

	if controls.brake {
		speed := math.Hypot(p.VX, p.VY)
		if speed <= playerBrakeAccel {
			p.VX = 0
			p.VY = 0
		} else {
			scale := (speed - playerBrakeAccel) / speed
			p.VX *= scale
			p.VY *= scale
		}
	}

	speed := math.Hypot(p.VX, p.VY)
	if speed > playerMaxSpeed {
		scale := playerMaxSpeed / speed
		p.VX *= scale
		p.VY *= scale
	}

	p.X += p.VX
	p.Y += p.VY
	p.X, p.Y = wrapPosition(p.X, p.Y)

	p.MainWeapon.Update()
	p.SecondaryWeapon.Update()

	var fired []*Projectile
	if controls.shootMain && p.MainWeapon.CanFire() {
		fired = append(fired, p.MainWeapon.Fire(p.X, p.Y, p.Heading)...)
	}
	if controls.shootSecondary && p.SecondaryWeapon.CanFire() {
		fired = append(fired, p.SecondaryWeapon.Fire(p.X, p.Y, p.Heading)...)
	}

	return fired
}

// TakeContactDamage costs the player damage health and pushes it along the unit
// vector (nx, ny), unless it is still invulnerable from an earlier contact.
// Returns true if damage was actually taken.
func (p *Player) TakeContactDamage(damage int, nx, ny float64) bool {
	return p.takeDamage(damage, nx, ny, 1)
}

// TakeBoltDamage is TakeContactDamage for a Bolt impact: it shares the
// Invulnerability window but knocks the player back only boltKnockbackScale
// as hard.
func (p *Player) TakeBoltDamage(damage int, nx, ny float64) bool {
	return p.takeDamage(damage, nx, ny, boltKnockbackScale)
}

func (p *Player) takeDamage(damage int, nx, ny, knockbackScale float64) bool {
	if p.InvulnerableTimer > 0 {
		return false
	}
	p.Health -= damage
	if p.Health < 0 {
		p.Health = 0
	}
	p.InvulnerableTimer = invulnerableTicks
	p.VX += nx * playerKnockbackSpeed * knockbackScale
	p.VY += ny * playerKnockbackSpeed * knockbackScale
	return true
}

// IsInvulnerable reports whether contact damage is currently suppressed.
func (p *Player) IsInvulnerable() bool {
	return p.InvulnerableTimer > 0
}

// IsRolling reports whether the Barrel roll is currently active.
func (p *Player) IsRolling() bool {
	return p.RollTicks > 0
}

// AimPoint returns the world position along the player's current heading.
func (p *Player) AimPoint() (float64, float64) {
	return p.X + math.Cos(p.Heading)*playerAimDistance,
		p.Y + math.Sin(p.Heading)*playerAimDistance
}

// Bounds returns the player's collision rectangle.
func (p *Player) Bounds() Rect {
	half := playerSize / 2
	return Rect{X: p.X - half, Y: p.Y - half, W: playerSize, H: playerSize}
}

// wrapPosition wraps a position around the screen edges, so the player can
// never escape into empty space off-screen.
func wrapPosition(x, y float64) (float64, float64) {
	if x < 0 {
		x += screenWidth
	} else if x >= screenWidth {
		x -= screenWidth
	}
	if y < 0 {
		y += screenHeight
	} else if y >= screenHeight {
		y -= screenHeight
	}
	return x, y
}
