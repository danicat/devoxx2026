package art

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestGenerateAssets(t *testing.T) {
	assets, err := GenerateAssets()
	if err != nil {
		t.Fatalf("GenerateAssets() returned error: %v", err)
	}
	if assets == nil {
		t.Fatalf("GenerateAssets() returned nil assets")
	}

	// Verify SiliconTile
	if assets.SiliconTile == nil {
		t.Errorf("SiliconTile is nil")
	} else {
		w, h := assets.SiliconTile.Bounds().Dx(), assets.SiliconTile.Bounds().Dy()
		if w != 30 || h != 30 {
			t.Errorf("SiliconTile bounds = (%d, %d), expected (30, 30)", w, h)
		}
	}

	// Verify BusTrace
	if assets.BusTrace == nil {
		t.Errorf("BusTrace is nil")
	} else {
		w, h := assets.BusTrace.Bounds().Dx(), assets.BusTrace.Bounds().Dy()
		if w != 30 || h != 30 {
			t.Errorf("BusTrace bounds = (%d, %d), expected (30, 30)", w, h)
		}
	}

	// Verify CollectorTurret map (keys 0 to 4)
	for i := 0; i <= 4; i++ {
		img, ok := assets.CollectorTurret[i]
		if !ok || img == nil {
			t.Errorf("CollectorTurret[%d] is missing or nil", i)
		} else {
			w, h := img.Bounds().Dx(), img.Bounds().Dy()
			if w != 30 || h != 30 {
				t.Errorf("CollectorTurret[%d] bounds = (%d, %d), expected (30, 30)", i, w, h)
			}
		}
	}

	// Verify ObjectSprites map (keys 0 to 3)
	for i := 0; i <= 3; i++ {
		img, ok := assets.ObjectSprites[i]
		if !ok || img == nil {
			t.Errorf("ObjectSprites[%d] is missing or nil", i)
		} else {
			w, h := img.Bounds().Dx(), img.Bounds().Dy()
			if w != 30 || h != 30 {
				t.Errorf("ObjectSprites[%d] bounds = (%d, %d), expected (30, 30)", i, w, h)
			}
		}
	}
}

func TestVectorHelpers(t *testing.T) {
	dst := ebiten.NewImage(64, 64)
	if dst == nil {
		t.Fatalf("failed to create destination test image")
	}

	// Test DrawLaserBeam
	DrawLaserBeam(dst, 5, 5, 50, 50, color.White, color.RGBA{R: 0, G: 255, B: 255, A: 128}, 2, 6)

	// Test DrawTargetLockReticle
	DrawTargetLockReticle(dst, 32, 32, 12, 0.5, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Test DrawRadialSweepRing
	DrawRadialSweepRing(dst, 32, 32, 10, 25, color.RGBA{R: 180, G: 0, B: 255, A: 200})

	// Test DrawHealthBar
	DrawHealthBar(dst, 10, 10, 40, 6, 75, 100)
	DrawHealthBar(dst, 10, 20, 40, 6, 25, 100)
	DrawHealthBar(dst, 10, 30, 40, 6, 5, 100)

	// Test DrawMemoryBadge
	DrawMemoryBadge(dst, 5, 45, 64, 2)

	// Test DrawGridCellHighlight
	DrawGridCellHighlight(dst, 0, 0, 30, true)
	DrawGridCellHighlight(dst, 1, 1, 30, false)

	// Test DrawRangeCircle
	DrawRangeCircle(dst, 32, 32, 20, color.RGBA{R: 0, G: 200, B: 255, A: 100})
}
