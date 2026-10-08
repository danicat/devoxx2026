package art

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ProceduralAssets stores all procedurally generated textures and sprites.
type ProceduralAssets struct {
	SiliconTile     *ebiten.Image
	BusTrace        *ebiten.Image
	CollectorTurret map[int]*ebiten.Image
	ObjectSprites   map[int]*ebiten.Image
}

// GenerateAssets procedurally generates all bitmap assets with pure code.
// Fails fast if any asset creation step returns nil or errors.
func GenerateAssets() (*ProceduralAssets, error) {
	assets := &ProceduralAssets{
		CollectorTurret: make(map[int]*ebiten.Image),
		ObjectSprites:   make(map[int]*ebiten.Image),
	}

	// 1. SiliconTile (30x30 dark silicon substrate with micro-traces)
	silicon := ebiten.NewImage(30, 30)
	if silicon == nil {
		return nil, fmt.Errorf("art: failed to allocate SiliconTile image")
	}
	renderSiliconTile(silicon)
	assets.SiliconTile = silicon

	// 2. BusTrace (30x30 glowing bus line texture)
	bus := ebiten.NewImage(30, 30)
	if bus == nil {
		return nil, fmt.Errorf("art: failed to allocate BusTrace image")
	}
	renderBusTrace(bus)
	assets.BusTrace = bus

	// 3. CollectorTurrets:
	// 0: ColSerial (single focus laser emitter, cyan/blue)
	// 1: ColParallel (dual burst emitters, neon green)
	// 2: ColCMS (circular radial pulse dish, purple/violet)
	// 3: ColG1 (diamond grid collector, amber/yellow)
	// 4: ColZGC (advanced prism low-latency laser, bright white/cyan)
	for id := 0; id <= 4; id++ {
		img := ebiten.NewImage(30, 30)
		if img == nil {
			return nil, fmt.Errorf("art: failed to allocate CollectorTurret image for id=%d", id)
		}
		switch id {
		case 0: // ColSerial
			renderColSerial(img)
		case 1: // ColParallel
			renderColParallel(img)
		case 2: // ColCMS
			renderColCMS(img)
		case 3: // ColG1
			renderColG1(img)
		case 4: // ColZGC
			renderColZGC(img)
		}
		assets.CollectorTurret[id] = img
	}

	// 4. ObjectSprites:
	// 0: ObjLambda (small fast green triangle/particle)
	// 1: ObjString (amber rounded capsule/block)
	// 2: ObjThreadLocal (armored purple hexagon with lock/shield ring)
	// 3: ObjLargeBlob (heavy pulsating crimson/red allocation block)
	for id := 0; id <= 3; id++ {
		img := ebiten.NewImage(30, 30)
		if img == nil {
			return nil, fmt.Errorf("art: failed to allocate ObjectSprite image for id=%d", id)
		}
		switch id {
		case 0: // ObjLambda
			renderObjLambda(img)
		case 1: // ObjString
			renderObjString(img)
		case 2: // ObjThreadLocal
			renderObjThreadLocal(img)
		case 3: // ObjLargeBlob
			renderObjLargeBlob(img)
		}
		assets.ObjectSprites[id] = img
	}

	return assets, nil
}

// renderSiliconTile draws a 30x30 dark silicon substrate with micro-traces and solder pads.
func renderSiliconTile(dst *ebiten.Image) {
	// Dark silicon base
	vector.DrawFilledRect(dst, 0, 0, 30, 30, color.RGBA{R: 10, G: 14, B: 22, A: 255}, true)
	// Border grid lines
	vector.StrokeRect(dst, 0.5, 0.5, 29, 29, 1, color.RGBA{R: 20, G: 32, B: 48, A: 255}, true)

	// Micro-traces
	vector.StrokeLine(dst, 4, 15, 12, 15, 1, color.RGBA{R: 26, G: 45, B: 66, A: 255}, true)
	vector.StrokeLine(dst, 12, 15, 15, 12, 1, color.RGBA{R: 26, G: 45, B: 66, A: 255}, true)
	vector.StrokeLine(dst, 15, 12, 26, 12, 1, color.RGBA{R: 26, G: 45, B: 66, A: 255}, true)

	vector.StrokeLine(dst, 15, 18, 15, 26, 1, color.RGBA{R: 26, G: 45, B: 66, A: 255}, true)
	vector.StrokeLine(dst, 15, 26, 22, 26, 1, color.RGBA{R: 26, G: 45, B: 66, A: 255}, true)

	// Micro solder pads
	vector.DrawFilledCircle(dst, 4, 15, 1.2, color.RGBA{R: 40, G: 70, B: 100, A: 255}, true)
	vector.DrawFilledCircle(dst, 26, 12, 1.2, color.RGBA{R: 40, G: 70, B: 100, A: 255}, true)
	vector.DrawFilledCircle(dst, 22, 26, 1.2, color.RGBA{R: 40, G: 70, B: 100, A: 255}, true)
}

