package entity_test

import (
	"testing"

	"heapdefender/internal/entity"
)

func TestNewProjectile(t *testing.T) {
	proj := entity.NewProjectile(101, entity.ProjBeam, 10.0, 20.0, 50.0, 60.0, 42, 0.25, 0x00E676)
	if proj == nil {
		t.Fatal("expected non-nil Projectile")
	}

	if proj.ID != 101 {
		t.Errorf("expected ID 101, got %d", proj.ID)
	}
	if proj.Type != entity.ProjBeam {
		t.Errorf("expected ProjBeam (%d), got %d", entity.ProjBeam, proj.Type)
	}
	if proj.StartX != 10.0 || proj.StartY != 20.0 {
		t.Errorf("expected Start (10, 20), got (%f, %f)", proj.StartX, proj.StartY)
	}
	if proj.TargetX != 50.0 || proj.TargetY != 60.0 {
		t.Errorf("expected Target (50, 60), got (%f, %f)", proj.TargetX, proj.TargetY)
	}
	if proj.TargetID != 42 {
		t.Errorf("expected TargetID 42, got %d", proj.TargetID)
	}
	if proj.Duration != 0.25 {
		t.Errorf("expected Duration 0.25, got %f", proj.Duration)
	}
	if proj.Elapsed != 0.0 {
		t.Errorf("expected initial Elapsed 0, got %f", proj.Elapsed)
	}
	if proj.IsExpired {
		t.Errorf("expected initial IsExpired false, got true")
	}
	if proj.ColorHex != 0x00E676 {
		t.Errorf("expected ColorHex 0x00E676, got 0x%X", proj.ColorHex)
	}
}

func TestProjectile_Update(t *testing.T) {
	proj := entity.NewProjectile(1, entity.ProjBurst, 0, 0, 10, 10, 5, 0.2, 0xFFD600)

	// Step 1: partial update
	proj.Update(0.1)
	if proj.Elapsed != 0.1 {
		t.Errorf("expected Elapsed 0.1, got %f", proj.Elapsed)
	}
	if proj.IsExpired {
		t.Errorf("expected IsExpired false after partial update")
	}

	// Step 2: update to reach duration
	proj.Update(0.1)
	if proj.Elapsed != 0.2 {
		t.Errorf("expected Elapsed 0.2, got %f", proj.Elapsed)
	}
	if !proj.IsExpired {
		t.Errorf("expected IsExpired true when Elapsed >= Duration")
	}

	// Step 3: further update after expiration should not modify state
	proj.Update(0.1)
	if proj.Elapsed != 0.2 {
		t.Errorf("expected Elapsed to remain 0.2 once expired, got %f", proj.Elapsed)
	}
}
