package widgets

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// TextArea is a retained, multiline text editor with keyboard scrolling.
type TextArea struct {
	WidgetBase
	Value       string
	Placeholder string
	Style       tcell.Style
	FocusStyle  tcell.Style
	CursorStyle tcell.Style
	Focused     bool
	Disabled    bool
	CursorRow   int
	CursorCol   int // rune index
	OffsetRow   int
	OffsetCol   int // rune index
}

// NewTextArea creates a multiline editor with the cursor at the end.
func NewTextArea(value string) *TextArea {
	a := &TextArea{Value: normalizeLines(value), Style: tcell.StyleDefault, FocusStyle: tcell.StyleDefault, CursorStyle: tcell.StyleDefault.Reverse(true)}
	lines := a.lines()
	a.CursorRow = len(lines) - 1
	a.CursorCol = len([]rune(lines[a.CursorRow]))
	a.DefaultAction = a.defaultAction
	return a
}

// SetValue replaces the editor contents and clamps cursor and scroll state.
func (a *TextArea) SetValue(value string) {
	a.Value = normalizeLines(value)
	a.clampCursor()
	a.ensureVisible()
}

func (a *TextArea) CanFocus() bool      { return !a.Disabled && !a.Hidden() }
func (a *TextArea) SetFocus(value bool) { a.Focused = value }
func (a *TextArea) HandlerEvent(*Event) {}

func (a *TextArea) Measure(c Constraints) Size {
	width := 1
	lines := a.lines()
	for _, line := range lines {
		width = max(width, uniseg.StringWidth(line))
	}
	return c.Constrain(Size{W: width, H: max(1, len(lines))})
}

func (a *TextArea) Arrange(bounds Rect) {
	a.SetBounds(bounds)
	a.ensureVisible()
}

func (a *TextArea) Render(screen tcell.Screen) {
	if a.Hidden() || a.Bounds().Empty() {
		return
	}
	style := a.Style
	if a.Focused {
		style = a.FocusStyle
	}
	fill(screen, a.Bounds(), ' ', style)
	lines := a.lines()
	if a.Value == "" && !a.Focused {
		lines = strings.Split(a.Placeholder, "\n")
	}
	for row := 0; row < a.Bounds().H && a.OffsetRow+row < len(lines); row++ {
		runes := []rune(lines[a.OffsetRow+row])
		start := min(a.OffsetCol, len(runes))
		drawText(screen, a.Bounds().X, a.Bounds().Y+row, a.Bounds().W, string(runes[start:]), style)
	}
	if a.Focused {
		lineRunes := []rune(lines[a.CursorRow])
		prefix := string(lineRunes[min(a.OffsetCol, len(lineRunes)):min(a.CursorCol, len(lineRunes))])
		x := a.Bounds().X + uniseg.StringWidth(prefix)
		y := a.Bounds().Y + a.CursorRow - a.OffsetRow
		if x >= a.Bounds().X && x < a.Bounds().X+a.Bounds().W && y >= a.Bounds().Y && y < a.Bounds().Y+a.Bounds().H {
			cursorRune := ' '
			if a.CursorCol < len(lineRunes) {
				cursorRune = lineRunes[a.CursorCol]
			}
			screen.SetContent(x, y, cursorRune, nil, a.CursorStyle)
		}
	}
}

