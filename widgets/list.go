package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// List is a keyboard- and mouse-selectable list of strings.
type List struct {
	WidgetBase
	Items         []string
	Selected      int
	Offset        int
	Style         tcell.Style
	SelectedStyle tcell.Style
	Focused       bool
	Disabled      bool
}

// NewList creates a selectable copy of items.
func NewList(items ...string) *List {
	list := &List{
		Items:         append([]string(nil), items...),
		Style:         tcell.StyleDefault,
		SelectedStyle: tcell.StyleDefault.Reverse(true),
	}
	list.DefaultAction = list.defaultAction
	return list
}

func (l *List) CanFocus() bool      { return !l.Disabled && !l.Hidden() }
func (l *List) SetFocus(value bool) { l.Focused = value }
func (l *List) Arrange(bounds Rect) {
	l.SetBounds(bounds)
	l.ensureVisible()
}
func (l *List) HandlerEvent(*Event) {}

func (l *List) Measure(c Constraints) Size {
	width := 0
	for _, item := range l.Items {
		width = max(width, uniseg.StringWidth(item))
	}
	return c.Constrain(Size{W: width, H: len(l.Items)})
}

func (l *List) Render(screen tcell.Screen) {
	if l.Hidden() || l.Bounds().Empty() {
		return
	}
	fill(screen, l.Bounds(), ' ', l.Style)
	for row := 0; row < l.Bounds().H && l.Offset+row < len(l.Items); row++ {
		index := l.Offset + row
		style := l.Style
		if index == l.Selected {
			style = l.SelectedStyle
			fill(screen, NewRect(l.Bounds().X, l.Bounds().Y+row, l.Bounds().W, 1), ' ', style)
		}
		drawText(screen, l.Bounds().X, l.Bounds().Y+row, l.Bounds().W, l.Items[index], style)
	}
}

// SetSelected clamps and selects index, returning true when it changed.
func (l *List) SetSelected(index int) bool {
	if len(l.Items) == 0 {
		l.Selected, l.Offset = 0, 0
		return false
	}
	index = clamp(index, 0, len(l.Items)-1)
	if index == l.Selected {
		return false
	}
	l.Selected = index
	l.ensureVisible()
	return true
}

func (l *List) ensureVisible() {
	l.Selected = clamp(l.Selected, 0, max(0, len(l.Items)-1))
	if l.Selected < l.Offset {
		l.Offset = l.Selected
	}
	if height := l.Bounds().H; height > 0 && l.Selected >= l.Offset+height {
		l.Offset = l.Selected - height + 1
	}
	l.Offset = clamp(l.Offset, 0, max(0, len(l.Items)-max(1, l.Bounds().H)))
}

func (l *List) defaultAction(event *Event) {
	if l.Disabled || len(l.Items) == 0 {
		return
	}
	previous := l.Selected
	switch event.Kind {
	case KeyDown:
		switch event.KeyCode {
		case tcell.KeyUp:
			l.SetSelected(l.Selected - 1)
		case tcell.KeyDown:
			l.SetSelected(l.Selected + 1)
		case tcell.KeyHome:
			l.SetSelected(0)
		case tcell.KeyEnd:
			l.SetSelected(len(l.Items) - 1)
		case tcell.KeyPgUp:
			l.SetSelected(l.Selected - max(1, l.Bounds().H))
		case tcell.KeyPgDn:
			l.SetSelected(l.Selected + max(1, l.Bounds().H))
		case tcell.KeyEnter:
			Dispatch(l, &Event{Kind: Submitted})
		}
		if event.Rune == 'j' {
			l.SetSelected(l.Selected + 1)
		} else if event.Rune == 'k' {
			l.SetSelected(l.Selected - 1)
		}
	case PointerScroll:
		l.SetSelected(l.Selected + event.DeltaY)
	case Click:
		row := event.Y - l.Bounds().Y
		if row >= 0 && row < l.Bounds().H && l.Offset+row < len(l.Items) {
			l.SetSelected(l.Offset + row)
			Dispatch(l, &Event{Kind: Submitted})
		}
	}
	if previous != l.Selected {
		Dispatch(l, &Event{Kind: Changed})
	}
}