// renderBusTrace draws a 30x30 glowing bus line texture.
func renderBusTrace(dst *ebiten.Image) {
	// Pathway background
	vector.DrawFilledRect(dst, 0, 0, 30, 30, color.RGBA{R: 14, G: 20, B: 30, A: 255}, true)

	// Outer glow lines for bus channel (center Y = 15, height 16)
	vector.DrawFilledRect(dst, 0, 7, 30, 16, color.RGBA{R: 16, G: 40, B: 55, A: 140}, true)

	// Rails
	vector.StrokeLine(dst, 0, 7.5, 30, 7.5, 1, color.RGBA{R: 0, G: 180, B: 216, A: 180}, true)
	vector.StrokeLine(dst, 0, 22.5, 30, 22.5, 1, color.RGBA{R: 0, G: 180, B: 216, A: 180}, true)

	// Core data bus trace
	vector.StrokeLine(dst, 0, 15, 30, 15, 2, color.RGBA{R: 72, G: 202, B: 228, A: 255}, true)
	// Bright center line
	vector.StrokeLine(dst, 0, 15, 30, 15, 1, color.RGBA{R: 202, G: 240, B: 248, A: 255}, true)

	// Cross-clock tick marks
	for x := float32(5); x < 30; x += 10 {
		vector.StrokeLine(dst, x, 10, x, 20, 1, color.RGBA{R: 0, G: 230, B: 255, A: 90}, true)
	}
}

// 0: ColSerial (single focus laser emitter, cyan/blue)
func renderColSerial(dst *ebiten.Image) {
	// Base octagon / circle
	cx, cy := float32(15), float32(15)
	vector.DrawFilledCircle(dst, cx, cy, 11, color.RGBA{R: 12, G: 32, B: 50, A: 255}, true)
	vector.StrokeCircle(dst, cx, cy, 11, 1.5, color.RGBA{R: 0, G: 180, B: 230, A: 255}, true)

	// Inner core
	vector.DrawFilledCircle(dst, cx, cy, 6, color.RGBA{R: 16, G: 64, B: 96, A: 255}, true)
	vector.StrokeCircle(dst, cx, cy, 6, 1, color.RGBA{R: 100, G: 220, B: 255, A: 255}, true)

	// Single forward focus emitter nozzle (pointing up/right)
	vector.DrawFilledRect(dst, 13, 2, 4, 8, color.RGBA{R: 0, G: 200, B: 255, A: 255}, true)
	vector.DrawFilledRect(dst, 14, 1, 2, 3, color.RGBA{R: 220, G: 255, B: 255, A: 255}, true)

	// Center focus crystal
	vector.DrawFilledCircle(dst, cx, cy, 3, color.RGBA{R: 200, G: 245, B: 255, A: 255}, true)
}

// 1: ColParallel (dual burst emitters, neon green)
func renderColParallel(dst *ebiten.Image) {
	cx, cy := float32(15), float32(15)
	// Base rounded chassis
	vector.DrawFilledRect(dst, 5, 5, 20, 20, color.RGBA{R: 10, G: 35, B: 20, A: 255}, true)
	vector.StrokeRect(dst, 5.5, 5.5, 19, 19, 1.5, color.RGBA{R: 0, G: 230, B: 118, A: 255}, true)

	// Dual barrels
	vector.DrawFilledRect(dst, 8, 2, 4, 8, color.RGBA{R: 0, G: 200, B: 100, A: 255}, true)
	vector.DrawFilledRect(dst, 18, 2, 4, 8, color.RGBA{R: 0, G: 200, B: 100, A: 255}, true)
	vector.DrawFilledRect(dst, 9, 1, 2, 3, color.RGBA{R: 180, G: 255, B: 200, A: 255}, true)
	vector.DrawFilledRect(dst, 19, 1, 2, 3, color.RGBA{R: 180, G: 255, B: 200, A: 255}, true)

	// Generator core
	vector.DrawFilledCircle(dst, cx, cy+2, 5, color.RGBA{R: 0, G: 160, B: 80, A: 255}, true)
	vector.DrawFilledCircle(dst, cx, cy+2, 2.5, color.RGBA{R: 180, G: 255, B: 220, A: 255}, true)
}

