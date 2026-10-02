package widgets

import "testing"

func FuzzConstraintsNeverReturnNegativeSize(f *testing.F) {
	f.Add(0, 0, 80, 24, 10, 5)
	f.Add(-10, -10, -1, -1, -50, -50)
	f.Fuzz(func(t *testing.T, minW, minH, maxW, maxH, width, height int) {
		size := (Constraints{MinW: minW, MinH: minH, MaxW: maxW, MaxH: maxH}).Constrain(Size{W: width, H: height})
		if size.W < 0 || size.H < 0 {
			t.Fatalf("negative size: %#v", size)
		}
	})
}

func FuzzCanvasBounds(f *testing.F) {
	f.Add(10, 5, 2, 3)
	f.Add(-1, -1, 0, 0)
	f.Fuzz(func(t *testing.T, width, height, x, y int) {
		if width > 1000 || height > 1000 || width < -1000 || height < -1000 {
			t.Skip()
		}
		canvas := NewCanvas(width, height)
		ok := canvas.SetCell(x, y, Cell{Main: 'x'})
		inside := x >= 0 && y >= 0 && x < canvas.Width && y < canvas.Height
		if ok != inside {
			t.Fatalf("SetCell(%d,%d)=%v, inside=%v", x, y, ok, inside)
		}
	})
}

func FuzzTextAreaCursorBounds(f *testing.F) {
	f.Add("hello\n世界", 1, 2, 20, 4)
	f.Add("\r\n", -10, 999, -1, -1)
	f.Fuzz(func(t *testing.T, value string, row, column, width, height int) {
		if len(value) > 10_000 {
			t.Skip()
		}
		area := NewTextArea(value)
		area.CursorRow, area.CursorCol = row, column
		area.Arrange(NewRect(0, 0, width, height))
		lines := area.lines()
		if area.CursorRow < 0 || area.CursorRow >= len(lines) {
			t.Fatalf("row %d outside %d lines", area.CursorRow, len(lines))
		}
		if area.CursorCol < 0 || area.CursorCol > len([]rune(lines[area.CursorRow])) {
			t.Fatalf("column %d outside line %q", area.CursorCol, lines[area.CursorRow])
		}
	})
}
