package game

import (
	"fmt"
	"testing"
)

// Separation is O(n²) per sweep, so the tick cost grows quickly with the pool.
// A frame budget at 60fps is ~16.6ms; this pins how much of it a crowded
// namespace consumes.
func BenchmarkSceneUpdateCrowd(b *testing.B) {
	for _, n := range []int{50, 200, 500} {
		b.Run(fmt.Sprintf("enemies=%d", n), func(b *testing.B) {
			scene := NewScene(make(chan GameEvent), make(chan KillEvent, 4096))
			scene.BeginSession(1, false)
			for i := range n {
				x, y := randomSpawnPosition()
				scene.enemies = append(scene.enemies, NewEnemy("pod", "default", x+float64(i%7), y))
			}
			player := NewPlayer(400, 300)
			// Let the crowd settle into its steady state before measuring.
			for range 300 {
				player.Health = PlayerMaxHealth
				scene.Update(player)
			}

			b.ResetTimer()
			for range b.N {
				player.Health = PlayerMaxHealth
				scene.Update(player)
			}
		})
	}
}
