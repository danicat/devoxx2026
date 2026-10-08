package entity

import (
	"math"
	"testing"
)

func TestNewAllocationObject(t *testing.T) {
	tests := []struct {
		name      string
		objType   ObjectType
		expSizeMB int
		expMaxHP  float64
		expSpeed  float64
		expArmor  float64
	}{
		{
			name:      "Lambda",
			objType:   ObjLambda,
			expSizeMB: 5,
			expMaxHP:  25.0,
			expSpeed:  2.2,
			expArmor:  0.0,
		},
		{
			name:      "String",
			objType:   ObjString,
			expSizeMB: 20,
			expMaxHP:  60.0,
			expSpeed:  1.4,
			expArmor:  0.0,
		},
		{
			name:      "ThreadLocal",
			objType:   ObjThreadLocal,
			expSizeMB: 50,
			expMaxHP:  150.0,
			expSpeed:  0.9,
			expArmor:  0.35,
		},
		{
			name:      "LargeBlob",
			objType:   ObjLargeBlob,
			expSizeMB: 200,
			expMaxHP:  600.0,
			expSpeed:  0.5,
			expArmor:  0.10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := NewAllocationObject(101, tc.objType, 10.0, 20.0)
			if obj == nil {
				t.Fatalf("expected non-nil AllocationObject")
			}
			if obj.ID != 101 {
				t.Errorf("expected ID 101, got %d", obj.ID)
			}
			if obj.Type != tc.objType {
				t.Errorf("expected Type %v, got %v", tc.objType, obj.Type)
			}
			if obj.SizeMB != tc.expSizeMB {
				t.Errorf("expected SizeMB %d, got %d", tc.expSizeMB, obj.SizeMB)
			}
			if math.Abs(obj.MaxHealth-tc.expMaxHP) > 1e-6 {
				t.Errorf("expected MaxHealth %f, got %f", tc.expMaxHP, obj.MaxHealth)
			}
			if math.Abs(obj.Health-tc.expMaxHP) > 1e-6 {
				t.Errorf("expected initial Health %f, got %f", tc.expMaxHP, obj.Health)
			}
			if math.Abs(obj.Speed-tc.expSpeed) > 1e-6 {
				t.Errorf("expected Speed %f, got %f", tc.expSpeed, obj.Speed)
			}
			if math.Abs(obj.Armor-tc.expArmor) > 1e-6 {
				t.Errorf("expected Armor %f, got %f", tc.expArmor, obj.Armor)
			}
			if obj.X != 10.0 || obj.Y != 20.0 {
				t.Errorf("expected pos (10, 20), got (%f, %f)", obj.X, obj.Y)
			}
			if obj.TenureAge != 0 {
				t.Errorf("expected TenureAge 0, got %d", obj.TenureAge)
			}
			if obj.Generation != GenEden {
				t.Errorf("expected Generation %d (GenEden), got %d", GenEden, obj.Generation)
			}
			if obj.IsDead {
				t.Errorf("expected IsDead false")
			}
			if obj.ReachedEnd {
				t.Errorf("expected ReachedEnd false")
			}
			if obj.PathIndex != 0 {
				t.Errorf("expected PathIndex 0, got %d", obj.PathIndex)
			}
			if obj.DistanceTraveled != 0 {
				t.Errorf("expected DistanceTraveled 0, got %f", obj.DistanceTraveled)
			}
		})
	}
}

func TestNewAllocationObject_UnknownPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on unknown ObjectType")
		}
	}()
	_ = NewAllocationObject(999, ObjectType(99), 0, 0)
}

