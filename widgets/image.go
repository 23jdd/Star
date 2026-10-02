package widgets

import "github.com/gdamore/tcell/v2"

// Cell is one terminal cell in a Canvas.
type Cell struct {
	Main      rune
	Combining []rune
	Style     tcell.Style
}

// Canvas is a retained cell buffer for custom charts, diagrams, and terminal
// art. Cells outside its arranged bounds are clipped.
type Canvas struct {
	WidgetBase
	Width, Height int
	Cells         []Cell
}

// NewCanvas creates an empty retained cell buffer.
func NewCanvas(width, height int) *Canvas {
	canvas := &Canvas{}
	canvas.Resize(width, height)
	return canvas
}

// Resize changes the canvas dimensions and preserves overlapping cells.
func (c *Canvas) Resize(width, height int) {
	width, height = max(0, width), max(0, height)
	resized := make([]Cell, width*height)
	copyWidth, copyHeight := min(width, c.Width), min(height, c.Height)
	for y := 0; y < copyHeight; y++ {
		copy(resized[y*width:y*width+copyWidth], c.Cells[y*c.Width:y*c.Width+copyWidth])
	}
	c.Width, c.Height, c.Cells = width, height, resized
}

// SetCell stores a copied cell and reports whether the coordinates were valid.
func (c *Canvas) SetCell(x, y int, cell Cell) bool {
	if x < 0 || y < 0 || x >= c.Width || y >= c.Height {
		return false
	}
	cell.Combining = append([]rune(nil), cell.Combining...)
	c.Cells[y*c.Width+x] = cell
	return true
}

// Cell returns a copy of the requested cell.
func (c *Canvas) Cell(x, y int) (Cell, bool) {
	if x < 0 || y < 0 || x >= c.Width || y >= c.Height {
		return Cell{}, false
	}
	cell := c.Cells[y*c.Width+x]
	cell.Combining = append([]rune(nil), cell.Combining...)
	return cell, true
}

func (c *Canvas) Measure(constraints Constraints) Size {
	return constraints.Constrain(Size{W: c.Width, H: c.Height})
}

func (c *Canvas) Arrange(bounds Rect) { c.SetBounds(bounds) }
func (c *Canvas) HandlerEvent(*Event) {}

func (c *Canvas) Render(screen tcell.Screen) {
	if c.Hidden() {
		return
	}
	height, width := min(c.Height, c.Bounds().H), min(c.Width, c.Bounds().W)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := c.Cells[y*c.Width+x]
			main := cell.Main
			if main == 0 {
				main = ' '
			}
			screen.SetContent(c.Bounds().X+x, c.Bounds().Y+y, main, cell.Combining, cell.Style)
		}
	}
}
