package widgets

import "github.com/gdamore/tcell/v2"

// Modal centers one child over a styled backdrop. Use the Modal as the
// Router root while open to trap hit testing and focus inside it.
type Modal struct {
	WidgetBase
	Child            Widget
	Title            string
	Width            int
	Height           int
	Padding          Insets
	Border           Border
	BackdropStyle    tcell.Style
	Style            tcell.Style
	BorderStyle      tcell.Style
	DismissOnEscape  bool
	DismissOnOutside bool
	panelBounds      Rect
}

// NewModal creates a centered modal with a rounded border.
func NewModal(child Widget) *Modal {
	m := &Modal{
		Child:            child,
		Width:            50,
		Height:           12,
		Padding:          UniformInsets(1),
		Border:           RoundedBorder,
		BackdropStyle:    tcell.StyleDefault.Background(tcell.ColorBlack),
		Style:            tcell.StyleDefault,
		BorderStyle:      tcell.StyleDefault.Bold(true),
		DismissOnEscape:  true,
		DismissOnOutside: true,
	}
	m.attach()
	m.DefaultAction = m.defaultAction
	return m
}

func (m *Modal) EventChildren() []Widget {
	if m.Child == nil {
		return nil
	}
	return []Widget{m.Child}
}

func (m *Modal) Measure(c Constraints) Size {
	width, height := m.Width, m.Height
	if width <= 0 || height <= 0 {
		child := Size{}
		if m.Child != nil {
			child = m.Child.Measure(Loose(c.MaxW, c.MaxH))
		}
		if width <= 0 {
			width = child.W + max(0, m.Padding.Left) + max(0, m.Padding.Right) + 2
		}
		if height <= 0 {
			height = child.H + max(0, m.Padding.Top) + max(0, m.Padding.Bottom) + 2
		}
	}
	return c.Constrain(Size{W: width, H: height})
}

func (m *Modal) Arrange(bounds Rect) {
	m.SetBounds(bounds)
	m.attach()
	width := min(bounds.W, max(2, m.Width))
	height := min(bounds.H, max(2, m.Height))
	if m.Width <= 0 || m.Height <= 0 {
		preferred := m.Measure(Loose(bounds.W, bounds.H))
		if m.Width <= 0 {
			width = preferred.W
		}
		if m.Height <= 0 {
			height = preferred.H
		}
	}
	m.panelBounds = NewRect(bounds.X+(bounds.W-width)/2, bounds.Y+(bounds.H-height)/2, width, height)
	if m.Child != nil {
		x := m.panelBounds.X + 1 + max(0, m.Padding.Left)
		y := m.panelBounds.Y + 1 + max(0, m.Padding.Top)
		childWidth := m.panelBounds.W - 2 - max(0, m.Padding.Left) - max(0, m.Padding.Right)
		childHeight := m.panelBounds.H - 2 - max(0, m.Padding.Top) - max(0, m.Padding.Bottom)
		m.Child.Arrange(NewRect(x, y, childWidth, childHeight))
	}
}

func (m *Modal) Render(screen tcell.Screen) {
	if m.Hidden() || m.Bounds().Empty() {
		return
	}
	fill(screen, m.Bounds(), ' ', m.BackdropStyle)
	fill(screen, m.panelBounds, ' ', m.Style)
	drawBorder(screen, m.panelBounds, m.Border, m.BorderStyle)
	if m.Title != "" && m.panelBounds.W > 4 {
		drawText(screen, m.panelBounds.X+2, m.panelBounds.Y, m.panelBounds.W-4, m.Title, m.BorderStyle)
	}
	if m.Child != nil {
		m.Child.Render(screen)
	}
}

func (m *Modal) HandlerEvent(event *Event) {
	if event.Phase != Capture && m.DismissOnEscape && event.Kind == KeyDown && event.KeyCode == tcell.KeyEscape {
		event.PreventDefault()
		Dispatch(m, &Event{Kind: Dismissed})
	}
}

// PanelBounds returns the centered panel rectangle after Arrange.
func (m *Modal) PanelBounds() Rect { return m.panelBounds }

func (m *Modal) attach() {
	if base := widgetBase(m.Child); base != nil {
		base.setParent(m)
	}
}

func (m *Modal) defaultAction(event *Event) {
	if m.DismissOnOutside && event.Kind == Click && !m.panelBounds.Contains(event.X, event.Y) {
		Dispatch(m, &Event{Kind: Dismissed})
	}
}
