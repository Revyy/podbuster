package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// droidLook is everything needed to pose one Seeker droid.
type droidLook struct {
	tick        float64
	seed        float64
	vx, vy      float64
	aimX, aimY  float64
	damage      bool
	contact     bool
	terminating bool
	explode     float64 // 0 = not exploding, otherwise progress in (0, 1]
	charge      float64 // telegraph 0..1
	symptoms    symptomLook
}

var droidPalettes = [...]droidPalette{
	ArchetypeProbe:    {hull: rgba(185, 45, 60), dark: rgba(55, 15, 30), light: rgba(235, 110, 120), eye: rgba(120, 230, 255), metal: rgba(210, 80, 90)},
	ArchetypeTurret:   {hull: rgba(215, 140, 30), dark: rgba(60, 38, 10), light: rgba(250, 200, 100), eye: rgba(255, 245, 190), metal: rgba(170, 110, 30)},
	ArchetypeSentinel: {hull: rgba(40, 170, 110), dark: rgba(10, 50, 40), light: rgba(120, 235, 170), eye: rgba(230, 255, 120), metal: rgba(60, 200, 130)},
	ArchetypeDart:     {hull: rgba(150, 70, 220), dark: rgba(35, 15, 60), light: rgba(210, 150, 255), eye: rgba(255, 130, 200), metal: rgba(170, 100, 235)},
}

// droidLookOf derives how an enemy is posed on the given draw tick.
func droidLookOf(e *Enemy, tick uint64, playerX, playerY float64) droidLook {
	l := droidLook{
		tick:        float64(tick),
		seed:        e.seed,
		vx:          e.HeadingX,
		vy:          e.HeadingY,
		damage:      e.State == EnemyActive && e.DamageFlashTimer > 0,
		contact:     e.State == EnemyActive && e.FlashTimer > 0 && (e.FlashTimer/4)%2 == 0,
		terminating: e.State == EnemyTerminating,
		charge:      math.Max(0, math.Min(1, e.Telegraph)),
		symptoms:    symptomLookOf(e, tick),
	}
	if e.State == EnemyExploding {
		l.explode = explosionProgress(e.StateTimer)
	}
	if dx, dy := playerX-e.X, playerY-e.Y; math.Hypot(dx, dy) > 0 {
		d := math.Hypot(dx, dy)
		l.aimX, l.aimY = dx/d, dy/d
	}
	return l
}

// explosionProgress is how far through its explosion an enemy is, in (0, 1].
func explosionProgress(stateTimer int) float64 {
	p := float64(enemyAnimationTicks-stateTimer) / float64(enemyAnimationTicks)
	return math.Max(0.01, math.Min(1, p))
}

func (g *Game) drawEnemy(screen *ebiten.Image, e *Enemy) {
	p := pen{dst: screen, cx: float32(e.X), cy: float32(e.Y), s: float32(e.Scale())}
	drawDroid(p, e.Archetype, e.Tier, droidLookOf(e, g.drawTick, g.player.X, g.player.Y))
}

// pen draws in droid-local units: (0, 0) is the enemy centre, one unit is one
// pixel at scale 1, and the 24x24 hitbox spans -12..12.
type pen struct {
	dst       *ebiten.Image
	cx, cy, s float32
	ox, oy    float32
}

func (p pen) at(dx, dy float32) pen { p.ox += dx; p.oy += dy; return p }

func (p pen) pt(x, y float32) (float32, float32) {
	return p.cx + (x+p.ox)*p.s, p.cy + (y+p.oy)*p.s
}

func (p pen) circle(x, y, r float32, c color.Color) {
	px, py := p.pt(x, y)
	vector.FillCircle(p.dst, px, py, r*p.s, c, true)
}

func (p pen) ring(x, y, r, w float32, c color.Color) {
	px, py := p.pt(x, y)
	vector.StrokeCircle(p.dst, px, py, r*p.s, w*p.s, c, true)
}

func (p pen) line(x0, y0, x1, y1, w float32, c color.Color) {
	ax, ay := p.pt(x0, y0)
	bx, by := p.pt(x1, y1)
	vector.StrokeLine(p.dst, ax, ay, bx, by, w*p.s, c, true)
}

func (p pen) poly(c color.Color, pts ...float32) {
	var path vector.Path
	for i := 0; i < len(pts); i += 2 {
		x, y := p.pt(pts[i], pts[i+1])
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(p.dst, &path, nil, op)
}

func rgba(r, g, b uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: 255} }

