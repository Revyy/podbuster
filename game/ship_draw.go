package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type shipLook struct {
	tick   float64
	speed  float64
	thrust bool
}

// shipPen draws in ship-local coordinates where +x points along the nose and
// +y points to the ship's right.
type shipPen struct {
	dst          *ebiten.Image
	cx, cy, c, s float32
	k            float32
}

func newShipPen(dst *ebiten.Image, x, y, heading float64, scale float32) shipPen {
	return shipPen{
		dst: dst,
		cx:  float32(x),
		cy:  float32(y),
		c:   float32(math.Cos(heading)),
		s:   float32(math.Sin(heading)),
		k:   scale,
	}
}

func (p shipPen) pt(x, y float32) (float32, float32) {
	return p.cx + (x*p.c-y*p.s)*p.k, p.cy + (x*p.s+y*p.c)*p.k
}

func (p shipPen) poly(c color.Color, pts ...float32) {
	var path vector.Path
	for i := 0; i < len(pts); i += 2 {
		x, y := p.pt(pts[i], pts[i+1])
		if i == 0 {
			path.MoveTo(x, y)
			continue
		}
		path.LineTo(x, y)
	}
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(p.dst, &path, nil, op)
}

func (p shipPen) circle(x, y, r float32, c color.Color) {
	px, py := p.pt(x, y)
	vector.FillCircle(p.dst, px, py, r*p.k, c, true)
}

func (p shipPen) mirror(c color.Color, pts ...float32) {
	p.poly(c, pts...)
	mirrored := make([]float32, len(pts))
	for i := 0; i < len(pts); i += 2 {
		mirrored[i] = pts[i]
		mirrored[i+1] = -pts[i+1]
	}
	p.poly(c, mirrored...)
}

var (
	engineGlowOuter = rgba(40, 110, 255)
	engineGlowMid   = rgba(90, 180, 255)
	engineGlowCore  = rgba(220, 240, 255)
)

func (p shipPen) plume(x, y, halfWidth float32, look shipLook, seed float64) {
	speed := float32(look.speed)
	flicker := float32(1 + 0.12*math.Sin(look.tick*0.9+seed) + 0.06*math.Sin(look.tick*2.3+seed*3))
	length := (2.5 + 9*speed) * flicker
	width := halfWidth * (0.75 + 0.35*speed)
	brightness := 0.45 + 0.35*float64(speed)
	if look.thrust {
		length += 7 * flicker
		width *= 1.25
		brightness = 1
	}

	p.circle(x, y, width*1.6, fade(engineGlowOuter, 0.25*brightness))
	p.poly(fade(engineGlowOuter, 0.45*brightness), x, y-width*1.3, x-length*1.25, y, x, y+width*1.3)
	p.poly(fade(engineGlowMid, 0.7*brightness), x, y-width, x-length*0.8, y, x, y+width)
	p.poly(fade(engineGlowCore, 0.9*brightness), x, y-width*0.5, x-length*0.4, y, x, y+width*0.5)
	p.circle(x, y, width*0.6, fade(engineGlowCore, brightness))
}

func (g *Game) drawPlayerShip(screen *ebiten.Image) {
	look := g.playerShipLook()
	if g.drawPlayerHitTreatment(screen, look) {
		return
	}
	if drawRollingShip(screen, g.player, look) {
		return
	}
	drawInterceptor(newShipPen(screen, g.player.X, g.player.Y, g.player.Heading, 1), look)
}

func drawInterceptor(p shipPen, look shipLook) {
	steel := rgba(150, 165, 190)
	dark := rgba(55, 65, 90)
	light := rgba(215, 225, 240)
	accent := rgba(0, 200, 255)

	p.plume(-11, -3, 2, look, 0)
	p.plume(-11, 3, 2, look, 1.7)

	p.mirror(dark, 4, -3, -7, -12, -11, -12, -8, -3)
	p.mirror(steel, 2, -3, -7, -11, -8, -11, -5, -3)

	p.poly(steel, 14, 0, 6, -3.5, -9, -4.5, -11, -2.5, -11, 2.5, -9, 4.5, 6, 3.5)
	p.poly(light, 14, 0, 6, -2, -6, -2, -6, 2, 6, 2)

	p.mirror(dark, -9, -4.5, -12, -4.5, -12, -1.5, -9, -1.5)

	p.poly(rgba(10, 30, 60), 9, 0, 5, -2, 0, -1.8, 0, 1.8, 5, 2)
	p.poly(fade(accent, 0.8), 8, -0.3, 5, -1.3, 3, -1, 5, -0.3)
}
