package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	blasterMuzzleTicks = 3
	rocketMuzzleTicks  = 4
)

var (
	tracerGlow  = rgba(40, 140, 255)
	tracerMid   = rgba(80, 220, 255)
	tracerCore  = rgba(235, 252, 255)
	fireRed     = rgba(230, 60, 30)
	fireOrange  = rgba(255, 140, 30)
	fireYellow  = rgba(255, 220, 90)
	fireWhite   = rgba(255, 250, 220)
	smokeGrey   = rgba(150, 150, 160)
	missileBody = rgba(225, 228, 235)
	missileBand = rgba(120, 125, 135)
	missileNose = rgba(210, 50, 45)
	missileFins = rgba(170, 40, 40)
)

func (g *Game) drawPlayerProjectile(screen *ebiten.Image, projectile *Projectile) {
	if projectile == nil || !projectile.Active {
		return
	}

	heading := math.Atan2(projectile.VY, projectile.VX)
	speed := math.Hypot(projectile.VX, projectile.VY)
	p := newShotPen(screen, projectile.X, projectile.Y, heading, 1)

	if projectile.Kind == projectileKindRocket {
		drawClassicMissile(p, projectile.Age, speed, float64(g.drawTick), shotSeed(projectile))
		return
	}
	drawTwinTracers(p, projectile.Age, speed)
}

func (g *Game) drawProjectileImpacts(screen *ebiten.Image) {
	for _, impact := range g.scene.ProjectileImpacts() {
		t := float64(impact.Age) / projectileImpactTicks
		p := newShotPen(screen, impact.X, impact.Y, impact.Direction, 1)
		if impact.Kind == projectileKindRocket {
			p.circle(0, 0, float32(4+12*t), fade(fireOrange, 0.6*(1-t)))
			p.circle(0, 0, float32(2+6*t), fade(fireYellow, 0.9*(1-t)))
			px, py := p.pt(0, 0)
			vector.StrokeCircle(screen, px, py, float32(6+16*t), 1.5, fade(smokeGrey, 0.5*(1-t)), true)
			continue
		}
		p.circle(0, 0, float32(2+4*t), fade(tracerMid, 0.6*(1-t)))
		for i := range 5 {
			a := math.Pi + (float64(i)-2)*0.5
			r0, r1 := 2+6*t, 4+10*t
			p.line(
				float32(math.Cos(a)*r0), float32(math.Sin(a)*r0),
				float32(math.Cos(a)*r1), float32(math.Sin(a)*r1),
				1, fade(tracerCore, 1-t),
			)
		}
	}
}

type shotPen struct {
	dst          *ebiten.Image
	cx, cy, c, s float32
	k            float32
}

func newShotPen(dst *ebiten.Image, x, y, heading float64, scale float32) shotPen {
	return shotPen{
		dst: dst,
		cx:  float32(x),
		cy:  float32(y),
		c:   float32(math.Cos(heading)),
		s:   float32(math.Sin(heading)),
		k:   scale,
	}
}

func (p shotPen) pt(x, y float32) (float32, float32) {
	return p.cx + (x*p.c-y*p.s)*p.k, p.cy + (x*p.s+y*p.c)*p.k
}

func (p shotPen) line(x0, y0, x1, y1, w float32, c color.Color) {
	ax, ay := p.pt(x0, y0)
	bx, by := p.pt(x1, y1)
	vector.StrokeLine(p.dst, ax, ay, bx, by, w*p.k, c, true)
}

func (p shotPen) circle(x, y, r float32, c color.Color) {
	px, py := p.pt(x, y)
	vector.FillCircle(p.dst, px, py, r*p.k, c, true)
}

func (p shotPen) poly(c color.Color, pts ...float32) {
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

func (p shotPen) mirror(c color.Color, pts ...float32) {
	p.poly(c, pts...)
	reflected := make([]float32, len(pts))
	for i := 0; i < len(pts); i += 2 {
		reflected[i], reflected[i+1] = pts[i], -pts[i+1]
	}
	p.poly(c, reflected...)
}

func drawTwinTracers(p shotPen, age int, speed float64) {
	drawTracerMuzzle(p, age, speed)

	tail := float32(math.Min(10, 3+float64(age)*speed))
	for _, y := range []float32{-2.5, 2.5} {
		p.line(1, y*0.8, -tail, y, 3, fade(tracerGlow, 0.3))
		p.line(1, y*0.8, -tail, y, 1.3, tracerMid)
		p.circle(1, y*0.8, 1, tracerCore)
	}
}

func drawTracerMuzzle(p shotPen, age int, speed float64) {
	if age >= blasterMuzzleTicks {
		return
	}
	t := float64(age) / float64(blasterMuzzleTicks)
	x := -float32(float64(age) * speed)
	r := float32(3 + 3*t)
	p.circle(x, 0, r, fade(tracerMid, 0.6*(1-t)))
	p.line(x-r, 0, x+r*1.4, 0, 1, fade(tracerCore, 1-t))
	p.line(x, -r, x, r, 1, fade(tracerCore, 0.7*(1-t)))
}

func drawClassicMissile(p shotPen, age int, speed, tick, seed float64) {
	drawSmokePuffs(p, -11, age, speed, 18, 2)
	drawRocketMuzzle(p, age, speed)
	drawFlame(p, -10, 1.6, 7, tick, seed)
	p.mirror(missileFins, -6, -1.6, -10.5, -4.2, -10.5, -1.6)
	p.poly(missileBody, 0, -1.6, -10, -1.6, -10, 1.6, 0, 1.6)
	p.line(-3, -1.6, -3, 1.6, 0.8, missileBand)
	p.poly(missileNose, 3.5, 0, 0, -1.6, 0, 1.6)
}

func drawSmokePuffs(p shotPen, tail float32, age int, speed float64, maxPuffs int, size float32) {
	for puffAge := 1; puffAge <= age && puffAge <= maxPuffs; puffAge += 2 {
		f := float64(puffAge) / float64(maxPuffs)
		x := tail - float32(float64(puffAge)*speed)
		jitter := float64(age - puffAge)
		y := float32(math.Sin(jitter*1.7) * float64(puffAge) * 0.12)
		p.circle(x, y, size*float32(0.6+1.6*f), fade(smokeGrey, 0.35*(1-f)))
	}
}

func drawRocketMuzzle(p shotPen, age int, speed float64) {
	if age >= rocketMuzzleTicks {
		return
	}
	t := float64(age) / float64(rocketMuzzleTicks)
	x := -float32(float64(age) * speed)
	p.circle(x, 0, float32(5+5*t), fade(fireOrange, 0.55*(1-t)))
	p.circle(x, 0, float32(3+2*t), fade(fireWhite, 0.9*(1-t)))
}

func drawFlame(p shotPen, x, w, length float32, tick, seed float64) {
	flicker := float32(1 + 0.22*math.Sin(tick*1.3+seed) + 0.12*math.Sin(tick*3.1+seed*2))
	l := length * flicker
	p.circle(x, 0, w*1.4, fade(fireOrange, 0.3))
	p.poly(fade(fireRed, 0.75), x, -w*1.2, x-l, 0, x, w*1.2)
	p.poly(fade(fireOrange, 0.9), x, -w*0.9, x-l*0.7, 0, x, w*0.9)
	p.poly(fireYellow, x, -w*0.55, x-l*0.42, 0, x, w*0.55)
	p.circle(x, 0, w*0.45, fireWhite)
}

func shotSeed(projectile *Projectile) float64 {
	return math.Mod(math.Abs(projectile.VX*13.1+projectile.VY*7.7), 2*math.Pi)
}
