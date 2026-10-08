package art

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ParticleShape determines the rendering style of a particle.
type ParticleShape int

const (
	ShapePoint ParticleShape = iota
	ShapeSpark
	ShapeSquare
	ShapeRing
)

// Particle represents an individual visual effect entity.
type Particle struct {
	X, Y       float32
	VX, VY     float32
	Color      color.RGBA
	Alpha      float32
	Life       float32 // Remaining life in frames or seconds
	MaxLife    float32 // Initial life
	Size       float32
	Shape      ParticleShape
	Active     bool
}

// ParticlePool manages reusable particle slices with zero runtime allocations.
type ParticlePool struct {
	particles []Particle
	capacity  int
	nextFree  int
}

// NewParticlePool allocates a fixed-size pool for particles.
func NewParticlePool(capacity int) *ParticlePool {
	if capacity <= 0 {
		capacity = 512
	}
	return &ParticlePool{
		particles: make([]Particle, capacity),
		capacity:  capacity,
		nextFree:  0,
	}
}

// Spawn searches for an inactive particle or reuses the next slot in round-robin fashion.
func (p *ParticlePool) Spawn(x, y, vx, vy float32, clr color.RGBA, life float32, size float32, shape ParticleShape) *Particle {
	// Try finding an inactive particle
	idx := -1
	for i := 0; i < p.capacity; i++ {
		candidate := (p.nextFree + i) % p.capacity
		if !p.particles[candidate].Active {
			idx = candidate
			break
		}
	}

	// If all are active, steal the oldest/next slot
	if idx == -1 {
		idx = p.nextFree
	}
	p.nextFree = (idx + 1) % p.capacity

	pt := &p.particles[idx]
	pt.X = x
	pt.Y = y
	pt.VX = vx
	pt.VY = vy
	pt.Color = clr
	pt.Alpha = float32(clr.A) / 255.0
	pt.Life = life
	pt.MaxLife = life
	pt.Size = size
	pt.Shape = shape
	pt.Active = true

	return pt
}

// EmitDeallocationBurst spawns a shower of spark/ring particles when an object is freed.
func (p *ParticlePool) EmitDeallocationBurst(x, y float32, count int, clr color.RGBA) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := float32(1.0 + rand.Float64()*3.5)
		vx := float32(math.Cos(angle)) * speed
		vy := float32(math.Sin(angle)) * speed

		life := float32(15 + rand.Intn(20))
		size := float32(1.5 + rand.Float64()*2.0)

		shape := ShapeSpark
		if i%3 == 0 {
			shape = ShapeRing
		} else if i%2 == 0 {
			shape = ShapeSquare
		}

		p.Spawn(x, y, vx, vy, clr, life, size, shape)
	}
}

// EmitBeamImpactSparks spawns high-velocity sparks at beam contact points.
func (p *ParticlePool) EmitBeamImpactSparks(x, y float32, normalX, normalY float32, count int, clr color.RGBA) {
	baseAngle := math.Atan2(float64(normalY), float64(normalX))
	for i := 0; i < count; i++ {
		// Cone of 90 degrees around normal
		angle := baseAngle + (rand.Float64()-0.5)*math.Pi/2
		speed := float32(1.5 + rand.Float64()*4.0)
		vx := float32(math.Cos(angle)) * speed
		vy := float32(math.Sin(angle)) * speed

		life := float32(8 + rand.Intn(12))
		size := float32(1.0 + rand.Float64()*1.5)

		p.Spawn(x, y, vx, vy, clr, life, size, ShapeSpark)
	}
}

// EmitPromotionTrail spawns subtle rising golden/amber energy trail particles.
func (p *ParticlePool) EmitPromotionTrail(x, y float32, clr color.RGBA) {
	vx := float32((rand.Float64() - 0.5) * 0.8)
	vy := float32(-0.8 - rand.Float64()*1.2) // Float upwards
	life := float32(20 + rand.Intn(15))
	size := float32(1.5 + rand.Float64()*1.5)

	p.Spawn(x, y, vx, vy, clr, life, size, ShapeRing)
}

// Update advances particle simulations and deactivates expired ones.
func (p *ParticlePool) Update() {
	for i := range p.particles {
		pt := &p.particles[i]
		if !pt.Active {
			continue
		}

		pt.Life--
		if pt.Life <= 0 {
			pt.Active = false
			continue
		}

		// Update position
		pt.X += pt.VX
		pt.Y += pt.VY

		// Add subtle drag
		pt.VX *= 0.95
		pt.VY *= 0.95

		// Fade alpha
		ratio := pt.Life / pt.MaxLife
		baseAlpha := float32(pt.Color.A) / 255.0
		pt.Alpha = baseAlpha * ratio
	}
}

// Draw renders all active particles onto the destination image.
func (p *ParticlePool) Draw(dst *ebiten.Image) {
	for i := range p.particles {
		pt := &p.particles[i]
		if !pt.Active || pt.Alpha <= 0 {
			continue
		}

		drawColor := color.RGBA{
			R: pt.Color.R,
			G: pt.Color.G,
			B: pt.Color.B,
			A: uint8(pt.Alpha * 255.0),
		}

		switch pt.Shape {
		case ShapePoint:
			vector.DrawFilledCircle(dst, pt.X, pt.Y, pt.Size, drawColor, true)
		case ShapeSpark:
			// Elongated spark in direction of velocity
			endX := pt.X + pt.VX*1.8
			endY := pt.Y + pt.VY*1.8
			vector.StrokeLine(dst, pt.X, pt.Y, endX, endY, pt.Size, drawColor, true)
		case ShapeSquare:
			half := pt.Size * 0.5
			vector.DrawFilledRect(dst, pt.X-half, pt.Y-half, pt.Size, pt.Size, drawColor, true)
		case ShapeRing:
			// Expanding ring
			ringRadius := pt.Size * (1.0 + (1.0 - pt.Life/pt.MaxLife)*2.0)
			vector.StrokeCircle(dst, pt.X, pt.Y, ringRadius, 1.0, drawColor, true)
		}
	}
}

// ActiveCount returns how many particles are currently active in the pool.
func (p *ParticlePool) ActiveCount() int {
	count := 0
	for i := range p.particles {
		if p.particles[i].Active {
			count++
		}
	}
	return count
}

// Capacity returns the total allocated capacity of the pool.
func (p *ParticlePool) Capacity() int {
	return p.capacity
}
