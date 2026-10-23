package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestDrawDroidAllVariantsDoesNotPanic(t *testing.T) {
	img := ebiten.NewImage(64, 64)
	for a := ArchetypeProbe; a <= ArchetypeDart; a++ {
		for ti := TierLight; ti <= TierHeavy; ti++ {
			for _, l := range []droidLook{{aimX: 1}, {damage: true}, {contact: true}, {terminating: true}, {explode: 0.5}, {charge: 1, vx: -1},
				{symptoms: symptomLook{eyeLevel: 0.3, sparks: true, plates: 2, jolt: 1}}, {symptoms: symptomLook{eyeLevel: 1, eyeShut: true, plates: 1}}} {
				drawDroid(pen{dst: img, cx: 32, cy: 32, s: 1}, a, ti, l)
			}
		}
	}
}
