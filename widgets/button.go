package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// Button is a focusable label activated by Router through mouse clicks,
// Enter, or Space. Register Click listeners with On.
type Button struct {
	WidgetBase
	Content       string
	Style         tcell.Style
	FocusedStyle  tcell.Style
	DisabledStyle tcell.Style
	Focused       bool
	Disabled      bool
	Padding       int
}

// NewButton creates an enabled button with default terminal styles.
func NewButton(content string) *Button {
	return &Button{
		Content:       content,
		Style:         tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue),
		FocusedStyle:  tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorLightBlue).Bold(true),
		DisabledStyle: tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack),
		Padding:       1,
	}
}

func (b *Button) Measure(constraints Constraints) Size {
	return constraints.Constrain(Size{W: uniseg.StringWidth(b.Content) + max(0, b.Padding)*2, H: 1})
}

func (b *Button) Arrange(bounds Rect) { b.SetBounds(bounds) }

func (b *Button) Render(screen tcell.Screen) {
	if b.Hidden() || b.Bounds().Empty() {
		return
	}
	style := b.Style
	if b.Disabled {
		style = b.DisabledStyle
	} else if b.Focused {
		style = b.FocusedStyle
	}
	fill(screen, b.Bounds(), ' ', style)
	padding := max(0, b.Padding)
	drawText(screen, b.Bounds().X+padding, b.Bounds().Y, b.Bounds().W-padding*2, b.Content, style)
}

func (b *Button) HandlerEvent(*Event) {}

func (b *Button) CanFocus() bool            { return !b.Disabled && !b.Hidden() }
func (b *Button) SetFocus(focused bool)     { b.Focused = focused }
func (b *Button) KeyboardActivatable() bool { return !b.Disabled }
