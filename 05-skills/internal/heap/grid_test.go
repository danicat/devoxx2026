package heap

import (
	"testing"
)

func TestGridConstants(t *testing.T) {
	if VirtualWidth != 960 {
		t.Errorf("expected VirtualWidth 960, got %d", VirtualWidth)
	}
	if VirtualHeight != 540 {
		t.Errorf("expected VirtualHeight 540, got %d", VirtualHeight)
	}
	if TileSize != 30 {
		t.Errorf("expected TileSize 30, got %d", TileSize)
	}
	if GridCols*TileSize != VirtualWidth {
		t.Errorf("GridCols * TileSize (%d * %d) != VirtualWidth (%d)", GridCols, TileSize, VirtualWidth)
	}
	if GridRows*TileSize != VirtualHeight {
		t.Errorf("GridRows * TileSize (%d * %d) != VirtualHeight (%d)", GridRows, TileSize, VirtualHeight)
	}
	if MinCol != 1 || MaxCol != 24 || MinRow != 1 || MaxRow != 16 {
		t.Errorf("unexpected play arena bounds: cols [%d..%d], rows [%d..%d]", MinCol, MaxCol, MinRow, MaxRow)
	}
}

func TestCoordinateConversions(t *testing.T) {
	// Center conversion
	cx, cy := GridToWorldCenter(1, 1)
	if cx != 45.0 || cy != 45.0 {
		t.Errorf("expected center (45.0, 45.0), got (%f, %f)", cx, cy)
	}

	// TopLeft conversion
	tlx, tly := GridToWorldTopLeft(2, 3)
	if tlx != 60.0 || tly != 90.0 {
		t.Errorf("expected top left (60.0, 90.0), got (%f, %f)", tlx, tly)
	}

	// WorldToGrid conversion
	col, row := WorldToGrid(45.0, 45.0)
	if col != 1 || row != 1 {
		t.Errorf("expected grid (1, 1), got (%d, %d)", col, row)
	}

	col, row = WorldToGrid(0.0, 0.0)
	if col != 0 || row != 0 {
		t.Errorf("expected grid (0, 0), got (%d, %d)", col, row)
	}

	col, row = WorldToGrid(29.99, 29.99)
	if col != 0 || row != 0 {
		t.Errorf("expected grid (0, 0), got (%d, %d)", col, row)
	}

	col, row = WorldToGrid(30.0, 30.0)
	if col != 1 || row != 1 {
		t.Errorf("expected grid (1, 1), got (%d, %d)", col, row)
	}

	col, row = WorldToGrid(-10.0, -10.0)
	if col >= 0 || row >= 0 {
		t.Errorf("expected negative grid indices for negative world pos, got (%d, %d)", col, row)
	}
}

func TestGridBoundsValidation(t *testing.T) {
	tests := []struct {
		col, row   int
		validGrid  bool
		inPlayArea bool
	}{
		{0, 0, true, false},
		{1, 1, true, true},
		{24, 16, true, true},
		{24, 17, true, false},
		{25, 16, true, false},
		{31, 17, true, false},
		{32, 17, false, false},
		{-1, 0, false, false},
		{0, -1, false, false},
		{10, 10, true, true},
	}

	for _, tc := range tests {
		vg := IsValidGridPos(tc.col, tc.row)
		pa := IsInPlayArea(tc.col, tc.row)
		if vg != tc.validGrid {
			t.Errorf("IsValidGridPos(%d, %d) = %v; want %v", tc.col, tc.row, vg, tc.validGrid)
		}
		if pa != tc.inPlayArea {
			t.Errorf("IsInPlayArea(%d, %d) = %v; want %v", tc.col, tc.row, pa, tc.inPlayArea)
		}
	}
}

func TestGridOperations(t *testing.T) {
	grid := NewGrid()

	// Initial state should be CellEmpty everywhere
	for c := 0; c < GridCols; c++ {
		for r := 0; r < GridRows; r++ {
			if grid.GetCellType(c, r) != CellEmpty {
				t.Fatalf("expected CellEmpty at (%d, %d), got %v", c, r, grid.GetCellType(c, r))
			}
		}
	}

	// Out of bounds GetCellType should return CellBlocked
	if grid.GetCellType(-1, 5) != CellBlocked {
		t.Errorf("expected CellBlocked for negative col, got %v", grid.GetCellType(-1, 5))
	}
	if grid.GetCellType(40, 5) != CellBlocked {
		t.Errorf("expected CellBlocked for col > 31, got %v", grid.GetCellType(40, 5))
	}

	// IsBuildable tests
	if !grid.IsBuildable(5, 5) {
		t.Errorf("expected (5, 5) to be buildable")
	}
	// Non-play area empty cell is not buildable
	if grid.IsBuildable(0, 0) {
		t.Errorf("expected (0, 0) outside play area to NOT be buildable")
	}
	if grid.IsBuildable(25, 5) {
		t.Errorf("expected (25, 5) outside play area to NOT be buildable")
	}

	// SetCellType on valid tile
	grid.SetCellType(5, 5, CellOccupied)
	if grid.GetCellType(5, 5) != CellOccupied {
		t.Errorf("expected CellOccupied at (5, 5), got %v", grid.GetCellType(5, 5))
	}
	if grid.IsBuildable(5, 5) {
		t.Errorf("expected (5, 5) occupied to NOT be buildable")
	}

	grid.SetCellType(6, 6, CellPath)
	if grid.IsBuildable(6, 6) {
		t.Errorf("expected CellPath at (6, 6) to NOT be buildable")
	}

	// SetCellType on out of bounds should be a safe no-op
	grid.SetCellType(-1, -1, CellBlocked)
	grid.SetCellType(99, 99, CellBlocked)

	// Reset
	grid.Reset()
	if grid.GetCellType(5, 5) != CellEmpty {
		t.Errorf("expected CellEmpty at (5, 5) after Reset(), got %v", grid.GetCellType(5, 5))
	}
	if grid.GetCellType(6, 6) != CellEmpty {
		t.Errorf("expected CellEmpty at (6, 6) after Reset(), got %v", grid.GetCellType(6, 6))
	}
}

func TestCellTypeString(t *testing.T) {
	if CellEmpty.String() != "Empty" {
		t.Errorf("expected 'Empty', got %s", CellEmpty.String())
	}
	if CellPath.String() != "Path" {
		t.Errorf("expected 'Path', got %s", CellPath.String())
	}
	if CellBlocked.String() != "Blocked" {
		t.Errorf("expected 'Blocked', got %s", CellBlocked.String())
	}
	if CellOccupied.String() != "Occupied" {
		t.Errorf("expected 'Occupied', got %s", CellOccupied.String())
	}
	if CellType(99).String() != "Unknown(99)" {
		t.Errorf("expected 'Unknown(99)', got %s", CellType(99).String())
	}
}
