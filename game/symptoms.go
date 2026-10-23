package game

import "math"

// rebootBlinkTicks is how long the reboot blink lasts after a restart.
const rebootBlinkTicks = 20

// patchPlates is how many welded patch plates a restart count earns: none, one
// or two for 0, 1-4 or 5+ restarts.
func patchPlates(restarts int) int {
	switch {
	case restarts >= 5:
		return 2
	case restarts >= 1:
		return 1
	default:
		return 0
	}
}

// symptomNoise is a deterministic hash of (seed, n) in [0, 1).
func symptomNoise(seed float64, n uint64) float64 {
	x := math.Sin(seed*12.9898+float64(n)*78.233) * 43758.5453
	return x - math.Floor(x)
}

// eyeFlicker is the eye brightness of a not-ready droid: irregular dips and
// stutters in 0.15..1, changing every few ticks.
func eyeFlicker(tick uint64, seed float64) float64 {
	n := symptomNoise(seed, tick/3)
	if n < 0.45 {
		return 1
	}
	return 0.15 + n*0.5
}

// antennaSparks reports whether a not-ready droid's antenna sparks this tick:
// a short crackle now and then.
func antennaSparks(tick uint64, seed float64) bool {
	return symptomNoise(seed+7, tick/6) < 0.25
}

// rebootBlink poses the reboot blink from the ticks it has left: the droid
// jolts first, then the eye is shut, then it reopens.
func rebootBlink(timer int) (eyeShut bool, jolt float64) {
	if timer <= 0 {
		return false, 0
	}
	progress := float64(rebootBlinkTicks-timer) / rebootBlinkTicks
	switch {
	case progress < 0.2:
		return false, math.Sin(float64(timer)*3) * 1.5
	case progress < 0.75:
		return true, 0
	default:
		return false, 0
	}
}

// SetSymptoms records new Symptoms. A restart increase starts the reboot blink.
func (e *Enemy) SetSymptoms(s Symptoms) {
	if s.Restarts > e.Symptoms.Restarts {
		e.rebootTimer = rebootBlinkTicks
	}
	e.Symptoms = s
}

// tickSymptoms advances the Symptom animations of every enemy by one tick.
func (s *Scene) tickSymptoms() {
	for _, e := range s.enemies {
		if e.rebootTimer > 0 {
			e.rebootTimer--
		}
	}
}

// updateSymptoms applies a Symptoms-changed event to the live enemy and to any
// pending practice respawn of the same Pod. A pending respawn never blinks.
func (s *Scene) updateSymptoms(podName, namespace string, symptoms Symptoms) {
	for _, e := range s.enemies {
		if e.PodName == podName && e.Namespace == namespace {
			e.SetSymptoms(symptoms)
		}
	}
	for i := range s.respawns {
		if r := &s.respawns[i]; r.podName == podName && r.namespace == namespace {
			r.symptoms = symptoms
		}
	}
}

// symptomLook is how Symptoms pose a droid on one draw tick.
type symptomLook struct {
	eyeLevel float64 // eye brightness, 1 when healthy
	eyeShut  bool
	sparks   bool
	plates   int
	jolt     float64
}

// symptomLookOf poses an enemy's Symptoms. Terminating and Exploding visuals
// replace Symptoms entirely.
func symptomLookOf(e *Enemy, tick uint64) symptomLook {
	if e.State != EnemyActive {
		return symptomLook{eyeLevel: 1}
	}
	return symptomPose(e.Symptoms, e.rebootTimer, tick, e.seed)
}

// drawSymptoms draws the hull marks: the antenna spark and the patch plates.
// top is the pen of the upper hull, rx and ry the hull radii.
func drawSymptoms(top pen, pal droidPalette, rx, ry float32, l symptomLook) {
	for i := 0; i < l.plates; i++ {
		x := -rx*0.5 + float32(i)*rx*0.75
		y := ry * (0.15 - 0.45*float32(i))
		top.poly(pal.dark, x-2.6, y-2, x+2.6, y-2.4, x+2.4, y+2, x-2.4, y+2.4)
		top.poly(pal.light, x-1.9, y-1.4, x+1.9, y-1.7, x+1.7, y+1.4, x-1.7, y+1.7)
		top.circle(x-1.1, y-0.7, 0.4, pal.dark)
		top.circle(x+1.1, y+0.7, 0.4, pal.dark)
	}
	if l.sparks {
		spark := rgba(255, 240, 150)
		top.line(0, -ry, 0, -ry-4, 0.8, pal.metal)
		top.circle(0, -ry-4.5, 1.6, fade(spark, 0.45))
		top.line(-2.5, -ry-6.5, 2.5, -ry-3, 0.8, spark)
		top.line(2.5, -ry-6.5, -2, -ry-3.5, 0.6, spark)
	}
}
