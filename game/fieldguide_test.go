package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func goToMode(t *testing.T) (*Game, chan SessionEvent) {
	t.Helper()
	g, ev := newTestGame(t)
	g.state = StateModeSelect
	return g, ev
}

func TestModeSelect_CursorReachesFieldGuideAndEnterOpensIt(t *testing.T) {
	g, sessionEvents := goToMode(t)

	g.handleModeSelect(menuKeys{down: true})
	g.handleModeSelect(menuKeys{down: true})
	if g.modeCursor != entryGuide {
		t.Fatalf("expected the cursor on the Field guide, got %v", g.modeCursor)
	}
	g.handleModeSelect(menuKeys{enter: true})

	if g.State() != StateFieldGuide {
		t.Fatalf("expected StateFieldGuide, got %v", g.State())
	}
	if len(sessionEvents) != 0 {
		t.Fatal("opening the Field guide must not start a session")
	}
}

func TestFieldGuide_BackReturnsToModeSelectOnItsEntry(t *testing.T) {
	for _, keys := range []menuKeys{{esc: true}, {back: true}} {
		g, _ := goToMode(t)
		g.modeCursor = entryGuide
		g.handleModeSelect(menuKeys{enter: true})

		g.handleFieldGuide(keys)

		if g.State() != StateModeSelect || g.modeCursor != entryGuide {
			t.Fatalf("expected mode select on the guide entry, got %v / %v", g.State(), g.modeCursor)
		}
	}
}

func TestFieldGuide_LeftRightTurnPagesAndWrap(t *testing.T) {
	g, _ := goToMode(t)
	g.modeCursor = entryGuide
	g.handleModeSelect(menuKeys{enter: true})

	g.handleFieldGuide(menuKeys{right: true})
	if g.guide.page != 1 {
		t.Fatalf("expected page 1, got %d", g.guide.page)
	}
	g.handleFieldGuide(menuKeys{left: true})
	g.handleFieldGuide(menuKeys{left: true})
	if g.guide.page != guidePageCount-1 {
		t.Fatalf("expected wrap to the last page, got %d", g.guide.page)
	}
	g.handleFieldGuide(menuKeys{right: true})
	if g.guide.page != 0 {
		t.Fatalf("expected wrap to the first page, got %d", g.guide.page)
	}
}

func TestFieldGuide_OpensOnFirstPage(t *testing.T) {
	g, _ := goToMode(t)
	g.modeCursor = entryGuide
	g.handleModeSelect(menuKeys{enter: true})
	g.handleFieldGuide(menuKeys{right: true})
	g.handleFieldGuide(menuKeys{esc: true})
	g.handleModeSelect(menuKeys{enter: true})
	if g.guide.page != 0 {
		t.Fatalf("expected the guide to reopen on page 0, got %d", g.guide.page)
	}
}

func TestModeSelect_ChaosStillNeedsSecondConfirm(t *testing.T) {
	g, sessionEvents := goToMode(t)
	g.handleModeSelect(menuKeys{down: true})
	g.handleModeSelect(menuKeys{enter: true})
	if g.State() != StateModeSelect || !g.confirmingChaos || len(sessionEvents) != 0 {
		t.Fatal("first Enter on chaos must only ask for confirmation")
	}
	// The cursor is frozen while confirming, so the guide cannot be reached.
	g.handleModeSelect(menuKeys{down: true})
	if g.modeCursor != entryChaos {
		t.Fatalf("expected the cursor to stay on chaos, got %v", g.modeCursor)
	}
	g.handleModeSelect(menuKeys{enter: true})
	if g.State() != StatePlaying || !g.chaosMode {
		t.Fatalf("expected a chaos session, got %v", g.State())
	}
}

func TestModeSelect_EscFromGuideEntryGoesToTitle(t *testing.T) {
	g, _ := goToMode(t)
	g.modeCursor = entryGuide
	g.handleModeSelect(menuKeys{esc: true})
	if g.State() != StateTitle {
		t.Fatalf("expected title, got %v", g.State())
	}
}

func TestFieldGuide_TierKeysPickDemoTier(t *testing.T) {
	g, _ := goToMode(t)
	g.modeCursor = entryGuide
	g.handleModeSelect(menuKeys{enter: true})
	g.handleFieldGuide(menuKeys{tier: 3})
	if g.guide.tier != TierHeavy {
		t.Fatalf("expected Heavy, got %v", g.guide.tier)
	}
}

func TestFieldGuide_TableHasOneRowPerOwnerKindAndAnUnmanagedRow(t *testing.T) {
	unmanaged := 0
	archetypes := map[Archetype]bool{}
	for _, p := range guidePods {
		if !p.mapped {
			unmanaged++
			continue
		}
		archetypes[p.arch] = true
	}
	if unmanaged != 1 {
		t.Fatalf("expected exactly one unmanaged row, got %d", unmanaged)
	}
	for a := ArchetypeProbe; a <= ArchetypeDart; a++ {
		if !archetypes[a] {
			t.Fatalf("expected a row for archetype %v", a)
		}
	}
}

func TestSymptomPose(t *testing.T) {
	for restarts, want := range map[int]int{0: 0, 1: 1, 4: 1, 5: 2, 9: 2} {
		if got := symptomPose(Symptoms{Ready: true, Restarts: restarts}, 0, 0, 1).plates; got != want {
			t.Fatalf("restarts %d: expected %d plates, got %d", restarts, want, got)
		}
	}
	if l := symptomPose(Symptoms{Ready: true}, 0, 5, 1); l.eyeLevel != 1 || l.sparks {
		t.Fatalf("a ready Pod must not flicker, got %+v", l)
	}
	flickered := false
	for tick := range uint64(300) {
		if symptomPose(Symptoms{}, 0, tick, 1).eyeLevel < 1 {
			flickered = true
		}
	}
	if !flickered {
		t.Fatal("expected a not-ready Pod to flicker")
	}
	if !symptomPose(Symptoms{Ready: true}, rebootBlinkTicks/2, 0, 1).eyeShut {
		t.Fatal("expected the eye shut mid-blink")
	}
}

func TestFieldGuide_DemoRunsEveryArchetypeAndTier(t *testing.T) {
	for a := ArchetypeProbe; a <= ArchetypeDart; a++ {
		for ti := TierLight; ti <= TierHeavy; ti++ {
			d := newGuideDemo(a, ti)
			for range 400 {
				d.step()
			}
			if len(d.scene.Enemies()) != 1 {
				t.Fatalf("%v/%v: expected one demo enemy", a, ti)
			}
		}
	}
}

func TestFieldGuide_DemoTurretFiresBolts(t *testing.T) {
	d := newGuideDemo(ArchetypeTurret, TierStandard)
	fired := false
	for range 600 {
		d.step()
		if len(d.scene.Bolts()) > 0 {
			fired = true
		}
	}
	if !fired {
		t.Fatal("expected the demo Turret to fire Bolts")
	}
}

func TestFieldGuide_SmokeDrawsEveryPage(t *testing.T) {
	g, _ := goToMode(t)
	g.modeCursor = entryGuide
	g.handleModeSelect(menuKeys{enter: true})
	img := ebitenNewImage()
	for range guidePageCount {
		for range 5 {
			g.handleFieldGuide(menuKeys{})
			g.Draw(img)
		}
		g.handleFieldGuide(menuKeys{right: true})
	}
}

func ebitenNewImage() *ebiten.Image { return ebiten.NewImage(int(screenWidth), int(screenHeight)) }
