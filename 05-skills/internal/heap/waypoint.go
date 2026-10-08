package heap

import "math"

// Waypoint defines a 2D position in world pixel space.
type Waypoint struct {
	X, Y float64
}

// MemoryPath represents a contiguous multi-segment route traversed by memory allocations.
type MemoryPath struct {
	Waypoints         []Waypoint
	Length            float64
	segmentLengths    []float64
	cumulativeLengths []float64
}

// NewMemoryPath creates a new MemoryPath, calculating total distance and precomputing segment metrics.
func NewMemoryPath(waypoints []Waypoint) *MemoryPath {
	n := len(waypoints)
	if n == 0 {
		return &MemoryPath{
			Waypoints:         nil,
			Length:            0,
			segmentLengths:    nil,
			cumulativeLengths: nil,
		}
	}

	wps := make([]Waypoint, n)
	copy(wps, waypoints)

	if n == 1 {
		return &MemoryPath{
			Waypoints:         wps,
			Length:            0,
			segmentLengths:    []float64{0},
			cumulativeLengths: []float64{0},
		}
	}

	segLengths := make([]float64, n-1)
	cumLengths := make([]float64, n)
	cumLengths[0] = 0

	totalLength := 0.0
	for i := 0; i < n-1; i++ {
		dx := wps[i+1].X - wps[i].X
		dy := wps[i+1].Y - wps[i].Y
		dist := math.Hypot(dx, dy)
		segLengths[i] = dist
		totalLength += dist
		cumLengths[i+1] = totalLength
	}

	return &MemoryPath{
		Waypoints:         wps,
		Length:            totalLength,
		segmentLengths:    segLengths,
		cumulativeLengths: cumLengths,
	}
}

// PositionAtDistance interpolates world coordinates (x, y) along the path at dist pixels from spawn.
// If dist exceeds Length, reachedEnd is true and the final waypoint coordinate is returned.
func (p *MemoryPath) PositionAtDistance(dist float64) (x, y float64, reachedEnd bool) {
	n := len(p.Waypoints)
	if n == 0 {
		return 0, 0, true
	}
	if n == 1 {
		return p.Waypoints[0].X, p.Waypoints[0].Y, true
	}
	if dist <= 0 {
		return p.Waypoints[0].X, p.Waypoints[0].Y, false
	}
	if dist >= p.Length || p.Length <= 0 {
		last := p.Waypoints[n-1]
		return last.X, last.Y, true
	}

	// Find the segment containing dist
	for i := 0; i < len(p.segmentLengths); i++ {
		startDist := p.cumulativeLengths[i]
		endDist := p.cumulativeLengths[i+1]
		if dist <= endDist {
			segLen := p.segmentLengths[i]
			if segLen <= 0 {
				return p.Waypoints[i].X, p.Waypoints[i].Y, false
			}
			t := (dist - startDist) / segLen
			w0 := p.Waypoints[i]
			w1 := p.Waypoints[i+1]
			x = w0.X + t*(w1.X-w0.X)
			y = w0.Y + t*(w1.Y-w0.Y)
			return x, y, false
		}
	}

	last := p.Waypoints[n-1]
	return last.X, last.Y, true
}

// GenerationAtDistance maps a travel distance along the path to its corresponding JVM memory generation zone:
// - Eden Space: [0% - 35% of total path)
// - Survivor Space: [35% - 70% of total path)
// - Tenured / Old Gen Space: [70% - 100% of total path]
func (p *MemoryPath) GenerationAtDistance(dist float64) GenerationType {
	if p.Length <= 0 || dist <= 0 {
		return GenEden
	}
	ratio := dist / p.Length
	if ratio < 0.35 {
		return GenEden
	}
	if ratio < 0.70 {
		return GenSurvivor0
	}
	return GenTenured
}

// DefaultMemoryPath generates the official memory bus path from Eden (spawn at X: 30, Y: 105),
// winding through Survivor space, and exiting through Tenured / Old Gen at X: 750.
func DefaultMemoryPath() *MemoryPath {
	waypoints := []Waypoint{
		// --- Eden Generation Zone ---
		{X: 30, Y: 105},  // Spawn boundary at left (Col 1, Row 3)
		{X: 255, Y: 105}, // Col 8, Row 3
		{X: 255, Y: 225}, // Col 8, Row 7
		{X: 105, Y: 225}, // Col 3, Row 7
		{X: 105, Y: 345}, // Col 3, Row 11
		{X: 285, Y: 345}, // Col 9, Row 11

		// --- Survivor Generation Zone ---
		{X: 285, Y: 225}, // Col 9, Row 7
		{X: 435, Y: 225}, // Col 14, Row 7
		{X: 435, Y: 105}, // Col 14, Row 3
		{X: 555, Y: 105}, // Col 18, Row 3
		{X: 555, Y: 285}, // Col 18, Row 9

		// --- Tenured Old Gen Zone ---
		{X: 435, Y: 285}, // Col 14, Row 9
		{X: 435, Y: 435}, // Col 14, Row 14
		{X: 675, Y: 435}, // Col 22, Row 14
		{X: 675, Y: 225}, // Col 22, Row 7
		{X: 750, Y: 225}, // Right exit arena boundary
	}
	return NewMemoryPath(waypoints)
}

// MarkPathOnGrid marks all grid tiles traversed by the memory path as CellPath.
func MarkPathOnGrid(grid *Grid, path *MemoryPath) {
	if grid == nil || path == nil || len(path.Waypoints) == 0 {
		return
	}

	step := float64(TileSize) / 4.0
	if step <= 0 {
		step = 5.0
	}

	for d := 0.0; d <= path.Length; d += step {
		x, y, _ := path.PositionAtDistance(d)
		col, row := WorldToGrid(x, y)
		if IsValidGridPos(col, row) {
			grid.SetCellType(col, row, CellPath)
		}
	}

	// Ensure the exact end position tile is marked
	xEnd, yEnd, _ := path.PositionAtDistance(path.Length)
	colEnd, rowEnd := WorldToGrid(xEnd, yEnd)
	if IsValidGridPos(colEnd, rowEnd) {
		grid.SetCellType(colEnd, rowEnd, CellPath)
	}

	// Ensure all explicit waypoint tiles are marked
	for _, wp := range path.Waypoints {
		col, row := WorldToGrid(wp.X, wp.Y)
		if IsValidGridPos(col, row) {
			grid.SetCellType(col, row, CellPath)
		}
	}
}
