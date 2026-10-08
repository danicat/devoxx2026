package entity_test

import (
	"math"
	"testing"

	"heapdefender/internal/entity"
)

func TestNewCollector_BaseStats(t *testing.T) {
	tests := []struct {
		colType         entity.CollectorType
		expectedCost    int
		expectedRange   float64
		expectedFireRate float64
		expectedDPS     float64
		expectedUpgrade int
	}{
		{entity.ColSerial, 50, 120.0, 2.0, 30.0, 40},
		{entity.ColParallel, 120, 100.0, 4.0, 60.0, 90},
		{entity.ColCMS, 200, 140.0, 1.5, 45.0, 150},
		{entity.ColG1, 350, 160.0, 2.5, 90.0, 250},
		{entity.ColZGC, 500, 180.0, 5.0, 150.0, 350},
	}

	for _, tc := range tests {
		c := entity.NewCollector(1, tc.colType, 5, 5, 150.0, 150.0)
		if c.ID != 1 {
			t.Errorf("expected ID 1, got %d", c.ID)
		}
		if c.Type != tc.colType {
			t.Errorf("expected Type %d, got %d", tc.colType, c.Type)
		}
		if c.Cost != tc.expectedCost {
			t.Errorf("type %d: expected Cost %d, got %d", tc.colType, tc.expectedCost, c.Cost)
		}
		if c.Range != tc.expectedRange {
			t.Errorf("type %d: expected Range %f, got %f", tc.colType, tc.expectedRange, c.Range)
		}
		if c.FireRate != tc.expectedFireRate {
			t.Errorf("type %d: expected FireRate %f, got %f", tc.colType, tc.expectedFireRate, c.FireRate)
		}
		if c.DPS != tc.expectedDPS {
			t.Errorf("type %d: expected DPS %f, got %f", tc.colType, tc.expectedDPS, c.DPS)
		}
		if c.UpgradeCost != tc.expectedUpgrade {
			t.Errorf("type %d: expected UpgradeCost %d, got %d", tc.colType, tc.expectedUpgrade, c.UpgradeCost)
		}
		if c.Level != 1 {
			t.Errorf("expected Level 1, got %d", c.Level)
		}
		if c.Cooldown != 0.0 {
			t.Errorf("expected initial Cooldown 0.0, got %f", c.Cooldown)
		}
	}
}

func TestNewCollector_UnknownPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unknown CollectorType")
		}
	}()

	_ = entity.NewCollector(99, entity.CollectorType(999), 0, 0, 0, 0)
}

func TestCollector_Upgrade(t *testing.T) {
	c := entity.NewCollector(1, entity.ColSerial, 0, 0, 100, 100)
	initialDPS := c.DPS
	initialRange := c.Range
	initialCost := c.UpgradeCost

	// Level 1 -> 2
	if !c.CanUpgrade(100) {
		t.Errorf("expected CanUpgrade true with 100 cycles (cost %d)", initialCost)
	}
	if c.CanUpgrade(20) {
		t.Errorf("expected CanUpgrade false with insufficient cycles 20 (cost %d)", initialCost)
	}

	err := c.Upgrade()
	if err != nil {
		t.Fatalf("unexpected error on upgrade: %v", err)
	}

	if c.Level != 2 {
		t.Errorf("expected Level 2, got %d", c.Level)
	}
	expectedDPS := initialDPS * 1.25
	if math.Abs(c.DPS-expectedDPS) > 1e-6 {
		t.Errorf("expected DPS %f, got %f", expectedDPS, c.DPS)
	}
	expectedRange := initialRange * 1.15
	if math.Abs(c.Range-expectedRange) > 1e-6 {
		t.Errorf("expected Range %f, got %f", expectedRange, c.Range)
	}

	// Upgrade until level 5 (max level)
	for c.Level < entity.MaxCollectorLevel {
		if err := c.Upgrade(); err != nil {
			t.Fatalf("failed to upgrade to level %d: %v", c.Level+1, err)
		}
	}

	if c.Level != entity.MaxCollectorLevel {
		t.Fatalf("expected level %d, got %d", entity.MaxCollectorLevel, c.Level)
	}

	if c.CanUpgrade(999999) {
		t.Errorf("expected CanUpgrade false when already at max level")
	}

	err = c.Upgrade()
	if err == nil {
		t.Fatalf("expected error upgrading past max level, got nil")
	}
}