func TestTakeDamage(t *testing.T) {
	t.Run("unarmored damage", func(t *testing.T) {
		obj := NewAllocationObject(1, ObjLambda, 0, 0) // MaxHP: 25, Armor: 0.0
		obj.TakeDamage(10.0)
		if math.Abs(obj.Health-15.0) > 1e-6 {
			t.Errorf("expected HP 15.0, got %f", obj.Health)
		}
		if obj.IsDead {
			t.Errorf("expected IsDead false")
		}

		// Negative damage does nothing
		obj.TakeDamage(-5.0)
		if math.Abs(obj.Health-15.0) > 1e-6 {
			t.Errorf("expected HP 15.0 after negative damage, got %f", obj.Health)
		}

		// Zero damage does nothing
		obj.TakeDamage(0.0)
		if math.Abs(obj.Health-15.0) > 1e-6 {
			t.Errorf("expected HP 15.0 after zero damage, got %f", obj.Health)
		}
	})

	t.Run("armored damage reduction", func(t *testing.T) {
		obj := NewAllocationObject(2, ObjThreadLocal, 0, 0) // MaxHP: 150, Armor: 0.35
		// Effective damage: 100 * (1 - 0.35) = 65.0
		obj.TakeDamage(100.0)
		expectedHP := 150.0 - 65.0
		if math.Abs(obj.Health-expectedHP) > 1e-6 {
			t.Errorf("expected HP %f, got %f", expectedHP, obj.Health)
		}
		if obj.IsDead {
			t.Errorf("expected IsDead false")
		}
	})

	t.Run("lethal damage", func(t *testing.T) {
		obj := NewAllocationObject(3, ObjString, 0, 0) // MaxHP: 60, Armor: 0.0
		obj.TakeDamage(60.0)
		if obj.Health != 0 {
			t.Errorf("expected HP 0, got %f", obj.Health)
		}
		if !obj.IsDead {
			t.Errorf("expected IsDead true")
		}

		// Extra damage after death should be ignored
		obj.TakeDamage(10.0)
		if obj.Health != 0 {
			t.Errorf("expected HP to stay 0, got %f", obj.Health)
		}
	})

	t.Run("overkill damage clamps to zero", func(t *testing.T) {
		obj := NewAllocationObject(4, ObjLambda, 0, 0) // MaxHP: 25, Armor: 0.0
		obj.TakeDamage(100.0)
		if obj.Health != 0 {
			t.Errorf("expected HP clamped to 0, got %f", obj.Health)
		}
		if !obj.IsDead {
			t.Errorf("expected IsDead true")
		}
	})
}

func TestAgePromotion(t *testing.T) {
	obj := NewAllocationObject(1, ObjString, 0, 0)

	if obj.TenureAge != 0 || obj.Generation != GenEden {
		t.Fatalf("expected age 0 and GenEden, got age %d gen %d", obj.TenureAge, obj.Generation)
	}

	// 1st cycle -> Survivor 0
	obj.Age()
	if obj.TenureAge != 1 || obj.Generation != GenSurvivor0 {
		t.Errorf("after age 1: expected age 1 and GenSurvivor0, got age %d gen %d", obj.TenureAge, obj.Generation)
	}

	// 2nd cycle -> Survivor 1
	obj.Age()
	if obj.TenureAge != 2 || obj.Generation != GenSurvivor1 {
		t.Errorf("after age 2: expected age 2 and GenSurvivor1, got age %d gen %d", obj.TenureAge, obj.Generation)
	}

	// 3rd cycle -> remains Survivor 1 until threshold 4
	obj.Age()
	if obj.TenureAge != 3 || obj.Generation != GenSurvivor1 {
		t.Errorf("after age 3: expected age 3 and GenSurvivor1, got age %d gen %d", obj.TenureAge, obj.Generation)
	}

	// 4th cycle -> Tenured
	obj.Age()
	if obj.TenureAge != 4 || obj.Generation != GenTenured {
		t.Errorf("after age 4: expected age 4 and GenTenured, got age %d gen %d", obj.TenureAge, obj.Generation)
	}

	// 5th cycle -> stays Tenured
	obj.Age()
	if obj.TenureAge != 5 || obj.Generation != GenTenured {
		t.Errorf("after age 5: expected age 5 and GenTenured, got age %d gen %d", obj.TenureAge, obj.Generation)
	}
}

func TestMoveTowards(t *testing.T) {
	t.Run("partial step towards target", func(t *testing.T) {
		obj := NewAllocationObject(1, ObjLambda, 0.0, 0.0)
		reached := obj.MoveTowards(10.0, 0.0, 4.0)
		if reached {
			t.Errorf("expected reached false")
		}
		if math.Abs(obj.X-4.0) > 1e-6 || math.Abs(obj.Y-0.0) > 1e-6 {
			t.Errorf("expected pos (4.0, 0.0), got (%f, %f)", obj.X, obj.Y)
		}
		if math.Abs(obj.DistanceTraveled-4.0) > 1e-6 {
			t.Errorf("expected DistanceTraveled 4.0, got %f", obj.DistanceTraveled)
		}
	})

	t.Run("reaching target exactly or overshooting step", func(t *testing.T) {
		obj := NewAllocationObject(2, ObjLambda, 4.0, 0.0)
		reached := obj.MoveTowards(10.0, 0.0, 10.0)
		if !reached {
			t.Errorf("expected reached true")
		}
		if math.Abs(obj.X-10.0) > 1e-6 || math.Abs(obj.Y-0.0) > 1e-6 {
			t.Errorf("expected pos (10.0, 0.0), got (%f, %f)", obj.X, obj.Y)
		}
		if math.Abs(obj.DistanceTraveled-6.0) > 1e-6 {
			t.Errorf("expected DistanceTraveled 6.0, got %f", obj.DistanceTraveled)
		}
	})

	t.Run("zero or negative step does not advance", func(t *testing.T) {
		obj := NewAllocationObject(3, ObjLambda, 5.0, 5.0)
		reached := obj.MoveTowards(10.0, 10.0, 0.0)
		if reached {
			t.Errorf("expected reached false")
		}
		if obj.X != 5.0 || obj.Y != 5.0 || obj.DistanceTraveled != 0 {
			t.Errorf("expected no movement on step 0")
		}

		reached = obj.MoveTowards(10.0, 10.0, -2.0)
		if reached {
			t.Errorf("expected reached false on negative step")
		}
	})
}

