package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	guideMargin = 24
	guideRight  = screenWidth - guideMargin
	// guideWrapCols keeps a text block inside the margins at the 6 px debug font.
	guideWrapCols = 120
)

const (
	guideTiersText = "A Pod's Tier comes from its CPU/memory requests relative to the other Pods in the namespace. No requests = Light."
	guideBoltsText = "Bolts are enemy shots. Each costs 1 Health. Dodge them: they can't be shot down and they pass through enemies."
)

var (
	guideColText = rgba(235, 240, 250)
	guideColDim  = rgba(170, 180, 205)
	guideColYou  = rgba(0, 255, 255)
)

var (
	guideArchNames = [4]string{"Probe", "Turret", "Sentinel", "Dart"}
	guideTierNames = [3]string{"Light", "Standard", "Heavy"}
)

func archLight(a Archetype) color.NRGBA { return droidPalettes[a].light }

// guideTierSize is the droid scale a Tier draws at.
func guideTierSize(t Tier) float32 { return float32(t.stats().size) }

// wrapText breaks s into lines of at most cols characters.
func wrapText(s string, cols int) []string {
	var out []string
	line := ""
	for w := range strings.FieldsSeq(s) {
		switch {
		case line == "":
			line = w
		case len(line)+1+len(w) > cols:
			out = append(out, line)
			line = w
		default:
			line += " " + w
		}
	}
	return append(out, line)
}

