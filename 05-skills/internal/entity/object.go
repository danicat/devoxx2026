package entity

import (
	"fmt"
	"math"
)

// ObjectType represents the type/profile of an allocated memory object.
type ObjectType int

const (
	ObjLambda      ObjectType = iota // Fast, low HP, ephemeral
	ObjString                        // Medium HP, prone to promotion
	ObjThreadLocal                   // Armored, cyclic references
	ObjLargeBlob                     // Boss heavyweight allocation
)

// Generation constants representing JVM memory spaces.
const (
	GenEden      = 0
	GenSurvivor0 = 1
	GenSurvivor1 = 2
	GenTenured   = 3
)

// AllocationObject represents a dynamic heap allocation progressing through memory lanes.
type AllocationObject struct {
	ID               int64
	Type             ObjectType
	SizeMB           int
	MaxHealth        float64
	Health           float64
	Speed            float64
	PathIndex        int
	X, Y             float64
	TenureAge        int
	IsDead           bool
	ReachedEnd       bool
	Armor            float64 // Damage reduction factor (0.0 to 1.0, e.g. 0.35 for ThreadLocal)
	Generation       int     // 0 = Eden, 1 = S0, 2 = S1, 3 = Tenured
	DistanceTraveled float64
}

// NewAllocationObject creates and initializes an AllocationObject with type-specific attributes.
func NewAllocationObject(id int64, objType ObjectType, startX, startY float64) *AllocationObject {
	var (
		sizeMB    int
		maxHealth float64
		speed     float64
		armor     float64
	)

	switch objType {
	case ObjLambda:
		sizeMB = 5
		maxHealth = 25.0
		speed = 2.2
		armor = 0.0
	case ObjString:
		sizeMB = 20
		maxHealth = 60.0
		speed = 1.4
		armor = 0.0
	case ObjThreadLocal:
		sizeMB = 50
		maxHealth = 150.0
		speed = 0.9
		armor = 0.35
	case ObjLargeBlob:
		sizeMB = 200
		maxHealth = 600.0
		speed = 0.5
		armor = 0.10
	default:
		panic(fmt.Sprintf("unknown ObjectType: %d", objType))
	}

	return &AllocationObject{
		ID:               id,
		Type:             objType,
		SizeMB:           sizeMB,
		MaxHealth:        maxHealth,
		Health:           maxHealth,
		Speed:            speed,
		PathIndex:        0,
		X:                startX,
		Y:                startY,
		TenureAge:        0,
		IsDead:           false,
		ReachedEnd:       false,
		Armor:            armor,
		Generation:       GenEden,
		DistanceTraveled: 0,
	}
}

// TakeDamage applies damage to the object factoring in its armor damage reduction.
// If Health reaches or drops below 0, Health is set to 0 and IsDead is marked true.
func (o *AllocationObject) TakeDamage(amount float64) {
	if o.IsDead || amount <= 0 {
		return
	}

	reduction := o.Armor
	if reduction < 0 {
		reduction = 0
	} else if reduction > 1.0 {
		reduction = 1.0
	}

	effectiveDamage := amount * (1.0 - reduction)
	if effectiveDamage < 0 {
		effectiveDamage = 0
	}

	o.Health -= effectiveDamage
	if o.Health <= 0 {
		o.Health = 0
		o.IsDead = true
	}
}

// Age increments TenureAge and handles generation transition logic.
// Generation transitions: Eden (age 0) -> S0 (age >= 1) -> S1 (age >= 2) -> Tenured (age >= 4).
func (o *AllocationObject) Age() {
	o.TenureAge++
	if o.TenureAge >= 4 {
		o.Generation = GenTenured
	} else if o.TenureAge >= 2 {
		o.Generation = GenSurvivor1
	} else if o.TenureAge >= 1 {
		o.Generation = GenSurvivor0
	} else {
		o.Generation = GenEden
	}
}

// MoveTowards advances X, Y towards (targetX, targetY) by at most step distance,
// updating DistanceTraveled. Returns true if target was reached.
func (o *AllocationObject) MoveTowards(targetX, targetY float64, step float64) bool {
	if step <= 0 {
		return false
	}
	dx := targetX - o.X
	dy := targetY - o.Y
	dist := math.Hypot(dx, dy)
	if dist <= step {
		o.X = targetX
		o.Y = targetY
		o.DistanceTraveled += dist
		return true
	}
	ratio := step / dist
	o.X += dx * ratio
	o.Y += dy * ratio
	o.DistanceTraveled += step
	return false
}

// UpdateWithPath navigates the object through the provided waypoints according to its Speed.
// When the final waypoint is reached, ReachedEnd is set to true.
func (o *AllocationObject) UpdateWithPath(waypoints [][2]float64) {
	if o.IsDead || o.ReachedEnd || len(waypoints) == 0 {
		return
	}

	if o.PathIndex >= len(waypoints) {
		o.ReachedEnd = true
		return
	}

	remainingStep := o.Speed
	for remainingStep > 0 && o.PathIndex < len(waypoints) {
		target := waypoints[o.PathIndex]
		dx := target[0] - o.X
		dy := target[1] - o.Y
		dist := math.Hypot(dx, dy)

		if dist <= remainingStep {
			o.X = target[0]
			o.Y = target[1]
			o.DistanceTraveled += dist
			remainingStep -= dist
			o.PathIndex++
			if o.PathIndex >= len(waypoints) {
				o.ReachedEnd = true
				break
			}
		} else {
			ratio := remainingStep / dist
			o.X += dx * ratio
			o.Y += dy * ratio
			o.DistanceTraveled += remainingStep
			remainingStep = 0
		}
	}
}
