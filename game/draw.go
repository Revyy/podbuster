package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Draw renders the current game state to screen.
func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)

	switch g.state {
	case StateTitle:
		g.drawTitle(screen)
	case StateModeSelect:
		g.drawModeSelect(screen)
	case StatePlaying:
		g.drawWorld(screen, true)
		g.drawHUD(screen)
	case StateDying:
		g.drawWorld(screen, false)
		g.drawHUD(screen)
		g.drawShipDeath(screen)
	case StateGameOver:
		g.drawWorld(screen, false)
		g.drawHUD(screen)
		g.drawGameOver(screen)
	case StateFieldGuide:
		g.guide.draw(screen)
	}
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	screen.Fill(colorBackground)
	for _, s := range g.stars {
		b := uint8(s.brightness * 255)
		screen.Set(int(s.x), int(s.y), color.NRGBA{R: b, G: b, B: b, A: 255})
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "P O D B U S T E R", 330, 240)
	ebitenutil.DebugPrintAt(screen, "Press ENTER to start", 325, 300)
	if err := g.scene.LastError(); err != "" {
		ebitenutil.DebugPrintAt(screen, "Last error: "+err, 20, 560)
	}
}

func (g *Game) drawModeSelect(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "SELECT MODE", 355, 180)
	ebitenutil.DebugPrintAt(screen, "Namespace: "+g.config.Namespace+"  (set with --namespace)", 260, 205)

	markers := [modeEntryCount]string{"  ", "  ", "  "}
	markers[g.modeCursor] = "> "
	ebitenutil.DebugPrintAt(screen, markers[entryPractice]+"PRACTICE   kills are scored, Pods are left alone", 230, 260)
	ebitenutil.DebugPrintAt(screen, markers[entryChaos]+"CHAOS      kills DELETE the real Pod", 230, 285)
	vector.StrokeLine(screen, 230, 311, 560, 311, 1, colorWeaponPanelBorder, false)
	ebitenutil.DebugPrintAt(screen, markers[entryGuide]+"FIELD GUIDE   enemy archetypes (not a mode)", 230, 318)

	if g.confirmingChaos {
		vector.FillRect(screen, 130, 340, 540, 90, color.NRGBA{R: 60, G: 10, B: 10, A: 230}, false)
		ebitenutil.DebugPrintAt(screen, "This will DELETE real Pods in namespace "+g.config.Namespace+".", 150, 360)
		ebitenutil.DebugPrintAt(screen, "Press ENTER again to confirm, ESC to back out.", 150, 385)
	} else {
		ebitenutil.DebugPrintAt(screen, "UP/DOWN to choose, ENTER to select, ESC for title", 230, 340)
	}

	if err := g.scene.LastError(); err != "" {
		ebitenutil.DebugPrintAt(screen, "Could not watch namespace: "+err, 20, 560)
	}
}

// drawWorld renders projectiles, enemies and (optionally) the ship.
func (g *Game) drawWorld(screen *ebiten.Image, withPlayer bool) {
	for _, projectile := range g.scene.Projectiles() {
		g.drawPlayerProjectile(screen, projectile)
	}

	for _, e := range g.scene.Enemies() {
		g.drawEnemy(screen, e)
	}
	g.drawProjectileImpacts(screen)
	g.drawBolts(screen)

	if withPlayer {
		g.drawCrosshair(screen)
	}

	if withPlayer {
		g.drawPlayerShip(screen)
	}
}

