package art

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestParticlePool(t *testing.T) {
	pool := NewParticlePool(100)
	if pool.Capacity() != 100 {
		t.Fatalf("expected capacity 100, got %d", pool.Capacity())
	}
	if pool.ActiveCount() != 0 {
		t.Fatalf("expected active count 0, got %d", pool.ActiveCount())
	}

	// Spawn a particle
	p := pool.Spawn(10, 10, 1, -1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 10, 2, ShapePoint)
	if p == nil {
		t.Fatalf("expected particle, got nil")
	}
	if !p.Active {
		t.Errorf("expected particle to be active")
	}
	if pool.ActiveCount() != 1 {
		t.Errorf("expected active count 1, got %d", pool.ActiveCount())
	}

	// Emit burst
	pool.EmitDeallocationBurst(20, 20, 15, color.RGBA{R: 0, G: 230, B: 118, A: 255})
	if pool.ActiveCount() != 16 {
		t.Errorf("expected active count 16, got %d", pool.ActiveCount())
	}

	// Emit impact sparks
	pool.EmitBeamImpactSparks(25, 25, 1, 0, 8, color.RGBA{R: 0, G: 200, B: 255, A: 255})
	if pool.ActiveCount() != 24 {
		t.Errorf("expected active count 24, got %d", pool.ActiveCount())
	}

	// Emit promotion trail
	pool.EmitPromotionTrail(30, 30, color.RGBA{R: 255, G: 215, B: 0, A: 255})
	if pool.ActiveCount() != 25 {
		t.Errorf("expected active count 25, got %d", pool.ActiveCount())
	}

	// Update simulation
	pool.Update()

	// Draw onto target
	dst := ebiten.NewImage(100, 100)
	pool.Draw(dst)

	// Simulate until all particles expire
	for i := 0; i < 50; i++ {
		pool.Update()
	}

	if pool.ActiveCount() != 0 {
		t.Errorf("expected all particles to expire after 50 frames, active count: %d", pool.ActiveCount())
	}
}