// 2: ColCMS (circular radial pulse dish, purple/violet)
func renderColCMS(dst *ebiten.Image) {
	cx, cy := float32(15), float32(15)
	// Concentric pulse rings
	vector.DrawFilledCircle(dst, cx, cy, 12, color.RGBA{R: 35, G: 12, B: 45, A: 255}, true)
	vector.StrokeCircle(dst, cx, cy, 12, 1.5, color.RGBA{R: 170, G: 0, B: 255, A: 255}, true)
	vector.StrokeCircle(dst, cx, cy, 8, 1, color.RGBA{R: 210, G: 100, B: 255, A: 200}, true)

	// 4 radial emitter nodes
	for angle := 0.0; angle < 2*math.Pi; angle += math.Pi / 2 {
		nx := cx + float32(math.Cos(angle))*9
		ny := cy + float32(math.Sin(angle))*9
		vector.DrawFilledCircle(dst, nx, ny, 2, color.RGBA{R: 240, G: 180, B: 255, A: 255}, true)
	}

	// Pulsing center sphere
	vector.DrawFilledCircle(dst, cx, cy, 4, color.RGBA{R: 220, G: 120, B: 255, A: 255}, true)
	vector.DrawFilledCircle(dst, cx, cy, 2, color.RGBA{R: 255, G: 240, B: 255, A: 255}, true)
}

// 3: ColG1 (diamond grid collector, amber/yellow)
func renderColG1(dst *ebiten.Image) {
	cx, cy := float32(15), float32(15)
	// Diamond outer polygon: (15,2), (28,15), (15,28), (2,15)
	var path vector.Path
	path.MoveTo(cx, 2)
	path.LineTo(28, cy)
	path.LineTo(cx, 28)
	path.LineTo(2, cy)
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	drawPathVertices(dst, vs, is, color.RGBA{R: 45, G: 35, B: 10, A: 255})

	// Diamond stroke
	vector.StrokeLine(dst, cx, 2, 28, cy, 1.5, color.RGBA{R: 255, G: 195, B: 0, A: 255}, true)
	vector.StrokeLine(dst, 28, cy, cx, 28, 1.5, color.RGBA{R: 255, G: 195, B: 0, A: 255}, true)
	vector.StrokeLine(dst, cx, 28, 2, cy, 1.5, color.RGBA{R: 255, G: 195, B: 0, A: 255}, true)
	vector.StrokeLine(dst, 2, cy, cx, 2, 1.5, color.RGBA{R: 255, G: 195, B: 0, A: 255}, true)

	// Inner grid cross
	vector.StrokeLine(dst, cx, 6, cx, 24, 1.2, color.RGBA{R: 255, G: 220, B: 100, A: 220}, true)
	vector.StrokeLine(dst, 6, cy, 24, cy, 1.2, color.RGBA{R: 255, G: 220, B: 100, A: 220}, true)

	// Core energy cell
	vector.DrawFilledRect(dst, 12, 12, 6, 6, color.RGBA{R: 255, G: 235, B: 140, A: 255}, true)
}

