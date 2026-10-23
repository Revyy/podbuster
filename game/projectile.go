package game

const projectileSize = 8.0

// Rect is an axis-aligned bounding box used for collision detection.
type Rect struct {
	X, Y, W, H float64
}

// Overlaps returns true if r and other intersect (touching edges do not count).
func (r Rect) Overlaps(other Rect) bool {
	return r.X < other.X+other.W &&
		r.X+r.W > other.X &&
		r.Y < other.Y+other.H &&
		r.Y+r.H > other.Y
}

// Projectile is a shared combat projectile. Each projectile carries the hit cost
// of the weapon that created it, so collisions apply a single data-driven
// impact rather than branching on projectile type.
type Projectile struct {
	X, Y    float64
	VX, VY  float64
	Active  bool
	HitCost int
	Size    float64
	Kind    ProjectileKind
	// Age counts ticks since creation. Rendering uses it for transient flight
	// and spawn effects; collision logic ignores it.
	Age int
}

// ProjectileKind distinguishes projectiles for rendering only; collision
// resolution never branches on it.
type ProjectileKind int

const (
	projectileKindBasic ProjectileKind = iota
	projectileKindRocket
)

// NewProjectile creates an active projectile at (x, y) moving at (vx, vy) and
// costing hitCost points of enemy hull.
func NewProjectile(x, y, vx, vy float64, hitCost int) *Projectile {
	if hitCost < 0 {
		hitCost = 0
	}
	return &Projectile{X: x, Y: y, VX: vx, VY: vy, Active: true, HitCost: hitCost, Size: projectileSize, Kind: projectileKindBasic}
}

// Update moves the projectile by its velocity.
func (p *Projectile) Update() {
	p.X += p.VX
	p.Y += p.VY
	p.Age++
}

// Bounds returns the collision rectangle for the projectile.
func (p *Projectile) Bounds() Rect {
	half := p.Size / 2
	return Rect{X: p.X - half, Y: p.Y - half, W: p.Size, H: p.Size}
}