func TestCollector_FindTarget_Serial(t *testing.T) {
	c := entity.NewCollector(1, entity.ColSerial, 0, 0, 0, 0) // Range 120

	obj1 := entity.NewAllocationObject(1, entity.ObjLambda, 50, 0)
	obj1.DistanceTraveled = 50.0

	obj2 := entity.NewAllocationObject(2, entity.ObjLambda, 80, 0)
	obj2.DistanceTraveled = 120.0 // Further progressed

	obj3 := entity.NewAllocationObject(3, entity.ObjLambda, 150, 0) // Out of range (>120)
	obj3.DistanceTraveled = 200.0

	objDead := entity.NewAllocationObject(4, entity.ObjLambda, 30, 0)
	objDead.DistanceTraveled = 300.0
	objDead.IsDead = true

	objects := []*entity.AllocationObject{obj1, obj2, obj3, objDead}
	target := c.FindTarget(objects)
	if target == nil {
		t.Fatalf("expected target, got nil")
	}
	if target.ID != 2 {
		t.Errorf("expected target ID 2 (greatest DistanceTraveled in range), got %d", target.ID)
	}
}

func TestCollector_FindTarget_Parallel(t *testing.T) {
	c := entity.NewCollector(1, entity.ColParallel, 0, 0, 0, 0) // Range 100

	objFar := entity.NewAllocationObject(1, entity.ObjLambda, 80, 0) // dist 80
	objNear := entity.NewAllocationObject(2, entity.ObjLambda, 20, 0) // dist 20
	objOut := entity.NewAllocationObject(3, entity.ObjLambda, 120, 0) // dist 120 (out)

	objects := []*entity.AllocationObject{objFar, objNear, objOut}
	target := c.FindTarget(objects)
	if target == nil {
		t.Fatalf("expected target, got nil")
	}
	if target.ID != 2 {
		t.Errorf("expected nearest target ID 2, got %d", target.ID)
	}
}

func TestCollector_FindTarget_G1(t *testing.T) {
	c := entity.NewCollector(1, entity.ColG1, 0, 0, 0, 0) // Range 160

	objLight := entity.NewAllocationObject(1, entity.ObjLambda, 50, 0) // HP 25
	objDense := entity.NewAllocationObject(2, entity.ObjLargeBlob, 60, 0) // HP 600

	objects := []*entity.AllocationObject{objLight, objDense}
	target := c.FindTarget(objects)
	if target == nil {
		t.Fatalf("expected target, got nil")
	}
	if target.ID != 2 {
		t.Errorf("expected densest target ID 2, got %d", target.ID)
	}
}

func TestCollector_FindTarget_ZGC(t *testing.T) {
	c := entity.NewCollector(1, entity.ColZGC, 0, 0, 0, 0) // Range 180

	objYoung := entity.NewAllocationObject(1, entity.ObjLambda, 50, 0)
	objYoung.TenureAge = 0

	objOld := entity.NewAllocationObject(2, entity.ObjString, 60, 0)
	objOld.TenureAge = 4
	objOld.Age() // Age >= 4 promotes to GenTenured

	objects := []*entity.AllocationObject{objYoung, objOld}
	target := c.FindTarget(objects)
	if target == nil {
		t.Fatalf("expected target, got nil")
	}
	if target.ID != 2 {
		t.Errorf("expected oldest tenure target ID 2, got %d", target.ID)
	}
}

