package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// TableColumn describes a table heading and its width. Width <= 0 shares the
// remaining space with other flexible columns.
type TableColumn struct {
	Title string
	Width int
}

// Table displays selectable rows with a fixed header.
type Table struct {
	WidgetBase
	Columns       []TableColumn
	Rows          [][]string
	Selected      int
	Offset        int
	HeaderStyle   tcell.Style
	Style         tcell.Style
	SelectedStyle tcell.Style
	Separator     rune
	Focused       bool
	Disabled      bool
}

// NewTable copies columns and creates an empty table.
func NewTable(columns ...TableColumn) *Table {
	t := &Table{
		Columns:       append([]TableColumn(nil), columns...),
		HeaderStyle:   tcell.StyleDefault.Bold(true),
		Style:         tcell.StyleDefault,
		SelectedStyle: tcell.StyleDefault.Reverse(true),
		Separator:     '│',
	}
	t.DefaultAction = t.defaultAction
	return t
}

func (t *Table) CanFocus() bool      { return !t.Disabled && !t.Hidden() }
func (t *Table) SetFocus(value bool) { t.Focused = value }
func (t *Table) HandlerEvent(*Event) {}

func (t *Table) Measure(c Constraints) Size {
	width := max(0, len(t.Columns)-1)
	for columnIndex, column := range t.Columns {
		columnWidth := max(column.Width, uniseg.StringWidth(column.Title))
		for _, row := range t.Rows {
			if columnIndex < len(row) {
				columnWidth = max(columnWidth, uniseg.StringWidth(row[columnIndex]))
			}
		}
		width += columnWidth
	}
	return c.Constrain(Size{W: width, H: len(t.Rows) + 1})
}

func (t *Table) Arrange(bounds Rect) {
	t.SetBounds(bounds)
	t.ensureVisible()
}

func (t *Table) Render(screen tcell.Screen) {
	if t.Hidden() || t.Bounds().Empty() {
		return
	}
	fill(screen, t.Bounds(), ' ', t.Style)
	widths := t.columnWidths(t.Bounds().W)
	t.drawRow(screen, t.Bounds().Y, columnTitles(t.Columns), widths, t.HeaderStyle)
	visibleRows := max(0, t.Bounds().H-1)
	for row := 0; row < visibleRows && t.Offset+row < len(t.Rows); row++ {
		index := t.Offset + row
		style := t.Style
		if index == t.Selected {
			style = t.SelectedStyle
		}
		y := t.Bounds().Y + 1 + row
		fill(screen, NewRect(t.Bounds().X, y, t.Bounds().W, 1), ' ', style)
		t.drawRow(screen, y, t.Rows[index], widths, style)
	}
}

// SetSelected selects and reveals a row, returning true when it changed.
func (t *Table) SetSelected(index int) bool {
	if len(t.Rows) == 0 {
		t.Selected, t.Offset = 0, 0
		return false
	}
	index = clamp(index, 0, len(t.Rows)-1)
	if index == t.Selected {
		return false
	}
	t.Selected = index
	t.ensureVisible()
	return true
}

func (t *Table) columnWidths(total int) []int {
	widths := make([]int, len(t.Columns))
	remaining := max(0, total-max(0, len(t.Columns)-1))
	flex := 0
	for index, column := range t.Columns {
		if column.Width > 0 {
			widths[index] = min(column.Width, remaining)
			remaining -= widths[index]
		} else {
			flex++
		}
	}
	used := 0
	for index, column := range t.Columns {
		if column.Width <= 0 && flex > 0 {
			next := remaining * (used + 1) / flex
			widths[index] = next - remaining*used/flex
			used++
		}
	}
	return widths
}

func (t *Table) drawRow(screen tcell.Screen, y int, values []string, widths []int, style tcell.Style) {
	x := t.Bounds().X
	right := t.Bounds().X + t.Bounds().W
	for index, width := range widths {
		if index > 0 {
			if x < right {
				screen.SetContent(x, y, t.Separator, nil, style)
			}
			x++
		}
		if x >= right {
			break
		}
		if index < len(values) {
			drawText(screen, x, y, min(width, right-x), values[index], style)
		}
		x += width
	}
}

func (t *Table) ensureVisible() {
	t.Selected = clamp(t.Selected, 0, max(0, len(t.Rows)-1))
	height := max(0, t.Bounds().H-1)
	if t.Selected < t.Offset {
		t.Offset = t.Selected
	} else if height > 0 && t.Selected >= t.Offset+height {
		t.Offset = t.Selected - height + 1
	}
	t.Offset = clamp(t.Offset, 0, max(0, len(t.Rows)-max(1, height)))
}

func (t *Table) defaultAction(event *Event) {
	if t.Disabled || len(t.Rows) == 0 {
		return
	}
	previous := t.Selected
	switch event.Kind {
	case KeyDown:
		switch event.KeyCode {
		case tcell.KeyUp:
			t.SetSelected(t.Selected - 1)
		case tcell.KeyDown:
			t.SetSelected(t.Selected + 1)
		case tcell.KeyHome:
			t.SetSelected(0)
		case tcell.KeyEnd:
			t.SetSelected(len(t.Rows) - 1)
		case tcell.KeyPgUp:
			t.SetSelected(t.Selected - max(1, t.Bounds().H-1))
		case tcell.KeyPgDn:
			t.SetSelected(t.Selected + max(1, t.Bounds().H-1))
		case tcell.KeyEnter:
			Dispatch(t, &Event{Kind: Submitted})
		}
	case PointerScroll:
		t.SetSelected(t.Selected + event.DeltaY)
	case Click:
		row := event.Y - t.Bounds().Y - 1
		if row >= 0 && row < t.Bounds().H-1 && t.Offset+row < len(t.Rows) {
			t.SetSelected(t.Offset + row)
			Dispatch(t, &Event{Kind: Submitted})
		}
	}
	if previous != t.Selected {
		Dispatch(t, &Event{Kind: Changed})
	}
}

func columnTitles(columns []TableColumn) []string {
	titles := make([]string, len(columns))
	for index, column := range columns {
		titles[index] = column.Title
	}
	return titles
}
