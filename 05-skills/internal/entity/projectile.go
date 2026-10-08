package entity

// ProjectileType represents the visual and mechanical category of a GC collector projectile/beam.
type ProjectileType int

const (
	ProjBeam   ProjectileType = iota // Single-target focused beam (e.g. Serial)
	ProjBurst                        // High-throughput burst/bolt (e.g. Parallel)
	ProjRadial                       // Radial mark-sweep pulse (e.g. CMS)
	ProjLaser                        // Ultra-low-latency colored-pointer laser (e.g. ZGC, G1)
)

// Projectile represents an active in-flight beam, ray, or pulse fired by a collector tower.
type Projectile struct {
	ID        int64
	Type      ProjectileType
	StartX    float64
	StartY    float64
	TargetX   float64
	TargetY   float64
	TargetID  int64
	Duration  float64 // Total lifespan in seconds
	Elapsed   float64 // Elapsed time in seconds
	IsExpired bool
	ColorHex  uint32
}

// NewProjectile creates and initializes a new Projectile.
func NewProjectile(id int64, projType ProjectileType, startX, startY, targetX, targetY float64, targetID int64, duration float64, colorHex uint32) *Projectile {
	return &Projectile{
		ID:        id,
		Type:      projType,
		StartX:    startX,
		StartY:    startY,
		TargetX:   targetX,
		TargetY:   targetY,
		TargetID:  targetID,
		Duration:  duration,
		Elapsed:   0,
		IsExpired: false,
		ColorHex:  colorHex,
	}
}

// Update advances the projectile lifetime by dt seconds.
// If Elapsed reaches or exceeds Duration, IsExpired is marked true.
func (p *Projectile) Update(dt float64) {
	if p.IsExpired {
		return
	}
	p.Elapsed += dt
	if p.Elapsed >= p.Duration {
		p.IsExpired = true
	}
}
