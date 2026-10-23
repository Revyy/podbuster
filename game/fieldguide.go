package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// guidePageCount is the number of Field guide pages.
const guidePageCount = 7

// Field guide pages, in reading order. The first four are the archetypes and
// share their numbering with Archetype.
const (
	guidePageTiers = iota + 4
	guidePageSymptoms
	guidePagePods
)

var guidePageTitles = [guidePageCount]string{
	"Probe", "Turret", "Sentinel", "Dart", "Tiers & Bolts", "Symptoms", "How Pods become enemies",
}

// guideDemo geometry: the panel the live demo plays in.
const (
	demoX, demoY, demoW, demoH = 24.0, 346.0, 752.0, 202.0
	demoHitTicks               = 40
	// demoSentinelZoom shrinks the demo world around the panel centre so a
	// Sentinel's real orbit fits the panel.
	demoSentinelZoom = 0.45
)

// guideArchetypeText is what each archetype page says. There is no look
// description: the player can see the droid.
type guideArchetypeText struct{ source, does, note string }

var guideArchetypes = [4]guideArchetypeText{
	{"ReplicaSet (Deployments) and any unrecognised controller",
		"Homes in on the player. The baseline enemy.", ""},
	{"StatefulSet",
		"Stationary and immovable. Fires a ring of Bolts; the eye glows before each volley. Heavier tiers add rotation and an aimed Bolt.", ""},
	{"DaemonSet",
		"Orbits the player at range, then dashes in after its ring flashes. Heavier tiers orbit tighter and dash more often.", ""},
	{"Job (covers CronJob)",
		"Fragile. Winds up, charges in a straight line and overshoots. Heavier tiers charge faster and further.",
		"Chaos: deleting a Job Pod may count against the Job's retry limit (backoffLimit)."},
}

// guidePod is one illustrative row of the last page's table.
type guidePod struct {
	name, owner string
	mapped      bool // false: no controller, never an enemy
	arch        Archetype
	tier        Tier
	ready       bool
	restarts    int
	requests    string
}

var guidePods = [...]guidePod{
	{"web-6d8f9-x7k2p", "ReplicaSet", true, ArchetypeProbe, TierStandard, true, 0, "100m"},
	{"db-0", "StatefulSet", true, ArchetypeTurret, TierStandard, true, 0, "250m"},
	{"db-1", "StatefulSet", true, ArchetypeTurret, TierHeavy, false, 6, "2000m"},
	{"logship-n4x8d", "DaemonSet", true, ArchetypeSentinel, TierLight, true, 1, "<none>"},
	{"backup-28911-z9", "Job", true, ArchetypeDart, TierLight, true, 0, "<none>"},
	{"rollout-ctl-0", "Rollout (?)", true, ArchetypeProbe, TierStandard, true, 5, "300m"},
	{"debug-shell", "<none>", false, 0, 0, true, 0, "<none>"},
}

// menuKeys is one tick of menu input, so navigation can be tested without a
// keyboard.
type menuKeys struct {
	up, down, left, right bool
	enter, esc, back      bool
	// tier is 1, 2 or 3 when that number key was pressed, else 0.
	tier int
	// notReady, restarts and reboot are the Symptoms page toggles.
	notReady, restarts, reboot bool
}

// fieldGuide is the state of the Field guide screen. It reads no cluster data.
type fieldGuide struct {
	page int
	tick int
	tier Tier
	row  int
	demo *guideDemo
	text map[string]*ebiten.Image
	// Symptoms page live row.
	notReady    bool
	restarts    int
	rebootTimer int
}

func newFieldGuide() *fieldGuide {
	return &fieldGuide{tier: TierStandard, text: map[string]*ebiten.Image{}}
}

// open returns the guide to its first page.
func (f *fieldGuide) open() {
	f.page, f.tick, f.row = 0, 0, 0
	f.demo = nil
}

