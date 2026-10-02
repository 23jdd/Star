package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// Input is a single-line editable text field. Retain the same Input instance
// across View calls when using its built-in editing behavior.
type Input struct {
	WidgetBase
	Value       string
	Placeholder string
	Style       tcell.Style
	FocusStyle  tcell.Style
	CursorStyle tcell.Style
	Focused     bool
	Disabled    bool
	Cursor      int // rune index
}

// NewInput creates a single-line input with its cursor at the end.
func NewInput(value string) *Input {
	input := &Input{
		Value:       value,
		Style:       tcell.StyleDefault,
		FocusStyle:  tcell.StyleDefault,
		CursorStyle: tcell.StyleDefault.Reverse(true),
		Cursor:      len([]rune(value)),
	}
	input.DefaultAction = input.defaultAction
	return input
}

// SetValue replaces the text and clamps the cursor to a valid rune index.
func (i *Input) SetValue(value string) {
	i.Value = value
	i.Cursor = clamp(i.Cursor, 0, len([]rune(value)))
}

func (i *Input) CanFocus() bool      { return !i.Disabled && !i.Hidden() }
func (i *Input) SetFocus(value bool) { i.Focused = value }
func (i *Input) Arrange(bounds Rect) { i.SetBounds(bounds) }
func (i *Input) HandlerEvent(*Event) {}
func (i *Input) Measure(c Constraints) Size {
	width := max(uniseg.StringWidth(i.Value), uniseg.StringWidth(i.Placeholder))
	return c.Constrain(Size{W: max(1, width), H: 1})
}

func (i *Input) Render(screen tcell.Screen) {
	if i.Hidden() || i.Bounds().Empty() {
		return
	}
	style := i.Style
	if i.Focused {
		style = i.FocusStyle
	}
	fill(screen, i.Bounds(), ' ', style)
	value := i.Value
	if value == "" && !i.Focused {
		value = i.Placeholder
	}
	runes := []rune(i.Value)
	cursor := clamp(i.Cursor, 0, len(runes))
	prefixWidth := uniseg.StringWidth(string(runes[:cursor]))
	offset := max(0, prefixWidth-i.Bounds().W+1)
	drawText(screen, i.Bounds().X-offset, i.Bounds().Y, i.Bounds().W+offset, value, style)
	if i.Focused {
		x := i.Bounds().X + prefixWidth - offset
		if x >= i.Bounds().X && x < i.Bounds().X+i.Bounds().W {
			cursorRune := ' '
			if cursor < len(runes) {
				cursorRune = runes[cursor]
			}
			screen.SetContent(x, i.Bounds().Y, cursorRune, nil, i.CursorStyle)
		}
	}
}

func (i *Input) defaultAction(event *Event) {
	if i.Disabled || event.Kind != KeyDown {
		return
	}
	runes := []rune(i.Value)
	i.Cursor = clamp(i.Cursor, 0, len(runes))
	changed := false
	switch event.KeyCode {
	case tcell.KeyLeft:
		i.Cursor = max(0, i.Cursor-1)
	case tcell.KeyRight:
		i.Cursor = min(len(runes), i.Cursor+1)
	case tcell.KeyHome, tcell.KeyCtrlA:
		i.Cursor = 0
	case tcell.KeyEnd, tcell.KeyCtrlE:
		i.Cursor = len(runes)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if i.Cursor > 0 {
			runes = append(runes[:i.Cursor-1], runes[i.Cursor:]...)
			i.Cursor--
			changed = true
		}
	case tcell.KeyDelete, tcell.KeyCtrlD:
		if i.Cursor < len(runes) {
			runes = append(runes[:i.Cursor], runes[i.Cursor+1:]...)
			changed = true
		}
	case tcell.KeyEnter:
		Dispatch(i, &Event{Kind: Submitted})
	case tcell.KeyRune:
		if event.Rune != 0 && event.Modifiers&(tcell.ModCtrl|tcell.ModAlt) == 0 {
			runes = append(runes, 0)
			copy(runes[i.Cursor+1:], runes[i.Cursor:])
			runes[i.Cursor] = event.Rune
			i.Cursor++
			changed = true
		}
	}
	if changed {
		i.Value = string(runes)
		Dispatch(i, &Event{Kind: Changed})
	}
}
