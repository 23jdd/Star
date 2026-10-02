package widgets

import "github.com/gdamore/tcell/v2"

// ScrollView displays scrollable text with keyboard, mouse-wheel, and optional
// scrollbar support.
type ScrollView struct {
	WidgetBase
	Content        string
	Offset         int
	Wrap           bool
	FollowEnd      bool
	ShowScrollbar  bool
	Style          tcell.Style
	ScrollbarStyle tcell.Style
	Focused        bool
	Disabled       bool
}

// NewScrollView creates a focusable text viewport.
func NewScrollView(content string) *ScrollView {
	view := &ScrollView{
		Content:        content,
		Style:          tcell.StyleDefault,
		ScrollbarStyle: tcell.StyleDefault.Foreground(tcell.ColorGray),
	}
	view.DefaultAction = view.defaultAction
	return view
}

func (v *ScrollView) CanFocus() bool      { return !v.Disabled && !v.Hidden() }
func (v *ScrollView) SetFocus(value bool) { v.Focused = value }
func (v *ScrollView) HandlerEvent(*Event) {}

func (v *ScrollView) Measure(c Constraints) Size {
	text := Text{Content: v.Content, Wrap: v.Wrap}
	size := text.Measure(c)
	if v.ShowScrollbar && size.W < c.MaxW {
		size.W++
	}
	return c.Constrain(size)
}

func (v *ScrollView) Arrange(bounds Rect) {
	v.SetBounds(bounds)
	v.clampOffset()
}

// SetContent replaces the text and updates the valid scroll range.
func (v *ScrollView) SetContent(content string) {
	v.Content = content
	v.clampOffset()
}

// ScrollTo moves to a clamped line offset and disables FollowEnd.
func (v *ScrollView) ScrollTo(offset int) {
	v.Offset = offset
	v.FollowEnd = false
	v.clampOffset()
}

func (v *ScrollView) Render(screen tcell.Screen) {
	if v.Hidden() || v.Bounds().Empty() {
		return
	}
	fill(screen, v.Bounds(), ' ', v.Style)
	contentWidth := v.contentWidth()
	lines := (&Text{Content: v.Content, Wrap: v.Wrap}).lines(contentWidth)
	v.clampOffsetFor(len(lines))
	for row := 0; row < v.Bounds().H && v.Offset+row < len(lines); row++ {
		drawText(screen, v.Bounds().X, v.Bounds().Y+row, contentWidth, lines[v.Offset+row], v.Style)
	}
	if v.ShowScrollbar && len(lines) > v.Bounds().H && v.Bounds().W > 0 {
		x := v.Bounds().X + v.Bounds().W - 1
		for row := 0; row < v.Bounds().H; row++ {
			screen.SetContent(x, v.Bounds().Y+row, '│', nil, v.ScrollbarStyle)
		}
		thumbSize := max(1, v.Bounds().H*v.Bounds().H/len(lines))
		thumbTop := v.Offset * (v.Bounds().H - thumbSize) / max(1, len(lines)-v.Bounds().H)
		for row := thumbTop; row < thumbTop+thumbSize; row++ {
			screen.SetContent(x, v.Bounds().Y+row, '█', nil, v.ScrollbarStyle)
		}
	}
}

func (v *ScrollView) contentWidth() int {
	width := v.Bounds().W
	if v.ShowScrollbar && width > 0 {
		width--
	}
	return max(0, width)
}

func (v *ScrollView) lines() []string {
	return (&Text{Content: v.Content, Wrap: v.Wrap}).lines(v.contentWidth())
}

func (v *ScrollView) clampOffset() { v.clampOffsetFor(len(v.lines())) }

func (v *ScrollView) clampOffsetFor(lineCount int) {
	maximum := max(0, lineCount-v.Bounds().H)
	if v.FollowEnd {
		v.Offset = maximum
	} else {
		v.Offset = clamp(v.Offset, 0, maximum)
	}
}

func (v *ScrollView) defaultAction(event *Event) {
	if v.Disabled {
		return
	}
	delta := 0
	switch event.Kind {
	case PointerScroll:
		delta = event.DeltaY
	case KeyDown:
		switch event.KeyCode {
		case tcell.KeyUp:
			delta = -1
		case tcell.KeyDown:
			delta = 1
		case tcell.KeyPgUp:
			delta = -max(1, v.Bounds().H-1)
		case tcell.KeyPgDn:
			delta = max(1, v.Bounds().H-1)
		case tcell.KeyHome:
			v.Offset, v.FollowEnd = 0, false
		case tcell.KeyEnd:
			v.Offset, v.FollowEnd = len(v.lines()), true
		}
	}
	if delta != 0 {
		v.Offset += delta
		v.FollowEnd = false
	}
	v.clampOffset()
}
