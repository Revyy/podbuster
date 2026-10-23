package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// State is the game's top-level mode of operation.
type State int

const (
	// StateTitle is the start screen. No cluster data is read here.
	StateTitle State = iota
	// StateModeSelect is where the player chooses practice or chaos mode.
	StateModeSelect
	// StatePlaying is an active session with a live watch.
	StatePlaying
	// StateDying is the brief beat between losing the last health and the
	// game over screen.
	StateDying
	// StateGameOver is the overlay offering retry or a return to the title.
	StateGameOver
	// StateFieldGuide is the paged reference opened from mode select. It
	// starts no session and reads no cluster data.
	StateFieldGuide
)

// modeEntry is an entry on the mode-select screen.
type modeEntry int

const (
	entryPractice modeEntry = iota
	entryChaos
	// entryGuide opens the Field guide. It is not a mode.
	entryGuide
	modeEntryCount
)

// deathTicks is how long the death animation runs before the game over overlay.
const deathTicks = 60

// Config carries everything main.go needs to hand the game. Chaos mode is
// deliberately absent: it is chosen in-game, never configured from outside.
type Config struct {
	// Namespace is the namespace a session watches. Displayed read-only on
	// the mode-select screen.
	Namespace string
}

// Game implements ebiten.Game and owns the top-level game loop.
type Game struct {
	config        Config
	state         State
	player        *Player
	scene         *Scene
	stars         []star
	sessionEvents chan<- SessionEvent
	// modeCursor is the entry highlighted on the mode-select screen.
	modeCursor modeEntry
	guide      *fieldGuide
	// confirmingChaos is true once chaos is selected and awaiting confirmation.
	confirmingChaos bool
	// chaosMode is the mode of the running session.
	chaosMode bool
	// sessionID increments per session so kills can never outlive their session.
	sessionID  uint64
	deathTimer int
	drawTick   uint64
}

type star struct {
	x, y, brightness float64
}

// NewGame creates the game wired to the given bridge channels.
func NewGame(config Config, gameEvents <-chan GameEvent, killEvents chan<- KillEvent, sessionEvents chan<- SessionEvent) *Game {
	g := &Game{
		config:        config,
		state:         StateTitle,
		player:        NewPlayer(screenWidth/2, screenHeight/2),
		scene:         NewScene(gameEvents, killEvents),
		sessionEvents: sessionEvents,
		guide:         newFieldGuide(),
	}
	g.stars = generateStars(120)
	return g
}

func generateStars(count int) []star {
	stars := make([]star, count)
	for i := range stars {
		stars[i] = star{
			x:          float64(i * 137 % int(screenWidth)),
			y:          float64(i * 97 % int(screenHeight)),
			brightness: 0.3 + float64(i%7)*0.1,
		}
	}
	return stars
}

// State returns the game's current top-level state.
func (g *Game) State() State {
	return g.state
}

// Update advances the game state by one tick.
func (g *Game) Update() error {
	g.drawTick++
	switch g.state {
	case StateTitle:
		g.updateTitle()
	case StateModeSelect:
		g.updateModeSelect()
	case StatePlaying:
		g.updatePlaying()
	case StateDying:
		g.updateDying()
	case StateGameOver:
		g.updateGameOver()
	case StateFieldGuide:
		g.handleFieldGuide(readMenuKeys())
	}
	return nil
}

func (g *Game) updateTitle() {
	// Discard anything left over from a session that has ended.
	g.scene.DrainStaleEvents()
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.modeCursor = entryPractice
		g.confirmingChaos = false
		g.scene.ClearError()
		g.state = StateModeSelect
	}
}

// readMenuKeys samples this tick's menu input.
func readMenuKeys() menuKeys {
	k := inpututil.IsKeyJustPressed
	tier := 0
	for i, key := range [...]ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3} {
		if k(key) {
			tier = i + 1
		}
	}
	return menuKeys{
		up:       k(ebiten.KeyArrowUp) || k(ebiten.KeyW),
		down:     k(ebiten.KeyArrowDown) || k(ebiten.KeyS),
		left:     k(ebiten.KeyArrowLeft) || k(ebiten.KeyA),
		right:    k(ebiten.KeyArrowRight) || k(ebiten.KeyD),
		enter:    k(ebiten.KeyEnter),
		esc:      k(ebiten.KeyEscape),
		back:     k(ebiten.KeyBackspace),
		tier:     tier,
		notReady: k(ebiten.KeyN),
		restarts: k(ebiten.KeyP),
		reboot:   k(ebiten.KeyR),
	}
}

func (g *Game) updateModeSelect() {
	g.handleModeSelect(readMenuKeys())
}

