package entity

import (
	"fmt"
	"math"
)

// CollectorType represents the operational garbage collector algorithm.
type CollectorType int

const (
	ColSerial   CollectorType = iota // Single-target beam
	ColParallel                      // High-throughput burst
	ColCMS                           // Radial mark-sweep
	ColG1                            // Prioritizes densest garbage
	ColZGC                           // High-speed low-latency laser
)

// MaxCollectorLevel specifies the maximum upgrade level for any collector tower.
const MaxCollectorLevel = 5

// Collector represents a GC tower placed on the heap grid capable of scanning and deallocating memory objects.
type Collector struct {
	ID          int64
	Type        CollectorType
	GridX       int
	GridY       int
	WorldX      float64
	WorldY      float64
	Range       float64
	DPS         float64
	FireRate    float64
	Cooldown    float64
	TargetID    int64
	Level       int
	Cost        int
	UpgradeCost int

	nextProjectileID int64
}

// NewCollector constructs and initializes a Collector with base stats matching the GDD and architecture.
func NewCollector(id int64, colType CollectorType, gridX, gridY int, worldX, worldY float64) *Collector {
	var (
		cost        int
		rng         float64
		fireRate    float64
		dps         float64
		upgradeCost int
	)

	switch colType {
	case ColSerial:
		cost = 50
		rng = 120.0
		fireRate = 2.0
		dps = 30.0
		upgradeCost = 40
	case ColParallel:
		cost = 120
		rng = 100.0
		fireRate = 4.0
		dps = 60.0
		upgradeCost = 90
	case ColCMS:
		cost = 200
		rng = 140.0
		fireRate = 1.5
		dps = 45.0
		upgradeCost = 150
	case ColG1:
		cost = 350
		rng = 160.0
		fireRate = 2.5
		dps = 90.0
		upgradeCost = 250
	case ColZGC:
		cost = 500
		rng = 180.0
		fireRate = 5.0
		dps = 150.0
		upgradeCost = 350
	default:
		panic(fmt.Sprintf("unknown CollectorType: %d", colType))
	}

	return &Collector{
		ID:               id,
		Type:             colType,
		GridX:            gridX,
		GridY:            gridY,
		WorldX:           worldX,
		WorldY:           worldY,
		Range:            rng,
		DPS:              dps,
		FireRate:         fireRate,
		Cooldown:         0.0,
		TargetID:         0,
		Level:            1,
		Cost:             cost,
		UpgradeCost:      upgradeCost,
		nextProjectileID: 1,
	}
}

// CanUpgrade returns true if the collector has not reached max level and the player has enough CPU cycles.
func (c *Collector) CanUpgrade(currentCycles int) bool {
	if c.Level >= MaxCollectorLevel {
		return false
	}
	return currentCycles >= c.UpgradeCost
}

// Upgrade advances the collector to the next level, boosting DPS by 25% and Range by 15%,
// and scaling the upgrade cost. Returns an error if the collector is already at max level.
func (c *Collector) Upgrade() error {
	if c.Level >= MaxCollectorLevel {
		return fmt.Errorf("collector %d (type %d) is already at maximum level %d", c.ID, c.Type, MaxCollectorLevel)
	}

	c.Level++
	c.DPS *= 1.25
	c.Range *= 1.15
	c.UpgradeCost = int(math.Round(float64(c.UpgradeCost) * 1.5))
	return nil
}