func (g *Game) drawCrosshair(screen *ebiten.Image) {
	x, y := g.player.AimPoint()
	vector.StrokeCircle(screen, float32(x), float32(y), 6, 1, colorCrosshair, false)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	g.drawHealthBar(screen)
	g.drawWeaponStatus(screen)
	g.drawRollStatus(screen)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Enemies: %d", len(g.scene.Enemies())), 690, 8)

	// The number means something different per mode, so the label does too.
	counterLabel := "Kills"
	if g.chaosMode {
		counterLabel = "Pods deleted"
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s: %d", counterLabel, g.scene.Kills()), 690, 28)

	if g.chaosMode {
		vector.FillRect(screen, 320, 4, 160, 22, color.NRGBA{R: 80, G: 0, B: 0, A: 220}, false)
		ebitenutil.DebugPrintAt(screen, "!! CHAOS MODE !!", 340, 10)
	}
}

func (g *Game) drawWeaponStatus(screen *ebiten.Image) {
	const (
		panelX      = 8.0
		panelY      = 500.0
		panelWidth  = 170.0
		panelHeight = 46.0
		barX        = panelX + 6
		barY        = panelY + 37
		barWidth    = panelWidth - 12
		barHeight   = 4.0
	)

	vector.FillRect(screen, panelX, panelY, panelWidth, panelHeight, colorWeaponPanel, false)
	vector.StrokeRect(screen, panelX, panelY, panelWidth, panelHeight, 1, colorWeaponPanelBorder, false)
	mainName, _, _ := weaponHUDStatus(g.player.MainWeapon)
	secondaryName, remaining, total := weaponHUDStatus(g.player.SecondaryWeapon)
	ebitenutil.DebugPrintAt(screen, "SPACE  "+mainName, int(panelX)+6, int(panelY)+6)
	ebitenutil.DebugPrintAt(screen, "X      "+secondaryName, int(panelX)+6, int(panelY)+20)

	ratio := cooldownFillRatio(remaining, total)
	fill := colorWeaponCooldownCharging
	if ratio == 1 {
		fill = colorWeaponCooldownReady
	}
	vector.FillRect(screen, barX, barY, barWidth, barHeight, colorHealthBack, false)
	vector.FillRect(screen, barX, barY, float32(barWidth*ratio), barHeight, fill, false)
}

func (g *Game) drawRollStatus(screen *ebiten.Image) {
	const (
		panelX      = 8.0
		panelY      = 550.0
		panelWidth  = 170.0
		panelHeight = 26.0
		barX        = panelX + 6
		barY        = panelY + 18
		barWidth    = panelWidth - 12
		barHeight   = 4.0
	)

	ratio := cooldownFillRatio(g.player.RollCooldown, barrelRollCooldownTicks)
	fill := colorWeaponCooldownCharging
	if ratio == 1 {
		fill = colorWeaponCooldownReady
	}

	vector.FillRect(screen, panelX, panelY, panelWidth, panelHeight, colorWeaponPanel, false)
	vector.StrokeRect(screen, panelX, panelY, panelWidth, panelHeight, 1, colorWeaponPanelBorder, false)
	ebitenutil.DebugPrintAt(screen, "Z      Roll", int(panelX)+6, int(panelY)+2)
	vector.FillRect(screen, barX, barY, barWidth, barHeight, colorHealthBack, false)
	vector.FillRect(screen, barX, barY, float32(barWidth*ratio), barHeight, fill, false)
}

func (g *Game) drawHealthBar(screen *ebiten.Image) {
	const barWidth, barHeight = 140.0, 14.0
	vector.FillRect(screen, 8, 8, barWidth, barHeight, colorHealthBack, false)

	ratio := float64(g.player.Health) / float64(PlayerMaxHealth)
	if ratio < 0 {
		ratio = 0
	}
	fill := colorHealthGood
	if ratio <= 0.3 {
		fill = colorHealthLow
	}
	vector.FillRect(screen, 8, 8, float32(barWidth*ratio), barHeight, fill, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP %d/%d", g.player.Health, PlayerMaxHealth), 156, 8)
}

func cooldownFillRatio(remaining, total int) float64 {
	if total <= 0 || remaining <= 0 {
		return 1
	}
	if remaining >= total {
		return 0
	}
	return float64(total-remaining) / float64(total)
}

func weaponHUDStatus(weapon Weapon) (name string, remaining, total int) {
	status, ok := weapon.(WeaponStatus)
	if !ok {
		return "Unknown", 0, 1
	}
	return status.Name(), status.CooldownRemaining(), status.CooldownTotal()
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.FillRect(screen, 0, 0, screenWidth, screenHeight, color.NRGBA{R: 0, G: 0, B: 0, A: 170}, false)
	ebitenutil.DebugPrintAt(screen, "G A M E   O V E R", 335, 250)

	summary := fmt.Sprintf("Kills: %d", g.scene.Kills())
	if g.chaosMode {
		summary = fmt.Sprintf("Pods deleted: %d", g.scene.Kills())
	}
	ebitenutil.DebugPrintAt(screen, summary, 355, 280)
	ebitenutil.DebugPrintAt(screen, "R to retry in the same mode", 310, 320)
	ebitenutil.DebugPrintAt(screen, "ESC to return to the title screen", 300, 340)

	if g.chaosMode {
		vector.FillRect(screen, 285, 370, 240, 20, color.NRGBA{R: 60, G: 0, B: 0, A: 200}, false)
		ebitenutil.DebugPrintAt(screen, "Chaos mode is still active on retry.", 290, 373)
	}
}

func (g *Game) drawBolts(screen *ebiten.Image) {
	for _, b := range g.scene.Bolts() {
		drawBoltAt(screen, float32(b.X), float32(b.Y), 1, b.Archetype)
	}
	for _, burst := range g.scene.Bursts() {
		pal := droidPalettes[burst.Archetype]
		t := float32(burst.Age) / burstTicks
		vector.StrokeCircle(screen, float32(burst.X), float32(burst.Y), 3+10*t, 2, withAlpha(pal.light, uint8(255*(1-t))), false)
	}
}

func withAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}