func (g *Game) handleModeSelect(in menuKeys) {
	g.scene.DrainStaleEvents()

	if in.esc {
		if g.confirmingChaos {
			g.confirmingChaos = false
			return
		}
		g.state = StateTitle
		return
	}

	if !g.confirmingChaos {
		switch {
		case in.up:
			g.modeCursor = (g.modeCursor + modeEntryCount - 1) % modeEntryCount
		case in.down:
			g.modeCursor = (g.modeCursor + 1) % modeEntryCount
		}
	}

	if !in.enter {
		return
	}

	if g.modeCursor == entryGuide {
		g.guide.open()
		g.state = StateFieldGuide
		return
	}

	// Chaos deletes real Pods, so it always costs a second, explicit press.
	if g.modeCursor == entryChaos && !g.confirmingChaos {
		g.confirmingChaos = true
		return
	}
	g.startSession(g.modeCursor == entryChaos)
}

// handleFieldGuide runs one tick of the Field guide. It never touches the
// session channel: the guide is not a session.
func (g *Game) handleFieldGuide(in menuKeys) {
	g.scene.DrainStaleEvents()
	if g.guide.update(in) {
		g.state = StateModeSelect
	}
}

func (g *Game) startSession(chaosMode bool) {
	g.chaosMode = chaosMode
	g.confirmingChaos = false
	g.sessionID++
	g.scene.BeginSession(g.sessionID, chaosMode)
	g.player.Reset(screenWidth/2, screenHeight/2)
	g.sessionEvents <- SessionEvent{
		Type:      EventSessionStart,
		ID:        g.sessionID,
		Namespace: g.config.Namespace,
		ChaosMode: chaosMode,
	}
	g.state = StatePlaying
}

func (g *Game) endSession() {
	g.sessionEvents <- SessionEvent{Type: EventSessionEnd}
	g.scene.Reset()
	g.state = StateTitle
}

func (g *Game) updatePlaying() {
	projectiles := g.player.Update()
	g.scene.AddProjectiles(projectiles)
	g.scene.Update(g.player)

	// A watch that failed leaves a session with no enemies; say why.
	if g.scene.LastError() != "" {
		g.sessionEvents <- SessionEvent{Type: EventSessionEnd}
		g.scene.Reset()
		g.state = StateModeSelect
		return
	}

	if g.player.Health <= 0 {
		g.player.RollTicks = 0
		g.deathTimer = deathTicks
		g.state = StateDying
	}
}

// updateDying advances the world with the ship destroyed. Nothing is fired, so
// no kill can escape during the death animation.
func (g *Game) updateDying() {
	g.scene.UpdateWithoutPlayer(g.player.X, g.player.Y)
	g.deathTimer--
	if g.deathTimer <= 0 {
		g.state = StateGameOver
	}
}

func (g *Game) updateGameOver() {
	g.scene.DrainStaleEvents()
	// In-flight Bolts finish their flight, projectile hit flashes keep fading,
	// and nothing can hit or fire.
	g.scene.updateBolts()
	g.scene.updateProjectileImpacts()

	// Retry keeps the session and its enemy pool: the Pods never died.
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.scene.ClearBolts()
		g.player.Reset(screenWidth/2, screenHeight/2)
		g.state = StatePlaying
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.endSession()
	}
}

// Layout returns the game's logical screen dimensions.
func (g *Game) Layout(_, _ int) (int, int) {
	return int(screenWidth), int(screenHeight)
}

var (
	colorBackground             = color.NRGBA{R: 5, G: 5, B: 15, A: 255}
	colorCrosshair              = color.NRGBA{R: 0, G: 255, B: 255, A: 140}
	colorEnemyFlash             = color.NRGBA{R: 255, G: 220, B: 220, A: 255}
	colorEnemyDamage            = color.NRGBA{R: 255, G: 180, B: 80, A: 255}
	colorHealthGood             = color.NRGBA{R: 60, G: 220, B: 90, A: 255}
	colorHealthLow              = color.NRGBA{R: 230, G: 160, B: 40, A: 255}
	colorHealthBack             = color.NRGBA{R: 40, G: 40, B: 55, A: 255}
	colorWeaponPanel            = color.NRGBA{R: 15, G: 20, B: 35, A: 220}
	colorWeaponPanelBorder      = color.NRGBA{R: 95, G: 115, B: 145, A: 255}
	colorWeaponCooldownCharging = color.NRGBA{R: 230, G: 160, B: 40, A: 255}
	colorWeaponCooldownReady    = color.NRGBA{R: 60, G: 220, B: 90, A: 255}
)
