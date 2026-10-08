package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// ButtonStyle holds color and aesthetic styling configurations for a Button.
type ButtonStyle struct {
	NormalBg     color.RGBA
	HoverBg      color.RGBA
	PressedBg    color.RGBA
	SelectedBg   color.RGBA
	DisabledBg   color.RGBA
	NormalBorder color.RGBA
	HoverBorder  color.RGBA
	SelectBorder color.RGBA
	DisableBorder color.RGBA
	TextColor    color.RGBA
	SubtextColor color.RGBA
	BadgeBg      color.RGBA
	BadgeText    color.RGBA
}

// DefaultButtonStyle returns a sci-fi cyber silicon theme.
func DefaultButtonStyle() ButtonStyle {
	return ButtonStyle{
		NormalBg:      color.RGBA{R: 12, G: 20, B: 30, A: 230},
		HoverBg:       color.RGBA{R: 20, G: 38, B: 58, A: 245},
		PressedBg:     color.RGBA{R: 10, G: 45, B: 70, A: 255},
		SelectedBg:    color.RGBA{R: 15, G: 55, B: 45, A: 245},
		DisabledBg:    color.RGBA{R: 15, G: 18, B: 22, A: 180},
		NormalBorder:  color.RGBA{R: 30, G: 75, B: 110, A: 220},
		HoverBorder:   color.RGBA{R: 0, G: 229, B: 255, A: 255}, // Neon Cyan
		SelectBorder:  color.RGBA{R: 0, G: 230, B: 118, A: 255}, // Neon Emerald
		DisableBorder: color.RGBA{R: 40, G: 50, B: 60, A: 160},
		TextColor:     color.RGBA{R: 224, G: 247, B: 250, A: 255},
		SubtextColor:  color.RGBA{R: 255, G: 214, B: 0, A: 240}, // Warm Amber / Gold
		BadgeBg:       color.RGBA{R: 8, G: 14, B: 22, A: 240},
		BadgeText:     color.RGBA{R: 0, G: 229, B: 255, A: 255},
	}
}

// Button represents an interactive HUD button with bounds, hover/click detection,
// keyboard shortcut support, and procedural vector drawing.
type Button struct {
	ID           string
	X, Y, W, H   float32
	Label        string
	Subtitle     string
	ShortcutHint string
	ShortcutKey  ebiten.Key
	HasShortcut  bool
	Disabled     bool
	Selected     bool
	Visible      bool
	Style        ButtonStyle

	// Interaction state
	isHovered    bool
	isPressed    bool
	justClicked  bool

	// Callbacks
	OnClick func()
}

// NewButton constructs a new interactive Button.
func NewButton(x, y, w, h float32, label string, shortcutHint string, shortcutKey ebiten.Key, onClick func()) *Button {
	hasShortcut := shortcutKey != 0 || shortcutHint != ""
	return &Button{
		X:            x,
		Y:            y,
		W:            w,
		H:            h,
		Label:        label,
		Subtitle:     "",
		ShortcutHint: shortcutHint,
		ShortcutKey:  shortcutKey,
		HasShortcut:  hasShortcut,
		Disabled:     false,
		Selected:     false,
		Visible:      true,
		Style:        DefaultButtonStyle(),
		OnClick:      onClick,
	}
}

// Contains checks if the given coordinate (px, py) is inside the button bounding box.
func (b *Button) Contains(px, py float32) bool {
	return px >= b.X && px <= b.X+b.W && py >= b.Y && py <= b.Y+b.H
}

// Update processes mouse, touch, and shortcut inputs. Returns true if the button was clicked this frame.
func (b *Button) Update() bool {
	if !b.Visible {
		b.isHovered = false
		b.isPressed = false
		b.justClicked = false
		return false
	}

	b.justClicked = false

	if b.Disabled {
		b.isHovered = false
		b.isPressed = false
		return false
	}

	// 1. Check Keyboard Shortcut trigger
	if b.HasShortcut && b.ShortcutKey != 0 && inpututil.IsKeyJustPressed(b.ShortcutKey) {
		b.triggerClick()
		return true
	}

	// 2. Check Mouse / Cursor input
	cx, cy := ebiten.CursorPosition()
	fx, fy := float32(cx), float32(cy)
	b.isHovered = b.Contains(fx, fy)

	if b.isHovered {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			b.isPressed = true
		}
		if b.isPressed && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			b.isPressed = false
			b.triggerClick()
			return true
		}
	} else {
		// If mouse was released outside after pressing inside, cancel press
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			b.isPressed = false
		}
	}

	// 3. Check Touch input (for mobile/tablet/WASM touch)
	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	for _, id := range touchIDs {
		tx, ty := ebiten.TouchPosition(id)
		if b.Contains(float32(tx), float32(ty)) {
			b.triggerClick()
			return true
		}
	}

	return false
}

