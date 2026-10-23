package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type rollShipPen struct {
	dst             *ebiten.Image
	cx, cy, c, s, k float32
	squash          float32
}

func newRollShipPen(dst *ebiten.Image, x, y, heading float64, scale, squash float32) rollShipPen {
	return rollShipPen{
		dst:    dst,
		cx:     float32(x),
		cy:     float32(y),
		c:      float32(math.Cos(heading)),
		s:      float32(math.Sin(heading)),
		k:      scale,
		squash: squash,
	}
}

func (p rollShipPen) pt(x, y float32) (float32, float32) {
	squashedY := y * p.squash
	return p.cx + (x*p.c-squashedY*p.s)*p.k, p.cy + (x*p.s+squashedY*p.c)*p.k
}

func (p rollShipPen) poly(c color.Color, pts ...float32) {
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

func (p rollShipPen) mirror(c color.Color, pts ...float32) {
	p.poly(c, pts...)
	mirrored := make([]float32, len(pts))
	for i := 0; i < len(pts); i += 2 {
		mirrored[i] = pts[i]
		mirrored[i+1] = -pts[i+1]
	}
	p.poly(c, mirrored...)
}

func (p rollShipPen) ellipse(x, y, rx, ry float32, c color.Color) {
	const steps = 24
	pts := make([]float32, 0, steps*2)
	for i := range steps {
		a := 2 * math.Pi * float64(i) / steps
		pts = append(pts,
			x+rx*float32(math.Cos(a)),
			y+ry*float32(math.Sin(a)),
		)
	}
	p.poly(c, pts...)
}

func (p rollShipPen) circle(x, y, r float32, c color.Color) {
	p.ellipse(x, y, r, r, c)
}

func (p rollShipPen) line(x1, y1, x2, y2, width float32, c color.Color) {
	ax, ay := p.pt(x1, y1)
	bx, by := p.pt(x2, y2)
	vector.StrokeLine(p.dst, ax, ay, bx, by, width*p.k, c, true)
}

func (p rollShipPen) plume(x, y, halfWidth float32, look shipLook, seed float64) {
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

func drawRollingShip(screen *ebiten.Image, player *Player, look shipLook) bool {
	if !player.IsRolling() {
		return false
	}

	progress := rollProgress(player.RollTicks)
	pen := newRollShipPen(
		screen,
		player.X,
		player.Y,
		player.Heading,
		float32(rollLiftScale(progress)),
		float32(rollHullWidthScale(progress)),
	)
	if math.Cos(2*math.Pi*progress) < 0 {
		drawInterceptorBelly(pen, look)
		return true
	}
	drawRollingInterceptorTop(pen, look, progress)
	return true
}

func rollProgress(rollTicks int) float64 {
	if rollTicks <= 0 {
		return 0
	}
	return 1 - float64(rollTicks)/float64(barrelRollDurationTicks)
}

func rollHullWidthScale(progress float64) float64 {
	return math.Max(0.12, math.Abs(math.Cos(2*math.Pi*progress)))
}

func rollLiftScale(progress float64) float64 {
	return 1 + 0.18*math.Sin(math.Pi*progress)
}

func drawRollingInterceptorTop(p rollShipPen, look shipLook, progress float64) {
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

	drawRollHighlight(p, progress)
}

func drawInterceptorBelly(p rollShipPen, look shipLook) {
	gunmetal := rgba(70, 78, 98)
	dark := rgba(35, 40, 55)
	line := rgba(110, 120, 140)

	p.plume(-11, -3, 2, look, 0)
	p.plume(-11, 3, 2, look, 1.7)

	p.mirror(dark, 4, -3, -7, -12, -11, -12, -8, -3)
	p.mirror(gunmetal, 2, -3, -7, -11, -8, -11, -5, -3)
	p.poly(gunmetal, 14, 0, 6, -3.5, -9, -4.5, -11, -2.5, -11, 2.5, -9, 4.5, 6, 3.5)
	p.mirror(dark, -9, -4.5, -12, -4.5, -12, -1.5, -9, -1.5)

	p.line(8.5, 0, -9.5, 0, 0.6, line)
	p.line(2.5, -3.1, 2.5, 3.1, 0.6, line)
	p.line(-4.2, -4.2, -4.2, 4.2, 0.6, line)
	p.mirror(dark, -9, -4.6, -12, -4.6, -12, -1.7, -9, -1.7)
}

func drawRollHighlight(p rollShipPen, progress float64) {
	strength := 1 - rollHullWidthScale(progress)
	if strength <= 0 {
		return
	}

	center := float32(-7 + 18*progress)
	lead := center + 4.5
	trail := center - 3.5
	if lead > 13.5 {
		lead = 13.5
	}
	if trail < -10.5 {
		trail = -10.5
	}

	p.poly(
		fade(rgba(255, 255, 255), 0.4*strength),
		lead, 0,
		center, -2.2,
		trail, -2.9,
		trail, 2.9,
		center, 2.2,
	)
}
