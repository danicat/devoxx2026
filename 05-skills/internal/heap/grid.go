package heap

import (
	"fmt"
	"math"
)

// Virtual canvas and grid constants as specified in ARCHITECTURE.md.
const (
	VirtualWidth  = 960
	VirtualHeight = 540
	TileSize      = 30
	GridCols      = 32 // 32 * 30 = 960
	GridRows      = 18 // 18 * 30 = 540

	// Play arena bounds (X: 30-750, Y: 45-495)
	MinCol = 1
	MaxCol = 24
	MinRow = 1
	MaxRow = 16
)

// CellType indicates the state of a tile on the heap grid.
type CellType int

const (
	CellEmpty CellType = iota
	CellPath
	CellBlocked
	CellOccupied
)

// String returns a human-readable representation of CellType.
func (c CellType) String() string {
	switch c {
	case CellEmpty:
		return "Empty"
	case CellPath:
		return "Path"
	case CellBlocked:
		return "Blocked"
	case CellOccupied:
		return "Occupied"
	default:
		return fmt.Sprintf("Unknown(%d)", c)
	}
}

// Grid represents the 32x18 tile arena for JVM heap memory and collector placement.
type Grid struct {
	cells [GridCols][GridRows]CellType
}

// NewGrid creates a new Grid with all cells initialized to CellEmpty.
func NewGrid() *Grid {
	return &Grid{}
}

// GridToWorldCenter returns the world coordinates (x, y) at the center of the tile.
func GridToWorldCenter(col, row int) (x, y float64) {
	x = float64(col)*float64(TileSize) + float64(TileSize)/2.0
	y = float64(row)*float64(TileSize) + float64(TileSize)/2.0
	return x, y
}

// GridToWorldTopLeft returns the world coordinates (x, y) at the top-left corner of the tile.
func GridToWorldTopLeft(col, row int) (x, y float64) {
	x = float64(col) * float64(TileSize)
	y = float64(row) * float64(TileSize)
	return x, y
}

// WorldToGrid converts world pixel coordinates (x, y) into grid column and row indices.
func WorldToGrid(x, y float64) (col, row int) {
	col = int(math.Floor(x / float64(TileSize)))
	row = int(math.Floor(y / float64(TileSize)))
	return col, row
}

// IsValidGridPos returns true if (col, row) is within the 32x18 grid canvas bounds.
func IsValidGridPos(col, row int) bool {
	return col >= 0 && col < GridCols && row >= 0 && row < GridRows
}

// IsInPlayArea returns true if (col, row) is within the interactive play arena bounds.
func IsInPlayArea(col, row int) bool {
	return col >= MinCol && col <= MaxCol && row >= MinRow && row <= MaxRow
}

// IsBuildable checks if a tower can be placed on the given grid coordinate.
func (g *Grid) IsBuildable(col, row int) bool {
	if !IsInPlayArea(col, row) {
		return false
	}
	return g.cells[col][row] == CellEmpty
}

// SetCellType updates the cell state at the specified coordinate.
func (g *Grid) SetCellType(col, row int, t CellType) {
	if !IsValidGridPos(col, row) {
		return
	}
	g.cells[col][row] = t
}

// GetCellType returns the cell state at the specified coordinate, or CellBlocked if out of bounds.
func (g *Grid) GetCellType(col, row int) CellType {
	if !IsValidGridPos(col, row) {
		return CellBlocked
	}
	return g.cells[col][row]
}

// Reset clears all cells back to CellEmpty.
func (g *Grid) Reset() {
	for c := 0; c < GridCols; c++ {
		for r := 0; r < GridRows; r++ {
			g.cells[c][r] = CellEmpty
		}
	}
}