func (a *TextArea) lines() []string {
	lines := strings.Split(a.Value, "\n")
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func (a *TextArea) clampCursor() {
	lines := a.lines()
	a.CursorRow = clamp(a.CursorRow, 0, len(lines)-1)
	a.CursorCol = clamp(a.CursorCol, 0, len([]rune(lines[a.CursorRow])))
}

func (a *TextArea) ensureVisible() {
	a.clampCursor()
	if a.CursorRow < a.OffsetRow {
		a.OffsetRow = a.CursorRow
	} else if a.Bounds().H > 0 && a.CursorRow >= a.OffsetRow+a.Bounds().H {
		a.OffsetRow = a.CursorRow - a.Bounds().H + 1
	}
	if a.CursorCol < a.OffsetCol {
		a.OffsetCol = a.CursorCol
	}
	line := []rune(a.lines()[a.CursorRow])
	a.OffsetCol = clamp(a.OffsetCol, 0, len(line))
	if a.Bounds().W > 0 {
		for a.OffsetCol < a.CursorCol && uniseg.StringWidth(string(line[a.OffsetCol:a.CursorCol])) >= a.Bounds().W {
			a.OffsetCol++
		}
	}
	a.OffsetRow, a.OffsetCol = max(0, a.OffsetRow), max(0, a.OffsetCol)
}

func (a *TextArea) defaultAction(event *Event) {
	if a.Disabled || event.Kind != KeyDown {
		return
	}
	lines := a.lines()
	a.clampCursor()
	changed := false
	switch event.KeyCode {
	case tcell.KeyLeft:
		if a.CursorCol > 0 {
			a.CursorCol--
		} else if a.CursorRow > 0 {
			a.CursorRow--
			a.CursorCol = len([]rune(lines[a.CursorRow]))
		}
	case tcell.KeyRight:
		if a.CursorCol < len([]rune(lines[a.CursorRow])) {
			a.CursorCol++
		} else if a.CursorRow+1 < len(lines) {
			a.CursorRow++
			a.CursorCol = 0
		}
	case tcell.KeyUp:
		a.CursorRow = max(0, a.CursorRow-1)
		a.CursorCol = min(a.CursorCol, len([]rune(lines[a.CursorRow])))
	case tcell.KeyDown:
		a.CursorRow = min(len(lines)-1, a.CursorRow+1)
		a.CursorCol = min(a.CursorCol, len([]rune(lines[a.CursorRow])))
	case tcell.KeyHome, tcell.KeyCtrlA:
		a.CursorCol = 0
	case tcell.KeyEnd, tcell.KeyCtrlE:
		a.CursorCol = len([]rune(lines[a.CursorRow]))
	case tcell.KeyPgUp:
		a.CursorRow = max(0, a.CursorRow-max(1, a.Bounds().H))
	case tcell.KeyPgDn:
		a.CursorRow = min(len(lines)-1, a.CursorRow+max(1, a.Bounds().H))
	case tcell.KeyEnter:
		runes := []rune(lines[a.CursorRow])
		before, after := string(runes[:a.CursorCol]), string(runes[a.CursorCol:])
		lines[a.CursorRow] = before
		lines = append(lines, "")
		copy(lines[a.CursorRow+2:], lines[a.CursorRow+1:])
		lines[a.CursorRow+1] = after
		a.CursorRow++
		a.CursorCol = 0
		changed = true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if a.CursorCol > 0 {
			runes := []rune(lines[a.CursorRow])
			lines[a.CursorRow] = string(append(runes[:a.CursorCol-1], runes[a.CursorCol:]...))
			a.CursorCol--
			changed = true
		} else if a.CursorRow > 0 {
			previous := len([]rune(lines[a.CursorRow-1]))
			lines[a.CursorRow-1] += lines[a.CursorRow]
			lines = append(lines[:a.CursorRow], lines[a.CursorRow+1:]...)
			a.CursorRow--
			a.CursorCol = previous
			changed = true
		}
	case tcell.KeyDelete, tcell.KeyCtrlD:
		runes := []rune(lines[a.CursorRow])
		if a.CursorCol < len(runes) {
			lines[a.CursorRow] = string(append(runes[:a.CursorCol], runes[a.CursorCol+1:]...))
			changed = true
		} else if a.CursorRow+1 < len(lines) {
			lines[a.CursorRow] += lines[a.CursorRow+1]
			lines = append(lines[:a.CursorRow+1], lines[a.CursorRow+2:]...)
			changed = true
		}
	case tcell.KeyRune:
		if event.Rune != 0 && event.Modifiers&(tcell.ModCtrl|tcell.ModAlt) == 0 {
			runes := []rune(lines[a.CursorRow])
			runes = append(runes, 0)
			copy(runes[a.CursorCol+1:], runes[a.CursorCol:])
			runes[a.CursorCol] = event.Rune
			lines[a.CursorRow] = string(runes)
			a.CursorCol++
			changed = true
		}
	}
	if changed {
		a.Value = strings.Join(lines, "\n")
		Dispatch(a, &Event{Kind: Changed})
	}
	a.ensureVisible()
}

func normalizeLines(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n")
}
