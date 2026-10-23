package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type shipDebrisPiece struct {
	pts  []float32
	col  color.NRGBA
	spin float64
}

var interceptorDeathPieces = []shipDebrisPiece{
	{
		pts:  []float32{14, 0, 6, -3.5, 2, -3.7, 2, 3.7, 6, 3.5},
		col:  rgba(150, 165, 190),
		spin: 0.9,
	},
	{
		pts:  []float32{2, -3.7, -9, -4.5, -11, -2.5, -11, 2.5, -9, 4.5, 2, 3.7},
		col:  rgba(130, 145, 170),
		spin: -0.6,
	},
	{
		pts:  []float32{4, -3, -7, -12, -11, -12, -8, -3},
		col:  rgba(55, 65, 90),
		spin: 2.2,
	},
	{
		pts:  []float32{4, 3, -7, 12, -11, 12, -8, 3},
		col:  rgba(55, 65, 90),
		spin: -2.6,
	},
}

func (g *Game) drawShipDeath(screen *ebiten.Image) {
	progress := float64(deathTicks-g.deathTimer) / float64(deathTicks)
	drawInterceptorDeath(screen, g.player.X, g.player.Y, g.player.Heading, progress, g.player.VX, g.player.VY)
}

func drawInterceptorDeath(screen *ebiten.Image, x, y, heading, t, vx, vy float64) {
	if t < 0.35 {
		fireball := t / 0.35
		p := newShipPen(screen, x, y, 0, 1)
		p.circle(0, 0, float32(4+16*fireball), fade(fireOrange, 0.7*(1-fireball)))
		p.circle(0, 0, float32(2+8*fireball), fade(fireYellow, 1-fireball))
	}

	ease := 1 - math.Pow(1-t, 2.2)
	driftX := vx * ease * 14
	driftY := vy * ease * 14
	cosHeading := math.Cos(heading)
	sinHeading := math.Sin(heading)
	hot := rgba(255, 150, 60)
	char := rgba(40, 35, 35)

	for i, piece := range interceptorDeathPieces {
		cx, cy := pieceCentroid(piece.pts)
		dx, dy := unit(cx, cy)
		dist := float32(ease * (22 + 6*float64(i)))
		px := cx + dx*dist
		py := cy + dy*dist

		rotated := make([]float32, len(piece.pts))
		for j := 0; j < len(piece.pts); j += 2 {
			rotated[j] = piece.pts[j] - cx
			rotated[j+1] = piece.pts[j+1] - cy
		}

		col := mix(hot, piece.col, math.Min(1, t*4))
		col = mix(col, char, math.Max(0, t-0.4))

		offsetX := float64(px)*cosHeading - float64(py)*sinHeading
		offsetY := float64(px)*sinHeading + float64(py)*cosHeading

		piecePen := newShipPen(
			screen,
			x+offsetX+driftX,
			y+offsetY+driftY,
			heading+piece.spin*t*3,
			1,
		)
		piecePen.poly(fade(col, math.Min(1, (1-t)*2.2)), rotated...)
	}

	p := newShipPen(screen, x, y, 0, 1)
	for i := range 8 {
		angle := emberHash(float64(i)) * 2 * math.Pi
		dist := float32(ease * (10 + 30*emberHash(float64(i)+2)))
		p.circle(float32(math.Cos(angle))*dist, float32(math.Sin(angle))*dist, 0.9, fade(fireYellow, 1-t))
	}
}

func pieceCentroid(pts []float32) (float32, float32) {
	var x, y float32
	for i := 0; i < len(pts); i += 2 {
		x += pts[i]
		y += pts[i+1]
	}
	count := float32(len(pts) / 2)
	return x / count, y / count
}

func unit(x, y float32) (float32, float32) {
	length := float32(math.Hypot(float64(x), float64(y)))
	if length == 0 {
		return 0, 0
	}
	return x / length, y / length
}

func emberHash(v float64) float64 {
	hash := math.Sin(v*12.9898) * 43758.5453
	return hash - math.Floor(hash)
}
