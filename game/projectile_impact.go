package game

import "math"

const projectileImpactTicks = 18

// ProjectileImpact is a short hit effect left by a player projectile.
type ProjectileImpact struct {
	X, Y      float64
	Kind      ProjectileKind
	Direction float64
	Age       int
}

// ProjectileImpacts returns the player projectile hit effects currently showing.
func (s *Scene) ProjectileImpacts() []ProjectileImpact { return s.projectileImpacts }

func (s *Scene) clearProjectileImpacts() {
	s.projectileImpacts = nil
}

func (s *Scene) updateProjectileImpacts() {
	impacts := s.projectileImpacts[:0]
	for _, impact := range s.projectileImpacts {
		impact.Age++
		if impact.Age < projectileImpactTicks {
			impacts = append(impacts, impact)
		}
	}
	s.projectileImpacts = impacts
}

func (s *Scene) recordProjectileImpact(projectile *Projectile) {
	if projectile == nil {
		return
	}
	s.projectileImpacts = append(s.projectileImpacts, ProjectileImpact{
		X:         projectile.X,
		Y:         projectile.Y,
		Kind:      projectile.Kind,
		Direction: math.Atan2(projectile.VY, projectile.VX),
	})
}