// update advances the guide one tick and reports whether the player left it.
func (f *fieldGuide) update(in menuKeys) (leave bool) {
	if in.esc || in.back {
		return true
	}
	f.tick++
	switch {
	case in.right:
		f.page = (f.page + 1) % guidePageCount
		f.demo, f.row = nil, 0
	case in.left:
		f.page = (f.page + guidePageCount - 1) % guidePageCount
		f.demo, f.row = nil, 0
	}
	if in.tier >= 1 && in.tier <= 3 {
		f.tier = Tier(in.tier - 1)
		f.demo = nil
	}
	if f.rebootTimer > 0 {
		f.rebootTimer--
	}

	switch {
	case f.page < guidePageTiers:
		if f.demo == nil {
			f.demo = newGuideDemo(Archetype(f.page), f.tier)
		}
		f.demo.step()
	case f.page == guidePageSymptoms:
		if in.up {
			f.row = (f.row + 3) % 4
		}
		if in.down {
			f.row = (f.row + 1) % 4
		}
		if in.notReady {
			f.notReady = !f.notReady
		}
		if in.restarts {
			f.restarts = [...]int{1, 5, 0}[patchPlates(f.restarts)]
		}
		if in.reboot {
			f.rebootTimer = rebootBlinkTicks
		}
	case f.page == guidePagePods:
		if in.up {
			f.row = (f.row + len(guidePods) - 1) % len(guidePods)
		}
		if in.down {
			f.row = (f.row + 1) % len(guidePods)
		}
	}
	return false
}

// symptomPose poses Symptoms and a reboot blink on one draw tick. Enemies and
// the Field guide share it.
func symptomPose(s Symptoms, rebootTimer int, tick uint64, seed float64) symptomLook {
	l := symptomLook{eyeLevel: 1, plates: patchPlates(s.Restarts)}
	if !s.Ready {
		l.eyeLevel = eyeFlicker(tick, seed)
		l.sparks = antennaSparks(tick, seed)
	}
	l.eyeShut, l.jolt = rebootBlink(rebootTimer)
	return l
}

// guideDemo plays one real enemy against a dummy player in the demo panel.
// It reuses Scene, so the demo is the shipped behaviour, not a copy of it.
type guideDemo struct {
	scene  *Scene
	player *Player
	arch   Archetype
	tier   Tier
	tick   uint64
	hit    int
}

func newGuideDemo(arch Archetype, tier Tier) *guideDemo {
	d := &guideDemo{scene: NewScene(nil, nil), player: NewPlayer(0, 0), arch: arch, tier: tier}
	cx, cy := demoX+demoW/2, demoY+demoH/2
	x, y := demoX+demoW*0.3, cy
	e := NewEnemyOf("demo", "guide", arch, tier, x, y)
	e.Symptoms.Ready = true
	switch arch {
	case ArchetypeTurret:
		e.X, e.Y = cx, cy
	case ArchetypeSentinel:
		e.initSentinel(d.scene.rng)
	}
	d.scene.enemies = []*Enemy{e}
	d.placePlayer()
	return d
}

// enemy is the single demo enemy.
func (d *guideDemo) enemy() *Enemy { return d.scene.enemies[0] }

func (d *guideDemo) zoom() float64 {
	if d.arch == ArchetypeSentinel {
		return demoSentinelZoom
	}
	return 1
}

// placePlayer moves the dummy player along a gentle path.
func (d *guideDemo) placePlayer() {
	t := float64(d.tick)
	cx, cy := demoX+demoW/2, demoY+demoH/2
	switch d.arch {
	case ArchetypeProbe:
		d.player.X, d.player.Y = cx+math.Cos(t*0.02)*220, cy+math.Sin(t*0.03)*80
	case ArchetypeTurret:
		d.player.X, d.player.Y = cx+math.Cos(t*0.012)*260, cy+math.Sin(t*0.02)*90
	case ArchetypeSentinel:
		d.player.X, d.player.Y = cx+math.Cos(t*0.015)*40, cy+math.Sin(t*0.02)*14
	default:
		d.player.X, d.player.Y = cx+math.Cos(t*0.01)*60, cy+math.Sin(t*0.014)*30
	}
}

// step advances the demo one tick. A hit on the dummy is shown, then undone.
func (d *guideDemo) step() {
	d.tick++
	d.placePlayer()
	if d.player.InvulnerableTimer > 0 {
		d.player.InvulnerableTimer--
	}
	d.scene.Update(d.player)
	if d.player.Health < PlayerMaxHealth {
		d.player.Health = PlayerMaxHealth
		d.hit = demoHitTicks
	}
	if d.hit > 0 {
		d.hit--
	}
	d.placePlayer()
}
