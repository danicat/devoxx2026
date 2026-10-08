package art

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawLaserBeam renders a laser beam between (x1, y1) and (x2, y2) with an outer glow and bright inner core.
func DrawLaserBeam(dst *ebiten.Image, x1, y1, x2, y2 float32, coreColor color.Color, glowColor color.Color, coreWidth, glowWidth float32) {
	if glowWidth > 0 {
		vector.StrokeLine(dst, x1, y1, x2, y2, glowWidth, glowColor, true)
	}
	if coreWidth > 0 {
		vector.StrokeLine(dst, x1, y1, x2, y2, coreWidth, coreColor, true)
	}
}

// DrawTargetLockReticle renders an animated or stationary target lock reticle around (cx, cy).
// angleRad specifies the rotation angle for rotating brackets/crosshairs.
func DrawTargetLockReticle(dst *ebiten.Image, cx, cy, radius float32, angleRad float64, clr color.Color) {
	// Draw 4 corner bracket reticles
	bracketLen := radius * 0.45
	for i := 0; i < 4; i++ {
		theta := angleRad + float64(i)*math.Pi/2
		// Corner position
		px := cx + radius*float32(math.Cos(theta))
		py := cy + radius*float32(math.Sin(theta))

		// Normal vectors for brackets
		tangentTheta := theta + math.Pi/2
		tx := px + bracketLen*float32(math.Cos(tangentTheta))
		ty := py + bracketLen*float32(math.Sin(tangentTheta))

		vector.StrokeLine(dst, px, py, tx, ty, 1.5, clr, true)
	}

	// Inner dot
	vector.DrawFilledCircle(dst, cx, cy, 1.5, clr, true)
}

// DrawRadialSweepRing renders a pulsing radial sweep ring (e.g. for CMS pulse or collector range).
func DrawRadialSweepRing(dst *ebiten.Image, cx, cy, currentRadius, maxRadius float32, ringColor color.RGBA) {
	if currentRadius <= 0 || maxRadius <= 0 {
		return
	}
	// Fade alpha as radius reaches max
	ratio := 1.0 - (currentRadius / maxRadius)
	if ratio < 0 {
		ratio = 0
	}
	alpha := float32(ringColor.A) * ratio
	fadedColor := color.RGBA{
		R: ringColor.R,
		G: ringColor.G,
		B: ringColor.B,
		A: uint8(alpha),
	}

	vector.StrokeCircle(dst, cx, cy, currentRadius, 2.0, fadedColor, true)
}

// DrawHealthBar renders an object's dynamic health bar with background, border, and colored progress fill.
func DrawHealthBar(dst *ebiten.Image, x, y, width, height float32, health, maxHealth float64) {
	if maxHealth <= 0 {
		return
	}
	ratio := float32(health / maxHealth)
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}

	// Background
	vector.DrawFilledRect(dst, x, y, width, height, color.RGBA{R: 20, G: 20, B: 25, A: 200}, true)

	// Health fill color: Green -> Yellow -> Red gradient
	var fillClr color.RGBA
	if ratio > 0.6 {
		fillClr = color.RGBA{R: 0, G: 230, B: 118, A: 230} // Green
	} else if ratio > 0.25 {
		fillClr = color.RGBA{R: 255, G: 214, B: 0, A: 230} // Yellow/Amber
	} else {
		fillClr = color.RGBA{R: 255, G: 23, B: 68, A: 240} // Red
	}

	fillWidth := width * ratio
	if fillWidth > 0 {
		vector.DrawFilledRect(dst, x, y, fillWidth, height, fillClr, true)
	}

	// Border
	vector.StrokeRect(dst, x, y, width, height, 1, color.RGBA{R: 80, G: 90, B: 100, A: 180}, true)
}

// DrawMemoryBadge renders a small byte size / allocation badge tag (e.g. "64M", "256M").
func DrawMemoryBadge(dst *ebiten.Image, x, y float32, sizeMB int, tenureAge int) {
	// Badge background
	w := float32(22)
	h := float32(9)

	badgeBg := color.RGBA{R: 15, G: 25, B: 40, A: 220}
	borderClr := color.RGBA{R: 0, G: 200, B: 255, A: 200}
	if tenureAge > 0 {
		// Survivor / Tenured border
		borderClr = color.RGBA{R: 255, G: 195, B: 0, A: 220}
	}
	if tenureAge >= 3 {
		borderClr = color.RGBA{R: 255, G: 50, B: 80, A: 240}
	}

	vector.DrawFilledRect(dst, x, y, w, h, badgeBg, true)
	vector.StrokeRect(dst, x, y, w, h, 1, borderClr, true)

	// Mini age pip indicator dots
	for i := 0; i < tenureAge && i < 4; i++ {
		px := x + 3 + float32(i)*4.5
		py := y + 4.5
		vector.DrawFilledCircle(dst, px, py, 1.2, borderClr, true)
	}
}

// DrawGridCellHighlight renders selection or hover highlights on a grid cell.
func DrawGridCellHighlight(dst *ebiten.Image, gridX, gridY int, cellSize float32, isValid bool) {
	x := float32(gridX) * cellSize
	y := float32(gridY) * cellSize

	var fillClr color.RGBA
	var strokeClr color.RGBA

	if isValid {
		fillClr = color.RGBA{R: 0, G: 230, B: 118, A: 40}
		strokeClr = color.RGBA{R: 0, G: 230, B: 118, A: 200}
	} else {
		fillClr = color.RGBA{R: 255, G: 23, B: 68, A: 50}
		strokeClr = color.RGBA{R: 255, G: 23, B: 68, A: 220}
	}

	vector.DrawFilledRect(dst, x, y, cellSize, cellSize, fillClr, true)
	vector.StrokeRect(dst, x+0.5, y+0.5, cellSize-1, cellSize-1, 1.5, strokeClr, true)

	// Corner notches
	notch := cellSize * 0.25
	vector.StrokeLine(dst, x, y, x+notch, y, 2, strokeClr, true)
	vector.StrokeLine(dst, x, y, x, y+notch, 2, strokeClr, true)

	vector.StrokeLine(dst, x+cellSize, y, x+cellSize-notch, y, 2, strokeClr, true)
	vector.StrokeLine(dst, x+cellSize, y, x+cellSize, y+notch, 2, strokeClr, true)

	vector.StrokeLine(dst, x, y+cellSize, x+notch, y+cellSize, 2, strokeClr, true)
	vector.StrokeLine(dst, x, y+cellSize, x, y+cellSize-notch, 2, strokeClr, true)

	vector.StrokeLine(dst, x+cellSize, y+cellSize, x+cellSize-notch, y+cellSize, 2, strokeClr, true)
	vector.StrokeLine(dst, x+cellSize, y+cellSize, x+cellSize, y+cellSize-notch, 2, strokeClr, true)
}

// DrawRangeCircle renders collector attack/sweep range overlay.
func DrawRangeCircle(dst *ebiten.Image, cx, cy, rangeRadius float32, clr color.RGBA) {
	fillClr := color.RGBA{R: clr.R, G: clr.G, B: clr.B, A: uint8(float32(clr.A) * 0.15)}
	vector.DrawFilledCircle(dst, cx, cy, rangeRadius, fillClr, true)
	vector.StrokeCircle(dst, cx, cy, rangeRadius, 1.2, clr, true)
}

// PrintVectorInfo is a helper for debugging vector parameters.
func PrintVectorInfo() string {
	return "internal/art: vector graphics primitives initialized"
}