func TestUpdateWithPath(t *testing.T) {
	t.Run("navigates through waypoints and reaches end", func(t *testing.T) {
		waypoints := [][2]float64{
			{0.0, 0.0},
			{10.0, 0.0},
			{10.0, 10.0},
		}

		// Use an object with speed 5.0 for predictable steps
		obj := NewAllocationObject(1, ObjLambda, 0.0, 0.0)
		obj.Speed = 5.0

		// Initial state
		if obj.ReachedEnd {
			t.Fatalf("expected ReachedEnd false")
		}

		// Frame 1: starts at (0,0), target waypoints[0]=(0,0) is dist 0, proceeds to waypoints[1]=(10,0) with step 5
		obj.UpdateWithPath(waypoints)
		if math.Abs(obj.X-5.0) > 1e-6 || math.Abs(obj.Y-0.0) > 1e-6 {
			t.Errorf("frame 1: expected pos (5.0, 0.0), got (%f, %f)", obj.X, obj.Y)
		}
		if obj.PathIndex != 1 {
			t.Errorf("frame 1: expected PathIndex 1, got %d", obj.PathIndex)
		}
		if obj.ReachedEnd {
			t.Errorf("frame 1: expected ReachedEnd false")
		}

		// Frame 2: moves remaining 5 to waypoints[1]=(10,0)
		obj.UpdateWithPath(waypoints)
		if math.Abs(obj.X-10.0) > 1e-6 || math.Abs(obj.Y-0.0) > 1e-6 {
			t.Errorf("frame 2: expected pos (10.0, 0.0), got (%f, %f)", obj.X, obj.Y)
		}
		if obj.PathIndex != 2 {
			t.Errorf("frame 2: expected PathIndex 2, got %d", obj.PathIndex)
		}
		if obj.ReachedEnd {
			t.Errorf("frame 2: expected ReachedEnd false")
		}

		// Frame 3: moves 5 towards waypoints[2]=(10,10)
		obj.UpdateWithPath(waypoints)
		if math.Abs(obj.X-10.0) > 1e-6 || math.Abs(obj.Y-5.0) > 1e-6 {
			t.Errorf("frame 3: expected pos (10.0, 5.0), got (%f, %f)", obj.X, obj.Y)
		}

		// Frame 4: moves 5 to waypoints[2]=(10,10) and reaches end
		obj.UpdateWithPath(waypoints)
		if math.Abs(obj.X-10.0) > 1e-6 || math.Abs(obj.Y-10.0) > 1e-6 {
			t.Errorf("frame 4: expected pos (10.0, 10.0), got (%f, %f)", obj.X, obj.Y)
		}
		if !obj.ReachedEnd {
			t.Errorf("frame 4: expected ReachedEnd true")
		}
		if math.Abs(obj.DistanceTraveled-20.0) > 1e-6 {
			t.Errorf("expected total DistanceTraveled 20.0, got %f", obj.DistanceTraveled)
		}

		// Calling UpdateWithPath after ReachedEnd does nothing
		obj.UpdateWithPath(waypoints)
		if math.Abs(obj.DistanceTraveled-20.0) > 1e-6 {
			t.Errorf("expected DistanceTraveled not to change after reaching end")
		}
	})

	t.Run("empty waypoints does not panic", func(t *testing.T) {
		obj := NewAllocationObject(1, ObjLambda, 0, 0)
		obj.UpdateWithPath(nil)
		if obj.ReachedEnd {
			t.Errorf("expected ReachedEnd false")
		}
	})

	t.Run("dead object does not move", func(t *testing.T) {
		obj := NewAllocationObject(1, ObjLambda, 0, 0)
		obj.IsDead = true
		waypoints := [][2]float64{{10.0, 10.0}}
		obj.UpdateWithPath(waypoints)
		if obj.X != 0 || obj.Y != 0 || obj.DistanceTraveled != 0 {
			t.Errorf("dead object should not move")
		}
	})
}
