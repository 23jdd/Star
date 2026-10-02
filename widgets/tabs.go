package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// Tab associates a title with one content widget.
type Tab struct {
	Title   string
	Content Widget
}

// Tabs displays a one-line tab strip and arranges only the active content.
type Tabs struct {
	WidgetBase
	Tabs          []Tab
	Active        int
	Gap           int
	Style         tcell.Style
	ActiveStyle   tcell.Style
	DisabledStyle tcell.Style
	Focused       bool
	Disabled      bool
}

// NewTabs copies tabs and selects the first tab.
func NewTabs(tabs ...Tab) *Tabs {
	t := &Tabs{Tabs: append([]Tab(nil), tabs...), Gap: 1, Style: tcell.StyleDefault, ActiveStyle: tcell.StyleDefault.Reverse(true), DisabledStyle: tcell.StyleDefault.Foreground(tcell.ColorGray)}
	t.attach()
	t.DefaultAction = t.defaultAction
	return t
}

func (t *Tabs) CanFocus() bool      { return !t.Disabled && !t.Hidden() }
func (t *Tabs) SetFocus(value bool) { t.Focused = value }
func (t *Tabs) HandlerEvent(*Event) {}

// SetActive clamps and selects a tab, returning true when it changed.
func (t *Tabs) SetActive(index int) bool {
	if len(t.Tabs) == 0 {
		t.Active = 0
		return false
	}
	index = clamp(index, 0, len(t.Tabs)-1)
	if index == t.Active {
		return false
	}
	t.Active = index
	t.attach()
	return true
}

func (t *Tabs) EventChildren() []Widget {
	if t.Disabled || len(t.Tabs) == 0 {
		return nil
	}
	if content := t.activeContent(); content != nil {
		return []Widget{content}
	}
	return nil
}

func (t *Tabs) activeContent() Widget {
	if len(t.Tabs) == 0 {
		return nil
	}
	t.Active = clamp(t.Active, 0, len(t.Tabs)-1)
	return t.Tabs[t.Active].Content
}

func (t *Tabs) Measure(c Constraints) Size {
	width := 0
	for index, tab := range t.Tabs {
		if index > 0 {
			width += max(0, t.Gap)
		}
		width += uniseg.StringWidth(tab.Title) + 2
	}
	height := 1
	if content := t.activeContent(); content != nil {
		size := content.Measure(Loose(c.MaxW, max(0, c.MaxH-1)))
		width, height = max(width, size.W), height+size.H
	}
	return c.Constrain(Size{W: width, H: height})
}

func (t *Tabs) Arrange(bounds Rect) {
	t.SetBounds(bounds)
	t.attach()
	if content := t.activeContent(); content != nil {
		content.Arrange(NewRect(bounds.X, bounds.Y+1, bounds.W, bounds.H-1))
	}
}

func (t *Tabs) Render(screen tcell.Screen) {
	if t.Hidden() || t.Bounds().Empty() {
		return
	}
	fill(screen, t.Bounds(), ' ', t.Style)
	x := t.Bounds().X
	for index, tab := range t.Tabs {
		if index > 0 {
			x += max(0, t.Gap)
		}
		label := " " + tab.Title + " "
		style := t.Style
		if t.Disabled {
			style = t.DisabledStyle
		} else if index == t.Active {
			style = t.ActiveStyle
		}
		drawText(screen, x, t.Bounds().Y, max(0, t.Bounds().X+t.Bounds().W-x), label, style)
		x += uniseg.StringWidth(label)
	}
	if content := t.activeContent(); content != nil {
		content.Render(screen)
	}
}

func (t *Tabs) attach() {
	for _, tab := range t.Tabs {
		if base := widgetBase(tab.Content); base != nil {
			base.setParent(t)
		}
	}
}

func (t *Tabs) defaultAction(event *Event) {
	if t.Disabled || len(t.Tabs) == 0 {
		return
	}
	previous := t.Active
	switch event.Kind {
	case KeyDown:
		switch event.KeyCode {
		case tcell.KeyLeft:
			t.Active = (t.Active - 1 + len(t.Tabs)) % len(t.Tabs)
		case tcell.KeyRight:
			t.Active = (t.Active + 1) % len(t.Tabs)
		case tcell.KeyHome:
			t.Active = 0
		case tcell.KeyEnd:
			t.Active = len(t.Tabs) - 1
		}
	case Click:
		x := t.Bounds().X
		for index, tab := range t.Tabs {
			if index > 0 {
				x += max(0, t.Gap)
			}
			width := uniseg.StringWidth(tab.Title) + 2
			if event.Y == t.Bounds().Y && event.X >= x && event.X < x+width {
				t.Active = index
				break
			}
			x += width
		}
	}
	if previous != t.Active {
		t.attach()
		Dispatch(t, &Event{Kind: Changed})
	}
}
