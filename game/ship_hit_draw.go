package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const shipHitFxTicks = 45

func (g *Game) playerShipLook() shipLook {
	return shipLook{
		tick:   float64(g.drawTick),
		speed:  math.Min(1, math.Hypot(g.player.VX, g.player.VY)/playerMaxSpeed),
		thrust: g.player.thrustHeld,
	}
}

func (g *Game) drawPlayerHitTreatment(screen *ebiten.Image, look shipLook) bool {
	if g.player.IsRolling() || !g.player.IsInvulnerable() {
		return false
	}

	since := invulnerableTicks - g.player.InvulnerableTimer
	x, y := g.hitShakeOffset(g.player.X, g.player.Y, since)
	p := newShipPen(screen, x, y, g.player.Heading, 1)

	g.drawHitSmoke(p, since)
	if g.hitEngineVisible(since) {
		p.plume(-11, -3, 2, look, 0)
		p.plume(-11, 3, 2, look, 1.7)
	}
	drawHitInterceptorHull(p, hitShipPalette())
	drawHitInterceptorHull(p, g.hitPulsePalette(since))
	g.drawHitSparks(p, since)
	return true
}

func (g *Game) hitShakeOffset(x, y float64, since int) (float64, float64) {
	if since >= 10 {
		return x, y
	}
	t := float64(since) / 10
	amp := 1.6 * (1 - t) * (1 - t)
	tick := float64(g.drawTick)
	return x + math.Sin(tick*2.7)*amp, y + math.Cos(tick*3.3)*amp
}

func (g *Game) hitEngineVisible(since int) bool {
	return since >= shipHitFxTicks || hitHash01(float64(g.drawTick)*1.3) >= 0.3
}

func (g *Game) drawHitSmoke(p shipPen, since int) {
	remaining := float64(invulnerableTicks-since) / invulnerableTicks
	if remaining <= 0 {
		return
	}
	for i := range 4 {
		f := math.Mod(float64(since)*0.05+float64(i)*0.25, 1)
		y := float32(math.Sin(float64(i)*2.1) * f * 4)
		p.circle(float32(-10-f*14), y, float32(1.5+3*f), fade(rgba(70, 70, 80), 0.45*(1-f)*remaining))
	}
}

func (g *Game) drawHitSparks(p shipPen, since int) {
	if since >= shipHitFxTicks {
		return
	}
	for i := range 6 {
		f := math.Mod(float64(since)*0.07+float64(i)/6, 1)
		sx := float32(hitHash01(float64(i)+0.3)*20 - 10)
		sy := float32(hitHash01(float64(i)+7.1)*16 - 8)
		angle := hitHash01(float64(i)+3.7) * 2 * math.Pi
		d := float32(f * 10)
		p.circle(sx+float32(math.Cos(angle))*d, sy+float32(math.Sin(angle))*d, 0.8,
			fade(mix(rgba(255, 240, 150), rgba(255, 110, 30), f), 1-f))
	}
}

func (g *Game) hitPulsePalette(since int) shipPalette {
	phase := float64(since%6) / 6
	alpha := 0.18 + 0.24*(1-phase)
	if (since/6)%2 == 1 {
		alpha *= 0.55
	}
	red := rgba(255, 70, 60)
	base := hitShipPalette()
	return shipPalette{
		steel:  fade(mix(base.steel, red, 0.65), alpha),
		dark:   fade(mix(base.dark, red, 0.45), alpha),
		light:  fade(mix(base.light, red, 0.7), alpha),
		accent: fade(mix(base.accent, red, 0.35), alpha),
	}
}

func hitHash01(a float64) float64 {
	v := math.Sin(a*12.9898) * 43758.5453
	return v - math.Floor(v)
}

type shipPalette struct {
	steel  color.NRGBA
	dark   color.NRGBA
	light  color.NRGBA
	accent color.NRGBA
}

func hitShipPalette() shipPalette {
	return shipPalette{
		steel:  rgba(150, 165, 190),
		dark:   rgba(55, 65, 90),
		light:  rgba(215, 225, 240),
		accent: rgba(0, 200, 255),
	}
}

func drawHitInterceptorHull(p shipPen, pal shipPalette) {
	p.mirror(pal.dark, 4, -3, -7, -12, -11, -12, -8, -3)
	p.mirror(pal.steel, 2, -3, -7, -11, -8, -11, -5, -3)

	p.poly(pal.steel, 14, 0, 6, -3.5, -9, -4.5, -11, -2.5, -11, 2.5, -9, 4.5, 6, 3.5)
	p.poly(pal.light, 14, 0, 6, -2, -6, -2, -6, 2, 6, 2)

	p.mirror(pal.dark, -9, -4.5, -12, -4.5, -12, -1.5, -9, -1.5)

	p.poly(rgba(10, 30, 60), 9, 0, 5, -2, 0, -1.8, 0, 1.8, 5, 2)
	p.poly(fade(pal.accent, 0.8), 8, -0.3, 5, -1.3, 3, -1, 5, -0.3)
}