func fade(c color.NRGBA, a float64) color.NRGBA {
	c.A = uint8(float64(c.A) * math.Max(0, math.Min(1, a)))
	return c
}

func gray(c color.NRGBA, brightness float64) color.NRGBA {
	l := (0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)) * brightness
	return color.NRGBA{R: uint8(l), G: uint8(l), B: uint8(l * 1.1), A: c.A}
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	f := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t) }
	return color.NRGBA{R: f(a.R, b.R), G: f(a.G, b.G), B: f(a.B, b.B), A: f(a.A, b.A)}
}

// droidPalette is one archetype's colours.
type droidPalette struct{ hull, dark, light, eye, metal color.NRGBA }

func paletteFor(l droidLook, base droidPalette) droidPalette {
	switch {
	case l.terminating:
		pulse := 0.45 + 0.25*(math.Sin(l.tick*2*math.Pi/50)+1)/2
		return droidPalette{gray(base.hull, pulse), gray(base.dark, pulse), gray(base.light, pulse), rgba(40, 40, 50), gray(base.metal, pulse)}
	case l.damage:
		t := 0.75
		return droidPalette{mix(base.hull, colorEnemyDamage, t), mix(base.dark, colorEnemyDamage, 0.4), mix(base.light, colorEnemyDamage, t), rgba(255, 255, 255), mix(base.metal, colorEnemyDamage, 0.5)}
	case l.contact:
		return droidPalette{colorEnemyFlash, mix(base.dark, colorEnemyFlash, 0.5), colorEnemyFlash, rgba(255, 255, 255), mix(base.metal, colorEnemyFlash, 0.6)}
	}
	return base
}

// ------------------------------------------------------------ shape helpers

func (p pen) ellipse(x, y, rx, ry float32, c color.Color) {
	const n = 28
	pts := make([]float32, 0, n*2)
	for i := 0; i < n; i++ {
		a := float64(i) / n * 2 * math.Pi
		pts = append(pts, x+rx*float32(math.Cos(a)), y+ry*float32(math.Sin(a)))
	}
	p.poly(c, pts...)
}

func (p pen) domeE(x, y, rx, ry float32, c color.Color) {
	const n = 16
	pts := make([]float32, 0, (n+1)*2)
	for i := 0; i <= n; i++ {
		a := math.Pi + float64(i)/n*math.Pi
		pts = append(pts, x+rx*float32(math.Cos(a)), y+ry*float32(math.Sin(a)))
	}
	p.poly(c, pts...)
}

// ringArc strokes an ellipse arc from a0 to a1 (radians).
func (p pen) ringArc(rx, ry float32, a0, a1 float64, w float32, c color.Color) {
	const n = 16
	var px, py float32
	for i := 0; i <= n; i++ {
		a := a0 + (a1-a0)*float64(i)/n
		x, y := rx*float32(math.Cos(a)), ry*float32(math.Sin(a))
		if i > 0 {
			p.line(px, py, x, y, w, c)
		}
		px, py = x, y
	}
}

// ------------------------------------------------------------ the droids

