package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// CollectorInspectorData contains snapshot details of the currently inspected GC collector.
type CollectorInspectorData struct {
	ID            int64
	Type          int // 0: Serial, 1: Parallel, 2: CMS, 3: G1, 4: ZGC
	TypeName      string
	Description   string
	Level         int
	MaxLevel      int
	Range         float64
	DPS           float64
	FireRate      float64
	UpgradeCost   int
	RecycleRefund int
	CanUpgrade    bool
	IsMaxLevel    bool
}

// InspectorCallbacks defines callback handlers for inspector panel actions.
type InspectorCallbacks struct {
	OnUpgrade func(id int64)
	OnRecycle func(id int64)
	OnClose   func()
}

// InspectorPanel is the side docked UI panel displaying stats, upgrade, and recycle controls
// for a selected collector tower.
type InspectorPanel struct {
	X, Y, W, H  float32
	Data        *CollectorInspectorData
	Callbacks   InspectorCallbacks

	btnUpgrade   *Button
	btnRecycle   *Button
	btnClose     *Button
}

// NewInspectorPanel initializes the collector inspector panel.
func NewInspectorPanel(x, y, w, h float32, callbacks InspectorCallbacks) *InspectorPanel {
	p := &InspectorPanel{
		X:         x,
		Y:         y,
		W:         w,
		H:         h,
		Callbacks: callbacks,
	}

	// Action buttons inside inspector
	btnW := w - 24
	btnX := x + 12

	p.btnUpgrade = NewButton(btnX, y+270, btnW, 36, "UPGRADE", "[U]", ebiten.KeyU, func() {
		if p.Data != nil && p.Callbacks.OnUpgrade != nil && p.Data.CanUpgrade && !p.Data.IsMaxLevel {
			p.Callbacks.OnUpgrade(p.Data.ID)
		}
	})
	p.btnUpgrade.Style.NormalBorder = color.RGBA{R: 0, G: 230, B: 118, A: 220} // Green accent

	p.btnRecycle = NewButton(btnX, y+316, btnW, 34, "RECYCLE", "[R]", ebiten.KeyR, func() {
		if p.Data != nil && p.Callbacks.OnRecycle != nil {
			p.Callbacks.OnRecycle(p.Data.ID)
		}
	})
	p.btnRecycle.Style.NormalBorder = color.RGBA{R: 255, G: 110, B: 64, A: 220} // Amber-orange accent
	p.btnRecycle.Style.TextColor = color.RGBA{R: 255, G: 171, B: 145, A: 255}

	p.btnClose = NewButton(x+w-28, y+8, 20, 20, "X", "[ESC]", ebiten.KeyEscape, func() {
		if p.Callbacks.OnClose != nil {
			p.Callbacks.OnClose()
		}
	})
	p.btnClose.Style.NormalBorder = color.RGBA{R: 120, G: 140, B: 160, A: 180}
	p.btnClose.Style.TextColor = color.RGBA{R: 176, G: 190, B: 197, A: 255}
	p.btnClose.ShortcutHint = "" // Keep button compact

	return p
}

// SetData updates the currently inspected collector data. Pass nil to hide panel.
func (p *InspectorPanel) SetData(data *CollectorInspectorData) {
	p.Data = data
	if data == nil {
		p.btnUpgrade.Visible = false
		p.btnRecycle.Visible = false
		p.btnClose.Visible = false
		return
	}

	p.btnUpgrade.Visible = true
	p.btnRecycle.Visible = true
	p.btnClose.Visible = true

	// Configure upgrade button
	if data.IsMaxLevel {
		p.btnUpgrade.Label = "MAX LEVEL"
		p.btnUpgrade.Subtitle = "Tuned fully"
		p.btnUpgrade.Disabled = true
	} else {
		p.btnUpgrade.Label = "UPGRADE"
		p.btnUpgrade.Subtitle = fmt.Sprintf("Cost: %d CPU", data.UpgradeCost)
		p.btnUpgrade.Disabled = !data.CanUpgrade
	}

	// Configure recycle button
	p.btnRecycle.Label = "RECYCLE"
	p.btnRecycle.Subtitle = fmt.Sprintf("Refund +%d CPU", data.RecycleRefund)
	p.btnRecycle.Disabled = false
}

// IsVisible returns whether the inspector has an active target.
func (p *InspectorPanel) IsVisible() bool {
	return p.Data != nil
}

// Contains checks if (px, py) is inside the inspector panel bounding box.
func (p *InspectorPanel) Contains(px, py float32) bool {
	if !p.IsVisible() {
		return false
	}
	return px >= p.X && px <= p.X+p.W && py >= p.Y && py <= p.Y+p.H
}

// Update processes input events on inspector buttons. Returns true if an action occurred.
func (p *InspectorPanel) Update() bool {
	if !p.IsVisible() {
		return false
	}

	action := false
	if p.btnClose.Update() {
		action = true
	}
	if p.btnUpgrade.Update() {
		action = true
	}
	if p.btnRecycle.Update() {
		action = true
	}

	return action
}

