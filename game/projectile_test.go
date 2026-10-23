package game

import "testing"

func TestRect_Overlaps_True(t *testing.T) {
	a := Rect{X: 0, Y: 0, W: 10, H: 10}
	b := Rect{X: 5, Y: 5, W: 10, H: 10}
	if !a.Overlaps(b) {
		t.Fatal("expected rects to overlap")
	}
}

func TestRect_Overlaps_False(t *testing.T) {
	a := Rect{X: 0, Y: 0, W: 10, H: 10}
	b := Rect{X: 20, Y: 20, W: 10, H: 10}
	if a.Overlaps(b) {
		t.Fatal("expected rects not to overlap")
	}
}

func TestRect_Overlaps_Touching(t *testing.T) {
	a := Rect{X: 0, Y: 0, W: 10, H: 10}
	b := Rect{X: 10, Y: 0, W: 10, H: 10} // edge-adjacent, not overlapping
	if a.Overlaps(b) {
		t.Fatal("edge-touching rects should not overlap")
	}
}

func TestProjectile_Update_MovesPosition(t *testing.T) {
	projectile := NewProjectile(100, 100, 5, -3, 1)
	projectile.Update()
	if projectile.X != 105 || projectile.Y != 97 {
		t.Fatalf("expected (105, 97), got (%v, %v)", projectile.X, projectile.Y)
	}
}

func TestProjectile_Bounds_CenteredOnPosition(t *testing.T) {
	projectile := NewProjectile(50, 60, 0, 0, 1)
	b := projectile.Bounds()
	if b.X != 46 || b.Y != 56 {
		t.Fatalf("expected bounds origin (46, 56), got (%v, %v)", b.X, b.Y)
	}
	if b.W != 8 || b.H != 8 {
		t.Fatalf("expected bounds 8x8, got %vx%v", b.W, b.H)
	}
}