func TestCollector_FindTarget_NoneInRange(t *testing.T) {
	c := entity.NewCollector(1, entity.ColSerial, 0, 0, 0, 0)
	objOut := entity.NewAllocationObject(1, entity.ObjLambda, 500, 500)

	target := c.FindTarget([]*entity.AllocationObject{objOut})
	if target != nil {
		t.Errorf("expected nil target for out of range object, got %d", target.ID)
	}
}

func TestCollector_Update_FiringAndCooldown(t *testing.T) {
	c := entity.NewCollector(1, entity.ColSerial, 0, 0, 0, 0) // FireRate 2.0 (cooldown 0.5s), DPS 30 -> 15 dmg per shot
	obj := entity.NewAllocationObject(1, entity.ObjLambda, 50, 0) // 25 HP, unarmored

	// 1st update: fires
	projs := c.Update(0.016, []*entity.AllocationObject{obj}, false)
	if len(projs) != 1 {
		t.Fatalf("expected 1 projectile, got %d", len(projs))
	}
	expectedHP := 25.0 - 15.0
	if math.Abs(obj.Health-expectedHP) > 1e-6 {
		t.Errorf("expected target Health %f, got %f", expectedHP, obj.Health)
	}
	if c.Cooldown <= 0 {
		t.Errorf("expected Cooldown to be set, got %f", c.Cooldown)
	}

	// Immediate next update: still on cooldown
	projs2 := c.Update(0.1, []*entity.AllocationObject{obj}, false)
	if len(projs2) != 0 {
		t.Errorf("expected 0 projectiles during cooldown, got %d", len(projs2))
	}

	// When remaining cooldown is exhausted, the collector fires immediately upon becoming ready!
	projs3 := c.Update(0.5, []*entity.AllocationObject{obj}, false)
	if len(projs3) != 1 {
		t.Fatalf("expected 1 projectile after cooldown, got %d", len(projs3))
	}
	if !obj.IsDead {
		t.Errorf("expected target to be dead after lethal damage")
	}
}

func TestCollector_Update_STWBonus(t *testing.T) {
	c := entity.NewCollector(1, entity.ColSerial, 0, 0, 0, 0) // FireRate 2.0, DPS 30 -> base 15 dmg, 2x STW = 30 dmg
	obj := entity.NewAllocationObject(1, entity.ObjString, 50, 0) // 60 HP, unarmored

	projs := c.Update(0.016, []*entity.AllocationObject{obj}, true)
	if len(projs) != 1 {
		t.Fatalf("expected 1 projectile, got %d", len(projs))
	}
	// 60 - 30 = 30 HP
	expectedHP := 30.0
	if math.Abs(obj.Health-expectedHP) > 1e-6 {
		t.Errorf("expected target Health %f under STW bonus, got %f", expectedHP, obj.Health)
	}
}

func TestCollector_Update_CMS_RadialAOE(t *testing.T) {
	c := entity.NewCollector(1, entity.ColCMS, 0, 0, 0, 0) // Range 140, FireRate 1.5, DPS 45 -> 30 dmg per shot
	obj1 := entity.NewAllocationObject(1, entity.ObjLambda, 50, 0)  // 25 HP -> dead
	obj2 := entity.NewAllocationObject(2, entity.ObjString, 80, 0)  // 60 HP -> 30 HP
	obj3 := entity.NewAllocationObject(3, entity.ObjLambda, 200, 0) // Out of range

	projs := c.Update(0.016, []*entity.AllocationObject{obj1, obj2, obj3}, false)
	if len(projs) != 2 {
		t.Fatalf("expected 2 projectiles for 2 in-range targets, got %d", len(projs))
	}

	if !obj1.IsDead {
		t.Errorf("expected obj1 to be dead (25 HP - 30 dmg)")
	}
	if math.Abs(obj2.Health-30.0) > 1e-6 {
		t.Errorf("expected obj2 Health to be 30.0, got %f", obj2.Health)
	}
	if obj3.Health != 25.0 {
		t.Errorf("expected out-of-range obj3 to take no damage, got Health %f", obj3.Health)
	}
}