// 4: ColZGC (advanced prism low-latency laser, bright white/cyan)
func renderColZGC(dst *ebiten.Image) {
	cx, cy := float32(15), float32(15)
	// Hexagonal / crystalline prism base
	radius := float32(12)
	var path vector.Path
	for i := 0; i < 6; i++ {
		rad := float64(i)*math.Pi/3 - math.Pi/6
		px := cx + radius*float32(math.Cos(rad))
		py := cy + radius*float32(math.Sin(rad))
		if i == 0 {
			path.MoveTo(px, py)
		} else {
			path.LineTo(px, py)
		}
	}
	path.Close()
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	drawPathVertices(dst, vs, is, color.RGBA{R: 20, G: 45, B: 55, A: 255})

	// Prism borders
	for i := 0; i < 6; i++ {
		rad1 := float64(i)*math.Pi/3 - math.Pi/6
		rad2 := float64((i+1)%6)*math.Pi/3 - math.Pi/6
		x1 := cx + radius*float32(math.Cos(rad1))
		y1 := cy + radius*float32(math.Sin(rad1))
		x2 := cx + radius*float32(math.Cos(rad2))
		y2 := cy + radius*float32(math.Sin(rad2))
		vector.StrokeLine(dst, x1, y1, x2, y2, 1.5, color.RGBA{R: 120, G: 240, B: 255, A: 255}, true)
	}

	// Tri-focus internal beam channels radiating to center
	for i := 0; i < 3; i++ {
		rad := float64(i)*2*math.Pi/3 - math.Pi/2
		px := cx + 9*float32(math.Cos(rad))
		py := cy + 9*float32(math.Sin(rad))
		vector.StrokeLine(dst, cx, cy, px, py, 1.5, color.RGBA{R: 180, G: 250, B: 255, A: 220}, true)
	}

	// White-hot center focal core
	vector.DrawFilledCircle(dst, cx, cy, 4, color.RGBA{R: 220, G: 255, B: 255, A: 255}, true)
	vector.DrawFilledCircle(dst, cx, cy, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)
}

// 0: ObjLambda (small fast green triangle/particle)
func renderObjLambda(dst *ebiten.Image) {
	// Pointing right/forward: (24, 15), (7, 6), (7, 24)
	var path vector.Path
	path.MoveTo(24, 15)
	path.LineTo(7, 7)
	path.LineTo(11, 15)
	path.LineTo(7, 23)
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	drawPathVertices(dst, vs, is, color.RGBA{R: 0, G: 230, B: 118, A: 220})

	// Bright edge stroke
	vector.StrokeLine(dst, 24, 15, 7, 7, 1.5, color.RGBA{R: 180, G: 255, B: 200, A: 255}, true)
	vector.StrokeLine(dst, 7, 7, 11, 15, 1.2, color.RGBA{R: 0, G: 200, B: 100, A: 255}, true)
	vector.StrokeLine(dst, 11, 15, 7, 23, 1.2, color.RGBA{R: 0, G: 200, B: 100, A: 255}, true)
	vector.StrokeLine(dst, 7, 23, 24, 15, 1.5, color.RGBA{R: 180, G: 255, B: 200, A: 255}, true)

	// Core neon bead
	vector.DrawFilledCircle(dst, 13, 15, 2, color.RGBA{R: 220, G: 255, B: 230, A: 255}, true)
}

// 1: ObjString (amber rounded capsule/block)
func renderObjString(dst *ebiten.Image) {
	// Rounded rectangle capsule: X=5, Y=8, W=20, H=14, R=5
	cx1, cy := float32(10), float32(15)
	cx2 := float32(20)

	// Base body
	vector.DrawFilledCircle(dst, cx1, cy, 6, color.RGBA{R: 255, G: 160, B: 0, A: 230}, true)
	vector.DrawFilledCircle(dst, cx2, cy, 6, color.RGBA{R: 255, G: 160, B: 0, A: 230}, true)
	vector.DrawFilledRect(dst, 10, 9, 10, 12, color.RGBA{R: 255, G: 160, B: 0, A: 230}, true)

	// Glowing outline
	vector.StrokeLine(dst, 10, 9, 20, 9, 1.5, color.RGBA{R: 255, G: 225, B: 100, A: 255}, true)
	vector.StrokeLine(dst, 10, 21, 20, 21, 1.5, color.RGBA{R: 255, G: 225, B: 100, A: 255}, true)

	// Internal character byte tick marks (representing char[] value)
	vector.StrokeLine(dst, 11, 12, 11, 18, 1, color.RGBA{R: 255, G: 255, B: 200, A: 255}, true)
	vector.StrokeLine(dst, 15, 12, 15, 18, 1, color.RGBA{R: 255, G: 255, B: 200, A: 255}, true)
	vector.StrokeLine(dst, 19, 12, 19, 18, 1, color.RGBA{R: 255, G: 255, B: 200, A: 255}, true)
}

