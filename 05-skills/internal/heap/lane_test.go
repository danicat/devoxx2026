package heap

import (
	"math"
	"testing"
)

func TestGenerationTypeString(t *testing.T) {
	if GenEden.String() != "Eden" {
		t.Errorf("expected 'Eden', got %s", GenEden.String())
	}
	if GenSurvivor0.String() != "Survivor0" {
		t.Errorf("expected 'Survivor0', got %s", GenSurvivor0.String())
	}
	if GenSurvivor1.String() != "Survivor1" {
		t.Errorf("expected 'Survivor1', got %s", GenSurvivor1.String())
	}
	if GenTenured.String() != "Tenured" {
		t.Errorf("expected 'Tenured', got %s", GenTenured.String())
	}
	if GenerationType(42).String() != "Generation(42)" {
		t.Errorf("expected 'Generation(42)', got %s", GenerationType(42).String())
	}
}

func TestMemoryLaneAllocationAndDeallocation(t *testing.T) {
	lane := NewMemoryLane(LaneConfig{
		GenType:    GenEden,
		ColorHex:   0x00E676,
		CapacityMB: 100,
	})

	if lane.CurrentAllocatedMB != 0 {
		t.Errorf("expected 0 allocated, got %d", lane.CurrentAllocatedMB)
	}
	if lane.IsOOM() {
		t.Errorf("new lane should not be in OOM state")
	}

	// Normal allocation
	ok := lane.Allocate(50)
	if !ok {
		t.Errorf("expected allocation of 50MB to succeed")
	}
	if lane.CurrentAllocatedMB != 50 {
		t.Errorf("expected 50MB, got %d", lane.CurrentAllocatedMB)
	}
	if math.Abs(lane.UsageRatio()-0.5) > 1e-6 {
		t.Errorf("expected usage ratio 0.5, got %f", lane.UsageRatio())
	}

	// Non-positive allocation
	ok = lane.Allocate(0)
	if !ok || lane.CurrentAllocatedMB != 50 {
		t.Errorf("allocate 0 should be no-op")
	}
	ok = lane.Allocate(-10)
	if !ok || lane.CurrentAllocatedMB != 50 {
		t.Errorf("allocate negative should be no-op")
	}

	// Allocate up to capacity limit -> triggers OOM condition
	ok = lane.Allocate(50)
	if ok {
		t.Errorf("expected allocation to reach capacity to return false (!IsOOM)")
	}
	if lane.CurrentAllocatedMB != 100 {
		t.Errorf("expected 100MB, got %d", lane.CurrentAllocatedMB)
	}
	if !lane.IsOOM() {
		t.Errorf("expected IsOOM() to be true at 100%% capacity")
	}

	// Over-allocation beyond capacity
	ok = lane.Allocate(20)
	if ok {
		t.Errorf("expected over-capacity allocation to return false")
	}
	if lane.CurrentAllocatedMB != 120 {
		t.Errorf("expected 120MB, got %d", lane.CurrentAllocatedMB)
	}
	if lane.UsageRatio() < 1.0 {
		t.Errorf("expected usage ratio >= 1.0, got %f", lane.UsageRatio())
	}

	// Deallocation
	lane.Deallocate(30)
	if lane.CurrentAllocatedMB != 90 {
		t.Errorf("expected 90MB after deallocating 30MB, got %d", lane.CurrentAllocatedMB)
	}
	if lane.IsOOM() {
		t.Errorf("expected lane to no longer be in OOM at 90/100")
	}

	// Non-positive deallocation
	lane.Deallocate(0)
	lane.Deallocate(-10)
	if lane.CurrentAllocatedMB != 90 {
		t.Errorf("deallocate non-positive should be no-op")
	}

	// Clamp at zero deallocation
	lane.Deallocate(200)
	if lane.CurrentAllocatedMB != 0 {
		t.Errorf("expected clamping at 0MB, got %d", lane.CurrentAllocatedMB)
	}

	// Reset
	lane.Allocate(80)
	lane.Reset()
	if lane.CurrentAllocatedMB != 0 {
		t.Errorf("expected 0MB after Reset(), got %d", lane.CurrentAllocatedMB)
	}
}

func TestUsageRatioZeroCapacity(t *testing.T) {
	lane := NewMemoryLane(LaneConfig{
		GenType:    GenEden,
		CapacityMB: 0,
	})
	if lane.UsageRatio() != 0.0 {
		t.Errorf("expected 0.0 usage ratio for 0 capacity, got %f", lane.UsageRatio())
	}
}

func TestDefaultLanes(t *testing.T) {
	lanes := NewDefaultLanes()

	if len(lanes) != 4 {
		t.Fatalf("expected 4 lanes, got %d", len(lanes))
	}

	// Eden
	eden, ok := lanes[GenEden]
	if !ok || eden.Config.CapacityMB != 256 || eden.Config.ColorHex != 0x00E676 {
		t.Errorf("unexpected Eden configuration: %+v", eden)
	}

	// Survivor 0
	s0, ok := lanes[GenSurvivor0]
	if !ok || s0.Config.CapacityMB != 64 || s0.Config.ColorHex != 0xFFD600 {
		t.Errorf("unexpected Survivor0 configuration: %+v", s0)
	}

	// Survivor 1
	s1, ok := lanes[GenSurvivor1]
	if !ok || s1.Config.CapacityMB != 64 || s1.Config.ColorHex != 0xFFD600 {
		t.Errorf("unexpected Survivor1 configuration: %+v", s1)
	}

	// Tenured
	tenured, ok := lanes[GenTenured]
	if !ok || tenured.Config.CapacityMB != 512 || tenured.Config.ColorHex != 0xFF1744 {
		t.Errorf("unexpected Tenured configuration: %+v", tenured)
	}
}