func drawDroid(p pen, arch Archetype, tier Tier, l droidLook) {
	pal := paletteFor(l, droidPalettes[arch])
	e := float32(l.explode)
	alpha := 1.0 - float64(e)
	light, heavy := tier == TierLight, tier == TierHeavy
	thin := float32(1)
	if light {
		thin = 0.8
	}
	rx, ry := float32(9)*thin, float32(9)
	switch arch {
	case ArchetypeTurret:
		rx, ry = 11*thin, 7
	case ArchetypeSentinel:
		rx, ry = 8*thin, 8
	case ArchetypeDart:
		rx, ry = 5*thin, 8.5
	}
	bob := float32(math.Sin(l.tick*0.09+l.seed) * 1.2)
	if arch == ArchetypeTurret {
		bob = 0 // planted
	}
	if l.terminating {
		bob = 2.5
	}
	if l.damage {
		p = p.at(float32(math.Sin(l.tick*4))*1.2, 0)
	}
	p = p.at(float32(l.symptoms.jolt), 0)
	p = p.at(0, bob)

	facing := float32(1)
	if l.vx < -0.05 || (math.Abs(l.vx) <= 0.05 && l.aimX < 0) {
		facing = -1
	}
	nT := 4
	if arch == ArchetypeDart {
		nT = 3
	}
	if heavy {
		nT++
	}
	wd := float32(1)
	if light {
		wd = 0.8
	}

	if arch == ArchetypeTurret {
		// Legs planted: knee out, foot down, little toe curl.
		for i := 0; i < nT; i++ {
			f := float32(i)/float32(nT-1)*2 - 1
			bx := f * rx * 0.85
			side := f
			if heavy && i == nT/2 {
				side = 0
			}
			kx, ky := bx+side*4.5, ry*0.4+3.5
			fx, fy := bx+side*6.5, ry+5
			if e > 0 {
				fx += side * e * 14
				fy += e * 6
			}
			p.line(bx, ry*0.4, kx, ky, 2*wd, fade(pal.metal, alpha))
			p.line(kx, ky, fx, fy, 1.7*wd, fade(pal.metal, alpha))
			p.line(fx-2, fy, fx+2, fy, 1.5, fade(pal.dark, alpha))
		}
	} else {
		tx, ty := float32(-l.vx*3), float32(1-l.vy*0.6)
		if l.terminating {
			tx, ty = 0, 1
		}
		if d := float32(math.Hypot(float64(tx), float64(ty))); d > 0 {
			tx, ty = tx/d, ty/d
		}
		segs, seglen, amp := 5, float32(2.6), float32(0.35)
		switch arch {
		case ArchetypeSentinel:
			segs = 4
		case ArchetypeDart:
			segs, seglen, amp = 9, 3.2, 0.25
		}
		for i := 0; i < nT; i++ {
			f := float32(i)/float32(nT-1)*2 - 1
			base := f * rx * 0.6
			px, py := base, ry*0.55
			if arch == ArchetypeDart {
				px, py = -float32(facing)*rx*0.5, ry*0.2+float32(i)*2.2
			}
			lastx, lasty := px, py
			for k := 1; k <= segs; k++ {
				wave := float32(math.Sin(l.tick*0.12-float64(k)*0.9+float64(i)*1.3+l.seed)) * float32(k) * amp
				if l.terminating {
					wave = 0
				}
				px += tx*seglen - ty*wave
				py += ty*seglen + tx*wave
				if e > 0 {
					px += (base / 5) * e * float32(k) * 1.5
					py += e * float32(k)
				}
				w := (1.8 - float32(k)*(1.5/float32(segs+1))) * wd
				p.line(lastx, lasty, px, py, w, fade(pal.metal, alpha))
				lastx, lasty = px, py
			}
		}
	}

	split := e * 10
	top, bot := p.at(-split*0.4, -split), p.at(split*0.4, split*0.5)

	if arch == ArchetypeSentinel {
		rw := float32(2)
		if heavy {
			rw = 3
		}
		if light {
			rw = 1.3
		}
		bot.ringArc(16, 4.5, math.Pi, 2*math.Pi, rw, fade(pal.dark, alpha))
	}

	bot.ellipse(0, 0, rx, ry, fade(pal.dark, alpha))
	top.domeE(0, 0, rx, ry, fade(pal.hull, alpha))
	top.domeE(-rx*0.28, -ry*0.28, rx*0.5, ry*0.5, fade(pal.light, alpha*0.45))
	trimW := float32(1)
	if light {
		trimW = 0.6
	}
	top.line(-rx, 0, rx, 0, trimW, fade(pal.dark, alpha))

	switch arch {
	case ArchetypeProbe:
		top.line(0, -ry, 0, -ry-2.5, 1, fade(pal.metal, alpha))
		top.circle(0, -ry-3, 1, fade(pal.light, alpha))
	case ArchetypeTurret:
		// Flat plated skirt.
		bot.poly(fade(pal.metal, alpha), -rx-2, 2, rx+2, 2, rx, ry*0.9, -rx, ry*0.9)
		bot.line(-rx-2, 2, rx+2, 2, 1, fade(pal.dark, alpha))
		for _, sx := range []float32{-6, 0, 6} {
			bot.line(sx, 2, sx, ry*0.9, 0.7, fade(pal.dark, alpha))
		}
	case ArchetypeSentinel:
		// ring front half in front of the orb
		rc := mix(pal.light, rgba(255, 255, 255), l.charge)
		if l.terminating {
			rc = gray(pal.light, 0.5)
		}
		rw := float32(2)
		if heavy {
			rw = 3
		}
		if light {
			rw = 1.3
		}
		top.ringArc(16, 4.5, 0, math.Pi, rw, fade(rc, alpha))
		top.ringArc(16, 4.5, math.Pi, 2*math.Pi, rw*0.6, fade(rc, alpha*0.5))
		for k := 0; k < 3; k++ {
			a := l.tick*0.05 + float64(k)*2.09
			top.circle(16*float32(math.Cos(a)), 4.5*float32(math.Sin(a)), 1.3, fade(pal.eye, alpha))
		}
	case ArchetypeDart:
		f := facing
		top.poly(fade(pal.hull, alpha), f*3, -2, f*15, 1, f*3, 3)
		top.poly(fade(pal.light, alpha), f*1, -ry+1, -f*11, -ry-8, -f*5, -ry+4)
		top.poly(fade(pal.dark, alpha), -f*1, ry-2, -f*9, ry+4, -f*4, ry-4)
	}

	if heavy {
		for _, by := range []float32{-0.4, 0.45} {
			hw := rx * float32(math.Sqrt(1-float64(by*by)))
			if by < 0 {
				top.line(-hw, by*ry, hw, by*ry, 1.8, fade(pal.dark, alpha))
				top.line(-hw, by*ry-1.4, hw, by*ry-1.4, 0.6, fade(pal.light, alpha*0.7))
			} else {
				bot.line(-hw, by*ry, hw, by*ry, 1.8, fade(pal.metal, alpha))
			}
		}
		for _, sx := range []float32{-0.7, 0.7} {
			top.circle(sx*rx, -ry*0.18, 0.8, fade(pal.light, alpha))
		}
	}

	drawSymptoms(top, pal, rx, ry, l.symptoms)

	// Eye.
	if e == 0 {
		eyeC := mix(pal.dark, mix(pal.eye, rgba(255, 255, 255), l.charge), l.symptoms.eyeLevel)
		eyeShut := l.terminating || l.symptoms.eyeShut
		if arch == ArchetypeTurret {
			ang := math.Atan2(math.Min(l.aimY, -0.35), l.aimX)
			if l.terminating {
				ang = math.Pi * 0.9
			}
			bx0, by0 := float32(0), -ry+1
			bx1, by1 := bx0+float32(math.Cos(ang))*8, by0+float32(math.Sin(ang))*8
			p.line(bx0, by0, bx1, by1, 4.6, pal.dark)
			p.line(bx0, by0, bx1, by1, 2.6, pal.metal)
			if l.charge > 0 && !l.terminating {
				p.circle(bx1, by1, 3.5+float32(l.charge)*5, fade(eyeC, 0.35))
			}
			p.circle(bx1, by1, 3.3, pal.dark)
			switch {
			case eyeShut:
				p.line(bx1-2, by1, bx1+2, by1, 0.8, pal.eye)
			case l.damage:
				p.circle(bx1, by1, 2.9, rgba(255, 255, 255))
			default:
				p.circle(bx1, by1, 2.4, eyeC)
				p.circle(bx1-0.6, by1-0.6, 0.6, rgba(255, 255, 255))
			}
		} else {
			er := float32(3.6)
			if arch == ArchetypeDart {
				er = 2.6
			}
			ex, ey := float32(l.aimX)*rx*0.35, float32(l.aimY)*ry*0.25-1
			if arch == ArchetypeDart {
				ex, ey = facing*2.2, -2
			}
			if l.charge > 0 && !l.terminating {
				p.circle(ex, ey, er+1+float32(l.charge)*4, fade(eyeC, 0.3))
			}
			p.circle(ex, ey, er, pal.dark)
			switch {
			case eyeShut:
				p.line(ex-er*0.7, ey, ex+er*0.7, ey, 0.8, pal.eye)
			case l.damage:
				p.circle(ex, ey, er-0.4, rgba(255, 255, 255))
			default:
				p.circle(ex, ey, er-1, eyeC)
				p.circle(ex+float32(l.aimX)*0.8, ey+float32(l.aimY)*0.8, er*0.3, rgba(20, 30, 50))
				p.circle(ex-0.9, ey-0.9, 0.6, rgba(255, 255, 255))
			}
		}
	} else {
		p.circle(0, -1-e*22, 2.6*(1-e)+0.4, fade(pal.eye, alpha))
		p.ring(0, 0, 5+e*14, 1.2, fade(droidPalettes[arch].eye, alpha))
	}
}