// triggerClick executes the OnClick callback and sets justClicked.
func (b *Button) triggerClick() {
	b.justClicked = true
	if b.OnClick != nil {
		b.OnClick()
	}
}

// Click programmatically simulates a button click.
func (b *Button) Click() {
	if b.Disabled || !b.Visible {
		return
	}
	b.triggerClick()
}

// Draw renders the procedural vector button onto dst.
func (b *Button) Draw(dst *ebiten.Image) {
	if !b.Visible {
		return
	}

	// Determine current background and border colors
	var bgClr, borderClr color.RGBA
	var textClr color.RGBA = b.Style.TextColor

	if b.Disabled {
		bgClr = b.Style.DisabledBg
		borderClr = b.Style.DisableBorder
		textClr = color.RGBA{R: 90, G: 105, B: 115, A: 180}
	} else if b.isPressed {
		bgClr = b.Style.PressedBg
		borderClr = b.Style.HoverBorder
	} else if b.Selected {
		bgClr = b.Style.SelectedBg
		borderClr = b.Style.SelectBorder
		textClr = color.RGBA{R: 167, G: 255, B: 235, A: 255}
	} else if b.isHovered {
		bgClr = b.Style.HoverBg
		borderClr = b.Style.HoverBorder
	} else {
		bgClr = b.Style.NormalBg
		borderClr = b.Style.NormalBorder
	}

	// 1. Draw Background
	vector.DrawFilledRect(dst, b.X, b.Y, b.W, b.H, bgClr, true)

	// 2. Draw Border (thicker if selected or hovered)
	var strokeWidth float32 = 1.0
	if b.Selected || b.isHovered {
		strokeWidth = 1.8
	}
	vector.StrokeRect(dst, b.X, b.Y, b.W, b.H, strokeWidth, borderClr, true)

	// 3. Draw Cyber Corner Notches for sci-fi look
	cornerLen := float32(4.0)
	if b.W < 30 || b.H < 20 {
		cornerLen = 2.0
	}
	// Top-left notch
	vector.StrokeLine(dst, b.X, b.Y, b.X+cornerLen, b.Y, strokeWidth+0.5, borderClr, true)
	vector.StrokeLine(dst, b.X, b.Y, b.X, b.Y+cornerLen, strokeWidth+0.5, borderClr, true)
	// Bottom-right notch
	vector.StrokeLine(dst, b.X+b.W, b.Y+b.H, b.X+b.W-cornerLen, b.Y+b.H, strokeWidth+0.5, borderClr, true)
	vector.StrokeLine(dst, b.X+b.W, b.Y+b.H, b.X+b.W, b.Y+b.H-cornerLen, strokeWidth+0.5, borderClr, true)

	// 4. Draw Shortcut Badge in top-right or top-left if provided
	if b.ShortcutHint != "" {
		badgeW := float32(len(b.ShortcutHint)*7 + 6)
		badgeH := float32(11)
		bx := b.X + b.W - badgeW - 2
		by := b.Y + 2
		vector.DrawFilledRect(dst, bx, by, badgeW, badgeH, b.Style.BadgeBg, true)
		vector.StrokeRect(dst, bx, by, badgeW, badgeH, 1, borderClr, true)
		text.Draw(dst, b.ShortcutHint, basicfont.Face7x13, int(bx)+3, int(by)+9, b.Style.BadgeText)
	}

	// 5. Draw Label & Subtitle Text
	labelLen := len(b.Label) * 7
	labelX := int(b.X) + int(b.W-float32(labelLen))/2
	if labelX < int(b.X)+2 {
		labelX = int(b.X) + 2
	}

	if b.Subtitle != "" {
		// Stacked label & subtitle
		labelY := int(b.Y) + int(b.H/2) - 4
		text.Draw(dst, b.Label, basicfont.Face7x13, labelX, labelY, textClr)

		subLen := len(b.Subtitle) * 7
		subX := int(b.X) + int(b.W-float32(subLen))/2
		if subX < int(b.X)+2 {
			subX = int(b.X) + 2
		}
		subY := labelY + 13
		subClr := b.Style.SubtextColor
		if b.Disabled {
			subClr = color.RGBA{R: 90, G: 105, B: 115, A: 180}
		}
		text.Draw(dst, b.Subtitle, basicfont.Face7x13, subX, subY, subClr)
	} else {
		// Single centered label
		labelY := int(b.Y) + int(b.H)/2 + 4
		text.Draw(dst, b.Label, basicfont.Face7x13, labelX, labelY, textClr)
	}
}

// IsHovered returns whether mouse cursor is hovering over the button.
func (b *Button) IsHovered() bool {
	return b.isHovered
}

// IsPressed returns whether mouse button is currently held down over the button.
func (b *Button) IsPressed() bool {
	return b.isPressed
}

// JustClicked returns true if triggered in the latest update cycle.
func (b *Button) JustClicked() bool {
	return b.justClicked
}
