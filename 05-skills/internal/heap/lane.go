package heap

import "fmt"

// GenerationType identifies the JVM memory generation pool.
type GenerationType int

const (
	GenEden GenerationType = iota
	GenSurvivor0
	GenSurvivor1
	GenTenured
)

// String returns a human-readable name for the memory generation.
func (g GenerationType) String() string {
	switch g {
	case GenEden:
		return "Eden"
	case GenSurvivor0:
		return "Survivor0"
	case GenSurvivor1:
		return "Survivor1"
	case GenTenured:
		return "Tenured"
	default:
		return fmt.Sprintf("Generation(%d)", g)
	}
}

// LaneConfig holds static parameters for a memory generation lane.
type LaneConfig struct {
	GenType    GenerationType
	ColorHex   uint32
	CapacityMB int
}

// MemoryLane tracks runtime allocation and capacity for a specific memory generation.
type MemoryLane struct {
	Config             LaneConfig
	CurrentAllocatedMB int
}

// NewMemoryLane creates a new MemoryLane with zero current allocations.
func NewMemoryLane(config LaneConfig) *MemoryLane {
	return &MemoryLane{
		Config:             config,
		CurrentAllocatedMB: 0,
	}
}

// Allocate increases the lane's allocated memory by mb.
// Returns true if the allocation succeeded without triggering OutOfMemory (usage < capacity),
// or false if the lane has reached or exceeded its capacity.
func (l *MemoryLane) Allocate(mb int) bool {
	if mb <= 0 {
		return !l.IsOOM()
	}
	l.CurrentAllocatedMB += mb
	return !l.IsOOM()
}

// Deallocate decreases the lane's allocated memory by mb, clamped at zero.
func (l *MemoryLane) Deallocate(mb int) {
	if mb <= 0 {
		return
	}
	l.CurrentAllocatedMB -= mb
	if l.CurrentAllocatedMB < 0 {
		l.CurrentAllocatedMB = 0
	}
}

// UsageRatio returns the fraction of capacity currently in use [0.0, 1.0+].
func (l *MemoryLane) UsageRatio() float64 {
	if l.Config.CapacityMB <= 0 {
		return 0.0
	}
	return float64(l.CurrentAllocatedMB) / float64(l.Config.CapacityMB)
}

// IsOOM returns true if the current allocated memory has reached or exceeded maximum capacity.
func (l *MemoryLane) IsOOM() bool {
	return l.CurrentAllocatedMB >= l.Config.CapacityMB
}

// Reset clears all allocated memory back to 0 MB.
func (l *MemoryLane) Reset() {
	l.CurrentAllocatedMB = 0
}

// NewDefaultLanes initializes the four standard JVM memory generation lanes with default capacities and colors:
// - Eden: 256MB, Neon Emerald Green (#00E676)
// - Survivor0: 64MB, Warm Amber/Gold (#FFD600)
// - Survivor1: 64MB, Warm Amber/Gold (#FFD600)
// - Tenured: 512MB, High-alert Crimson/Neon Red (#FF1744)
func NewDefaultLanes() map[GenerationType]*MemoryLane {
	return map[GenerationType]*MemoryLane{
		GenEden: NewMemoryLane(LaneConfig{
			GenType:    GenEden,
			ColorHex:   0x00E676,
			CapacityMB: 256,
		}),
		GenSurvivor0: NewMemoryLane(LaneConfig{
			GenType:    GenSurvivor0,
			ColorHex:   0xFFD600,
			CapacityMB: 64,
		}),
		GenSurvivor1: NewMemoryLane(LaneConfig{
			GenType:    GenSurvivor1,
			ColorHex:   0xFFD600,
			CapacityMB: 64,
		}),
		GenTenured: NewMemoryLane(LaneConfig{
			GenType:    GenTenured,
			ColorHex:   0xFF1744,
			CapacityMB: 512,
		}),
	}
}
