package widgets

import "github.com/gdamore/tcell/v2"

// Spacer reserves at least the requested number of cells.
type Spacer struct {
	WidgetBase
	Width, Height int
}

// NewSpacer creates an empty widget with a preferred size.
func NewSpacer(width, height int) *Spacer {
	return &Spacer{Width: max(0, width), Height: max(0, height)}
}
func (s *Spacer) Measure(c Constraints) Size { return c.Constrain(Size{W: s.Width, H: s.Height}) }
func (s *Spacer) Arrange(bounds Rect)        { s.SetBounds(bounds) }
func (s *Spacer) Render(tcell.Screen)        {}
func (s *Spacer) HandlerEvent(*Event)        {}