// FindTarget selects the optimal target object within range according to the collector type's targeting strategy.
// Returns nil if no valid alive target is in range.
func (c *Collector) FindTarget(objects []*AllocationObject) *AllocationObject {
	var (
		bestTarget *AllocationObject
		maxVal     float64
		minVal     float64 = math.MaxFloat64
	)

	for _, obj := range objects {
		if obj == nil || obj.IsDead {
			continue
		}

		dist := math.Hypot(obj.X-c.WorldX, obj.Y-c.WorldY)
		if dist > c.Range {
			continue
		}

		switch c.Type {
		case ColSerial:
			// Prioritizes object closest to the end (greatest DistanceTraveled), breaking ties by closest distance to collector
			if bestTarget == nil || obj.DistanceTraveled > maxVal || (obj.DistanceTraveled == maxVal && dist < minVal) {
				bestTarget = obj
				maxVal = obj.DistanceTraveled
				minVal = dist
			}
		case ColParallel:
			// Nearest object in range
			if bestTarget == nil || dist < minVal {
				bestTarget = obj
				minVal = dist
			}
		case ColCMS:
			// Radial AOE: any candidate or nearest object as primary anchor
			if bestTarget == nil || dist < minVal {
				bestTarget = obj
				minVal = dist
			}
		case ColG1:
			// Prioritizes highest current Health (or densest SizeMB)
			// Primary metric: Health; tiebreaker: SizeMB
			score := obj.Health
			if bestTarget == nil || score > maxVal || (score == maxVal && float64(obj.SizeMB) > minVal) {
				bestTarget = obj
				maxVal = score
				minVal = float64(obj.SizeMB)
			}
		case ColZGC:
			// Prioritizes highest TenureAge (or highest Generation)
			// Score formula: TenureAge * 100 + Generation
			score := float64(obj.TenureAge*100 + obj.Generation)
			if bestTarget == nil || score > maxVal || (score == maxVal && obj.Health > minVal) {
				bestTarget = obj
				maxVal = score
				minVal = obj.Health
			}
		}
	}

	return bestTarget
}

// Update ticks the collector cooldown, checks for targets, applies damage, and spawns projectiles.
// isSTW grants a 2.0x damage multiplier during Stop-The-World compaction.
func (c *Collector) Update(dt float64, objects []*AllocationObject, isSTW bool) []*Projectile {
	if c.Cooldown > 0 {
		c.Cooldown -= dt
		if c.Cooldown < 0 {
			c.Cooldown = 0
		}
	}

	if c.Cooldown > 0 {
		return nil
	}

	baseCooldown := 1.0 / c.FireRate
	damageMultiplier := 1.0
	if isSTW {
		damageMultiplier = 2.0
	}

	damagePerShot := (c.DPS / c.FireRate) * damageMultiplier

	// ColCMS deals radial AOE damage to ALL alive targets in range
	if c.Type == ColCMS {
		var (
			projectiles []*Projectile
			hitAny      bool
		)

		for _, obj := range objects {
			if obj == nil || obj.IsDead {
				continue
			}
			dist := math.Hypot(obj.X-c.WorldX, obj.Y-c.WorldY)
			if dist <= c.Range {
				hitAny = true
				obj.TakeDamage(damagePerShot)
				c.TargetID = obj.ID

				projID := c.ID*100000 + c.nextProjectileID
				c.nextProjectileID++

				projectiles = append(projectiles, NewProjectile(
					projID,
					ProjRadial,
					c.WorldX,
					c.WorldY,
					obj.X,
					obj.Y,
					obj.ID,
					0.18,
					0x00E676, // Neon emerald green
				))
			}
		}

		if hitAny {
			c.Cooldown = baseCooldown
			return projectiles
		}
		return nil
	}

	// Single target collectors
	target := c.FindTarget(objects)
	if target == nil {
		c.TargetID = 0
		return nil
	}

	c.TargetID = target.ID
	target.TakeDamage(damagePerShot)
	c.Cooldown = baseCooldown

	var (
		projType ProjectileType
		colorHex uint32
		duration float64
	)

	switch c.Type {
	case ColSerial:
		projType = ProjBeam
		colorHex = 0x00E676 // Emerald Beam
		duration = 0.15
	case ColParallel:
		projType = ProjBurst
		colorHex = 0xFFD600 // Amber Burst
		duration = 0.12
	case ColG1:
		projType = ProjLaser
		colorHex = 0x2979FF // Electric Blue Laser
		duration = 0.15
	case ColZGC:
		projType = ProjLaser
		colorHex = 0xFF1744 // Neon Red Colored Pointer Laser
		duration = 0.10
	default:
		projType = ProjBeam
		colorHex = 0x00E676
		duration = 0.15
	}

	projID := c.ID*100000 + c.nextProjectileID
	c.nextProjectileID++

	return []*Projectile{
		NewProjectile(
			projID,
			projType,
			c.WorldX,
			c.WorldY,
			target.X,
			target.Y,
			target.ID,
			duration,
			colorHex,
		),
	}
}