// text draws s tinted and scaled. The debug font is fixed-size, so each string
// is rendered once to a sprite and scaled from there.
func (f *fieldGuide) label(dst *ebiten.Image, s string, x, y, scale float64, c color.Color) {
	img, ok := f.text[s]
	if !ok {
		img = ebiten.NewImage(len(s)*6+2, 16)
		ebitenutil.DebugPrint(img, s)
		f.text[s] = img
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	dst.DrawImage(img, op)
}

// centred draws dim or bright text centred on cx.
func (f *fieldGuide) centred(dst *ebiten.Image, s string, cx, y float64, c color.Color) {
	f.label(dst, s, cx-float64(len(s))*3, y, 1, c)
}

// block draws wrapped text and returns the y below it.
func (f *fieldGuide) block(dst *ebiten.Image, s string, x, y float64, c color.Color) float64 {
	for _, l := range wrapText(s, guideWrapCols) {
		f.label(dst, l, x, y, 1, c)
		y += 20
	}
	return y
}

func guidePanel(dst *ebiten.Image, x, y, w, h float32, border color.Color) *ebiten.Image {
	vector.FillRect(dst, x, y, w, h, color.NRGBA{R: 12, G: 15, B: 30, A: 255}, false)
	vector.StrokeRect(dst, x, y, w, h, 1, border, false)
	return dst.SubImage(image.Rect(int(x)+1, int(y)+1, int(x+w)-1, int(y+h)-1)).(*ebiten.Image)
}

// guideTurretCharge is the Telegraph a Turret of tier t shows tick updates
// after arming with no first-volley offset, so the portrait glows on the real cycle.
func guideTurretCharge(t Tier, tick int) float64 {
	pat := turretPatternOf(t)
	since := tick % pat.interval
	charge := windUp(pat.interval-since, turretRingTelegraph)
	if pat.aimed {
		charge = math.Max(charge, windUp(turretAimedDelay-since, turretAimedTelegraph))
	}
	return charge
}

// guideDroid draws a droid of the given archetype and tier at (cx, cy). Its
// telegraph loops so a wind-up is visible.
func guideDroid(dst *ebiten.Image, cx, cy, scale float32, a Archetype, t Tier, sym symptomLook, tick int, seed float64) {
	ft := float64(tick)
	l := droidLook{tick: ft, seed: seed, aimX: math.Cos(ft * 0.03), aimY: math.Sin(ft * 0.03), vx: math.Sin(ft*0.02) * 0.5, symptoms: sym}
	switch a {
	case ArchetypeTurret:
		l.charge = guideTurretCharge(t, tick)
	case ArchetypeSentinel:
		l.charge = math.Max(0, float64(tick%150-130)/20)
	case ArchetypeDart:
		l.charge = math.Max(0, float64(tick%100-70)/30)
	}
	drawDroid(pen{dst: dst, cx: cx, cy: cy, s: guideTierSize(t) * scale}, a, t, l)
}

func (f *fieldGuide) draw(screen *ebiten.Image) {
	vector.FillRect(screen, 0, 0, screenWidth, screenHeight, color.NRGBA{R: 6, G: 8, B: 18, A: 255}, false)
	f.label(screen, "FIELD GUIDE - "+guidePageTitles[f.page], guideMargin, 12, 2, guideColText)
	pg := fmt.Sprintf("page %d/%d", f.page+1, guidePageCount)
	f.label(screen, pg, guideRight-float64(len(pg)*6), 20, 1, guideColDim)
	vector.StrokeLine(screen, guideMargin, 50, guideRight, 50, 1, colorWeaponPanelBorder, false)

	keys := ""
	switch {
	case f.page < guidePageTiers:
		f.drawArchetypePage(screen)
		keys = "[1/2/3] demo tier: " + guideTierNames[f.tier]
	case f.page == guidePageTiers:
		f.drawTiersPage(screen)
	case f.page == guidePageSymptoms:
		f.drawSymptomsPage(screen)
		keys = fmt.Sprintf("[N] not ready: %s  [P] restarts: %d  [R] reboot blink  [1/2/3] tier: %s  [Up/Down] reference: %s",
			map[bool]string{true: "on", false: "off"}[f.notReady], f.restarts, guideTierNames[f.tier], guideArchNames[f.row])
	default:
		f.drawPodsPage(screen)
		keys = "[Up/Down] choose a Pod"
	}
	vector.FillRect(screen, 0, 560, screenWidth, 40, color.NRGBA{R: 20, G: 24, B: 44, A: 255}, false)
	f.label(screen, keys, guideMargin, 566, 1, guideColText)
	f.label(screen, "[Left/Right] page  [Esc/Backspace] back to mode select", guideMargin, 582, 1, guideColDim)
}

func (f *fieldGuide) drawArchetypePage(screen *ebiten.Image) {
	a := Archetype(f.page)
	t := guideArchetypes[a]
	f.label(screen, "Source Pod kind: "+t.source, guideMargin, 64, 1, archLight(a))
	y := f.block(screen, t.does, guideMargin, 90, guideColText)
	if t.note != "" {
		f.block(screen, t.note, guideMargin, y, guideColDim)
	}

	clip := guidePanel(screen, guideMargin, 164, 752, 152, droidPalettes[a].hull)
	for ti := TierLight; ti <= TierHeavy; ti++ {
		cx := float32(guideMargin + 125 + int(ti)*250)
		f.centred(screen, fmt.Sprintf("%s %.2fx", guideTierNames[ti], ti.stats().size), float64(cx), 172, guideColText)
		guideDroid(clip, cx, 248, 2.4, a, ti, symptomLook{eyeLevel: 1}, f.tick, float64(ti))
	}
	f.label(screen, "Live demo vs a dummy player (cyan). Tier: "+guideTierNames[f.tier], guideMargin, 326, 1, guideColDim)
	if f.demo != nil {
		f.demo.draw(f, screen)
	}
}

// draw renders the demo panel: the real enemy and Bolts, and the dummy player.
func (d *guideDemo) draw(f *fieldGuide, screen *ebiten.Image) {
	clip := guidePanel(screen, demoX, demoY, demoW, demoH, colorWeaponPanelBorder)
	cx, cy, k := demoX+demoW/2, demoY+demoH/2, d.zoom()
	toPanel := func(x, y float64) (float32, float32) {
		return float32(cx + (x-cx)*k), float32(cy + (y-cy)*k)
	}
	px, py := toPanel(d.player.X, d.player.Y)

	e := d.enemy()
	ex, ey := toPanel(e.X, e.Y)
	look := droidLookOf(e, d.tick, d.player.X, d.player.Y)
	drawDroid(pen{dst: clip, cx: ex, cy: ey, s: float32(e.Scale()) * 1.3}, d.arch, d.tier, look)
	for _, b := range d.scene.Bolts() {
		bx, by := toPanel(b.X, b.Y)
		drawBoltAt(clip, bx, by, 1.5, b.Archetype)
	}
	for _, burst := range d.scene.Bursts() {
		bx, by := toPanel(burst.X, burst.Y)
		t := float32(burst.Age) / burstTicks
		vector.StrokeCircle(clip, bx, by, 3+10*t, 2, withAlpha(archLight(burst.Archetype), uint8(255*(1-t))), true)
	}

	ship := [3][2]float32{{px + 12, py}, {px - 9, py - 8}, {px - 9, py + 8}}
	for i := range ship {
		n := ship[(i+1)%3]
		vector.StrokeLine(clip, ship[i][0], ship[i][1], n[0], n[1], 2, guideColYou, true)
	}
	if d.hit > 0 {
		vector.StrokeCircle(clip, px, py, 18, 2, rgba(255, 240, 160), true)
		ebitenutil.DebugPrintAt(screen, "-1 Health", int(px)-26, int(py)-36-(demoHitTicks-d.hit)/3)
	}
	f.label(screen, d.caption(), demoX+8, demoY+demoH-20, 1, guideColDim)
}

func (d *guideDemo) caption() string {
	switch d.arch {
	case ArchetypeProbe:
		return "Homing: trails the player, never stops."
	case ArchetypeTurret:
		return [...]string{"Light: 4-way fixed ring", "Standard: 6-way rotating ring", "Heavy: rotating ring + aimed Bolt"}[d.tier] + " (eye glows first)"
	case ArchetypeSentinel:
		return "Orbit, ring flashes, dash in, back to orbit. " + [...]string{"Light: wide, rare", "Standard", "Heavy: tight, frequent"}[d.tier]
	}
	return "Wind-up (eye glows), straight charge, overshoot. " + [...]string{"Light: slow/short", "Standard", "Heavy: fast/far"}[d.tier]
}

// drawBoltAt draws one Bolt, scaled from its in-game size.
func drawBoltAt(dst *ebiten.Image, x, y, scale float32, a Archetype) {
	pal := droidPalettes[a]
	vector.FillCircle(dst, x, y, boltSize*scale, withAlpha(pal.hull, 70), false)
	vector.FillCircle(dst, x, y, boltSize/2*scale, pal.light, false)
}

func (f *fieldGuide) drawTiersPage(screen *ebiten.Image) {
	f.block(screen, guideTiersText, guideMargin, 64, guideColText)
	clip := guidePanel(screen, guideMargin, 108, 752, 164, colorWeaponPanelBorder)
	caps := [3][3]string{
		{"Light 0.85x", "thinner hull", "2 Hits, faster"},
		{"Standard 1x", "baseline", "3 Hits"},
		{"Heavy 1.3x", "armour, extra tentacle,", "5 Hits, slower, 2 dmg"},
	}
	for ti := TierLight; ti <= TierHeavy; ti++ {
		cx := float32(guideMargin + 125 + int(ti)*250)
		guideDroid(clip, cx, 168, 2.2, ArchetypeProbe, ti, symptomLook{eyeLevel: 1}, f.tick, float64(ti))
		f.centred(screen, caps[ti][0], float64(cx), 222, guideColText)
		f.centred(screen, caps[ti][1], float64(cx), 238, guideColDim)
		f.centred(screen, caps[ti][2], float64(cx), 254, guideColDim)
	}

	f.block(screen, guideBoltsText, guideMargin, 288, guideColText)
	clip = guidePanel(screen, guideMargin, 330, 752, 218, colorWeaponPanelBorder)
	const cy = 440
	guideDroid(clip, 400, cy, 2.6, ArchetypeTurret, TierStandard, symptomLook{eyeLevel: 1}, f.tick, 1)
	run := float32(f.tick % 220)
	for a := range Archetype(4) {
		drawBoltAt(clip, 60+run*3.2-float32(a)*40, cy-60+float32(a)*30, 1.5, a)
	}
	vector.StrokeCircle(screen, 700, cy, 22, 2, guideColYou, true)
	f.centred(screen, "you", 700, cy+30, guideColText)
	f.label(screen, "Bolts pass through the droid and cost you 1 Health", 250, 520, 1, guideColDim)
}

func (f *fieldGuide) drawSymptomsPage(screen *ebiten.Image) {
	f.label(screen, "Symptoms show a Pod's live health. They never change behaviour.", guideMargin, 64, 1, guideColText)
	clip := guidePanel(screen, guideMargin, 90, 752, 156, colorWeaponPanelBorder)
	f.label(screen, "Live (toggle with the keys below)", 32, 96, 1, guideColDim)
	live := Symptoms{Ready: !f.notReady, Restarts: f.restarts}
	for a := ArchetypeProbe; a <= ArchetypeDart; a++ {
		cx := float32(guideMargin + 94 + int(a)*188)
		sym := symptomPose(live, f.rebootTimer, uint64(f.tick), float64(a))
		guideDroid(clip, cx, 170, 2, a, f.tier, sym, f.tick, float64(a))
		f.centred(screen, guideArchNames[a], float64(cx), 228, guideColText)
	}

	clip = guidePanel(screen, guideMargin, 260, 752, 288, colorWeaponPanelBorder)
	ref := Archetype(f.row)
	f.label(screen, "Reference: "+guideArchNames[ref], 32, 266, 1, guideColDim)
	loopTimer := max(0, rebootBlinkTicks-f.tick%90)
	refs := [...]struct {
		l1, l2 string
		s      Symptoms
		timer  int
	}{
		{"0 restarts", "no patch", Symptoms{Ready: true}, 0},
		{"1-4 restarts", "one patch", Symptoms{Ready: true, Restarts: 1}, 0},
		{"5+ restarts", "two patches", Symptoms{Ready: true, Restarts: 5}, 0},
		{"Not ready", "eye flickers", Symptoms{}, 0},
		{"Restart in play", "reboot blink", Symptoms{Ready: true}, loopTimer},
	}
	for i, r := range refs {
		cx := float32(guideMargin + 75 + i*150)
		guideDroid(clip, cx, 380, 2.8, ref, f.tier, symptomPose(r.s, r.timer, uint64(f.tick), float64(i)), f.tick, float64(i))
		f.centred(screen, r.l1, float64(cx), 496, guideColText)
		f.centred(screen, r.l2, float64(cx), 514, guideColDim)
	}
}

func (f *fieldGuide) drawPodsPage(screen *ebiten.Image) {
	f.label(screen, "Example Pods - illustrative, not your cluster", guideMargin, 64, 1, guideColText)
	const hy = 112
	for _, c := range [...]struct {
		x float64
		s string
	}{{40, "NAME"}, {176, "OWNER"}, {276, "READY"}, {336, "RESTARTS"}, {416, "REQUESTS"}, {520, "DROID"}, {580, "ARCHETYPE / TIER"}} {
		f.label(screen, c.s, c.x, hy, 1, guideColDim)
	}
	vector.StrokeLine(screen, guideMargin, hy+20, guideRight, hy+20, 1, colorWeaponPanelBorder, false)
	const rowH = 36
	rows := screen.SubImage(image.Rect(guideMargin, hy+22, guideRight, hy+22+len(guidePods)*rowH)).(*ebiten.Image)
	for i, p := range guidePods {
		y := hy + 26 + i*rowH
		if i == f.row {
			vector.FillRect(screen, guideMargin, float32(y), 752, 34, color.NRGBA{R: 30, G: 40, B: 80, A: 255}, false)
			ebitenutil.DebugPrintAt(screen, ">", 28, y+9)
		}
		ready := "1/1"
		if !p.ready {
			ready = "0/1"
		}
		for _, c := range [...]struct {
			x int
			s string
		}{{40, p.name}, {176, p.owner}, {276, ready}, {336, fmt.Sprint(p.restarts)}, {416, p.requests}} {
			ebitenutil.DebugPrintAt(screen, c.s, c.x, y+9)
		}
		if !p.mapped {
			f.label(screen, "no controller: never an enemy", 520, float64(y+9), 1, guideColDim)
			continue
		}
		sym := symptomPose(Symptoms{Ready: p.ready, Restarts: p.restarts}, 0, uint64(f.tick), float64(i))
		guideDroid(rows, 536, float32(y)+18, 0.9, p.arch, p.tier, sym, f.tick, float64(i))
		f.label(screen, guideArchNames[p.arch]+" "+guideTierNames[p.tier], 580, float64(y+9), 1, archLight(p.arch))
	}

	p := guidePods[f.row]
	const py = 402
	clip := guidePanel(screen, guideMargin, py, 752, 146, colorWeaponPanelBorder)
	if !p.mapped {
		for i, l := range wrapText(p.name+" has no controller, so it never appears as an enemy.", 60) {
			f.label(screen, l, 200, float64(py+50+i*20), 1, guideColText)
		}
		return
	}
	sym := symptomPose(Symptoms{Ready: p.ready, Restarts: p.restarts}, 0, uint64(f.tick), 3)
	guideDroid(clip, 100, py+78, 3.2, p.arch, p.tier, sym, f.tick, 3)
	f.label(screen, fmt.Sprintf("%s -> %s %s: %s, %s requests", p.name, guideTierNames[p.tier], guideArchNames[p.arch], p.owner, p.requests), 200, py+36, 1, archLight(p.arch))
	ready := "ready, no flicker"
	if !p.ready {
		ready = "not ready, eye flickers"
	}
	f.label(screen, fmt.Sprintf("Symptoms: %s; %d restarts -> %d patch(es)", ready, p.restarts, patchPlates(p.restarts)), 200, py+66, 1, guideColText)
}