// 2: ObjThreadLocal (armored purple hexagon with lock/shield ring)
func renderObjThreadLocal(dst *ebiten.Image) {
	cx, cy := float32(15), float32(15)
	radius := float32(11)

	// Hexagon path
	var path vector.Path
	for i := 0; i < 6; i++ {
		rad := float64(i) * math.Pi / 3
		px := cx + radius*float32(math.Cos(rad))
		py := cy + radius*float32(math.Sin(rad))
		if i == 0 {
			path.MoveTo(px, py)
		} else {
			path.LineTo(px, py)
		}
	}
	path.Close()
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	drawPathVertices(dst, vs, is, color.RGBA{R: 120, G: 25, B: 170, A: 240})

	// Outer shield ring / border
	for i := 0; i < 6; i++ {
		rad1 := float64(i) * math.Pi / 3
		rad2 := float64((i+1)%6) * math.Pi / 3
		x1 := cx + radius*float32(math.Cos(rad1))
		y1 := cy + radius*float32(math.Sin(rad1))
		x2 := cx + radius*float32(math.Cos(rad2))
		y2 := cy + radius*float32(math.Sin(rad2))
		vector.StrokeLine(dst, x1, y1, x2, y2, 2, color.RGBA{R: 210, G: 90, B: 255, A: 255}, true)
	}

	// Inner lock emblem / keyhole circle
	vector.StrokeCircle(dst, cx, cy-2, 3, 1.2, color.RGBA{R: 255, G: 220, B: 255, A: 255}, true)
	vector.DrawFilledRect(dst, 13.5, 14, 3, 4, color.RGBA{R: 255, G: 220, B: 255, A: 255}, true)
}

// 3: ObjLargeBlob (heavy pulsating crimson/red allocation block)
func renderObjLargeBlob(dst *ebiten.Image) {
	// Heavy rounded octagonal block
	var path vector.Path
	path.MoveTo(9, 4)
	path.LineTo(21, 4)
	path.LineTo(26, 9)
	path.LineTo(26, 21)
	path.LineTo(21, 26)
	path.LineTo(9, 26)
	path.LineTo(4, 21)
	path.LineTo(4, 9)
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	drawPathVertices(dst, vs, is, color.RGBA{R: 180, G: 15, B: 40, A: 250})

	// Alert border
	vector.StrokeLine(dst, 9, 4, 21, 4, 2, color.RGBA{R: 255, G: 50, B: 80, A: 255}, true)
	vector.StrokeLine(dst, 21, 4, 26, 9, 2, color.RGBA{R: 255, G: 50, B: 80, A: 255}, true)
	vector.StrokeLine(dst, 26, 9, 26, 21, 2, color.RGBA{R: 255, G: 50, B: 80, A: 255}, true)
	vector.StrokeLine(dst, 26, 21, 21, 26, 2, color.RGBA{R: 255, G: 50, B: 80, A: 255}, true)
	vector.StrokeLine(dst, 21, 26, 9, 26, 2, color.RGBA{R: 255, G: 50, B: 80, A: 255}, true)
	vector.StrokeLine(dst, 9, 26, 4, 21, 2, color.RGBA{R: 255, G: 50, B: 80, A: 255}, true)
	vector.StrokeLine(dst, 4, 21, 4, 9, 2, color.RGBA{R: 255, G: 50, B: 80, A: 255}, true)
	vector.StrokeLine(dst, 4, 9, 9, 4, 2, color.RGBA{R: 255, G: 50, B: 80, A: 255}, true)

	// Cross warning stripes inside block
	vector.StrokeLine(dst, 10, 10, 20, 20, 1.5, color.RGBA{R: 255, G: 120, B: 140, A: 220}, true)
	vector.StrokeLine(dst, 20, 10, 10, 20, 1.5, color.RGBA{R: 255, G: 120, B: 140, A: 220}, true)
	vector.DrawFilledCircle(dst, 15, 15, 3, color.RGBA{R: 255, G: 200, B: 210, A: 255}, true)
}

// drawPathVertices fills a vector path using ebiten Triangles on a white 1x1 sub-image.
var whiteSubImage *ebiten.Image

func init() {
	whiteSubImage = ebiten.NewImage(3, 3)
	whiteSubImage.Fill(color.White)
}

func drawPathVertices(dst *ebiten.Image, vs []ebiten.Vertex, is []uint16, clr color.RGBA) {
	r := float32(clr.R) / 255.0
	g := float32(clr.G) / 255.0
	b := float32(clr.B) / 255.0
	a := float32(clr.A) / 255.0

	// Set color and uv on each vertex
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = r
		vs[i].ColorG = g
		vs[i].ColorB = b
		vs[i].ColorA = a
	}

	opt := &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
	}
	dst.DrawTriangles(vs, is, whiteSubImage.SubImage(whiteSubImage.Bounds()).(*ebiten.Image), opt)
}