// Draw renders the cyber inspector panel, metrics, gauges, and buttons onto dst.
func (p *InspectorPanel) Draw(dst *ebiten.Image) {
	if !p.IsVisible() {
		return
	}

	data := p.Data

	// 1. Panel Background
	bgClr := color.RGBA{R: 10, G: 16, B: 24, A: 240}
	vector.DrawFilledRect(dst, p.X, p.Y, p.W, p.H, bgClr, true)

	// Panel Border & Accent Line
	borderClr := color.RGBA{R: 26, G: 70, B: 102, A: 220}
	vector.StrokeRect(dst, p.X, p.Y, p.W, p.H, 1.5, borderClr, true)

	// Top Header Banner
	headerBg := color.RGBA{R: 16, G: 32, B: 48, A: 240}
	vector.DrawFilledRect(dst, p.X+1, p.Y+1, p.W-2, 32, headerBg, true)
	vector.StrokeLine(dst, p.X, p.Y+33, p.X+p.W, p.Y+33, 1.5, color.RGBA{R: 0, G: 229, B: 255, A: 200}, true)

	// Header Title
	text.Draw(dst, "COLLECTOR STATS", basicfont.Face7x13, int(p.X)+12, int(p.Y)+21, color.RGBA{R: 0, G: 229, B: 255, A: 255})

	// Close Button
	p.btnClose.Draw(dst)

	// 2. Collector Name & Type Color
	nameClr := getCollectorTypeColor(data.Type)
	nameStr := data.TypeName
	if nameStr == "" {
		nameStr = defaultCollectorName(data.Type)
	}
	text.Draw(dst, nameStr, basicfont.Face7x13, int(p.X)+14, int(p.Y)+54, nameClr)

	// Level Indicator Badge
	lvlStr := fmt.Sprintf("LEVEL: %d", data.Level)
	if data.MaxLevel > 0 {
		lvlStr = fmt.Sprintf("LEVEL: %d/%d", data.Level, data.MaxLevel)
	}
	text.Draw(dst, lvlStr, basicfont.Face7x13, int(p.X)+14, int(p.Y)+72, color.RGBA{R: 255, G: 214, B: 0, A: 240})

	// Level Pip Bars
	maxPips := data.MaxLevel
	if maxPips <= 0 {
		maxPips = 5
	}
	pipW := float32((p.W - 28 - float32(maxPips-1)*4) / float32(maxPips))
	if pipW > 30 {
		pipW = 30
	}
	for i := 0; i < maxPips; i++ {
		pipX := p.X + 14 + float32(i)*(pipW+4)
		pipY := p.Y + 80
		if i < data.Level {
			vector.DrawFilledRect(dst, pipX, pipY, pipW, 6, color.RGBA{R: 0, G: 230, B: 118, A: 240}, true)
		} else {
			vector.DrawFilledRect(dst, pipX, pipY, pipW, 6, color.RGBA{R: 35, G: 45, B: 55, A: 200}, true)
		}
		vector.StrokeRect(dst, pipX, pipY, pipW, 6, 1, color.RGBA{R: 60, G: 80, B: 100, A: 200}, true)
	}

	// 3. Stats Section Divider
	vector.StrokeLine(dst, p.X+12, p.Y+98, p.X+p.W-12, p.Y+98, 1, color.RGBA{R: 35, G: 60, B: 85, A: 180}, true)

	// Numerical Metrics
	labelClr := color.RGBA{R: 140, G: 165, B: 180, A: 240}
	valClr := color.RGBA{R: 230, G: 245, B: 255, A: 255}

	drawMetricRow := func(yOff int, label, val string) {
		text.Draw(dst, label, basicfont.Face7x13, int(p.X)+14, int(p.Y)+yOff, labelClr)
		valX := int(p.X+p.W) - 14 - len(val)*7
		text.Draw(dst, val, basicfont.Face7x13, valX, int(p.Y)+yOff, valClr)
	}

	drawMetricRow(120, "RANGE", fmt.Sprintf("%.0f px", data.Range))
	drawMetricRow(142, "DPS", fmt.Sprintf("%.1f", data.DPS))
	drawMetricRow(164, "FIRE RATE", fmt.Sprintf("%.1f/s", data.FireRate))

	// Description or specialization note
	descY := int(p.Y) + 190
	if data.Description != "" {
		text.Draw(dst, "ROLE:", basicfont.Face7x13, int(p.X)+14, descY, labelClr)
		text.Draw(dst, data.Description, basicfont.Face7x13, int(p.X)+14, descY+16, color.RGBA{R: 180, G: 215, B: 230, A: 220})
	}

	// 4. Action Buttons
	p.btnUpgrade.Draw(dst)
	p.btnRecycle.Draw(dst)
}

// defaultCollectorName returns friendly name for collector enum.
func defaultCollectorName(colType int) string {
	switch colType {
	case 0:
		return "SERIAL GC"
	case 1:
		return "PARALLEL GC"
	case 2:
		return "CMS SWEEPER"
	case 3:
		return "G1 REGIONAL"
	case 4:
		return "ZGC LASER"
	default:
		return "COLLECTOR"
	}
}

// getCollectorTypeColor provides distinctive cyber hues per collector archetype.
func getCollectorTypeColor(colType int) color.RGBA {
	switch colType {
	case 0:
		return color.RGBA{R: 0, G: 230, B: 118, A: 255} // Emerald Green
	case 1:
		return color.RGBA{R: 0, G: 229, B: 255, A: 255} // Neon Cyan
	case 2:
		return color.RGBA{R: 224, G: 64, B: 251, A: 255} // Neon Purple / Orchid
	case 3:
		return color.RGBA{R: 255, G: 214, B: 0, A: 255} // Warm Amber
	case 4:
		return color.RGBA{R: 255, G: 23, B: 68, A: 255} // Neon Crimson
	default:
		return color.RGBA{R: 200, G: 220, B: 240, A: 255}
	}
}
