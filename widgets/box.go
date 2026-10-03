package widgets

import "github.com/gdamore/tcell/v2"

// Insets defines top, right, bottom, and left padding in terminal cells.
type Insets struct{ Top, Right, Bottom, Left int }

// UniformInsets creates equal padding on every side.
func UniformInsets(value int) Insets {
	value = max(0, value)
	return Insets{Top: value, Right: value, Bottom: value, Left: value}
}

// Border contains the eight runes used to draw a rectangular border.
type Border struct {
	Top, Bottom, Left, Right rune
	TopLeft, TopRight        rune
	BottomLeft, BottomRight  rune
}

var (
	// RoundedBorder draws rounded corners.
	RoundedBorder = Border{Top: '─', Bottom: '─', Left: '│', Right: '│', TopLeft: '╭', TopRight: '╮', BottomLeft: '╰', BottomRight: '╯'}
	// SquareBorder draws square single-line corners.
	SquareBorder = Border{Top: '─', Bottom: '─', Left: '│', Right: '│', TopLeft: '┌', TopRight: '┐', BottomLeft: '└', BottomRight: '┘'}
	// DoubleBorder draws a double-line border.
	DoubleBorder = Border{Top: '═', Bottom: '═', Left: '║', Right: '║', TopLeft: '╔', TopRight: '╗', BottomLeft: '╚', BottomRight: '╝'}
)

// Box decorates a single child with padding, background, border, and title.
type Box struct {
	WidgetBase
	Child       Widget
	Padding     Insets
	Border      *Border
	Style       tcell.Style
	BorderStyle tcell.Style
	Title       string
	// FitContent gives Child its measured size instead of filling the inner
	// area. Align and VerticalAlign position that measured rectangle.
	FitContent    bool
	Align         TextAlign
	VerticalAlign VerticalAlign
}

// NewBox creates an undecorated single-child box.
func NewBox(child Widget) *Box {
	box := &Box{
		Child:         child,
		Style:         tcell.StyleDefault,
		BorderStyle:   tcell.StyleDefault,
		Align:         AlignCenter,
		VerticalAlign: AlignMiddle,
	}
	box.attach()
	return box
}

func (b *Box) attach() {
	if base := widgetBase(b.Child); base != nil {
		base.setParent(b)
	}
}

func (b *Box) EventChildren() []Widget {
	if b.Child == nil {
		return nil
	}
	return []Widget{b.Child}
}

func (b *Box) Measure(constraints Constraints) Size {
	insetW := max(0, b.Padding.Left) + max(0, b.Padding.Right)
	insetH := max(0, b.Padding.Top) + max(0, b.Padding.Bottom)
	if b.Border != nil {
		insetW += 2
		insetH += 2
	}
	size := Size{W: insetW, H: insetH}
	if b.Child != nil {
		childConstraints := Loose(max(0, constraints.MaxW-insetW), max(0, constraints.MaxH-insetH))
		child := b.Child.Measure(childConstraints)
		size.W += child.W
		size.H += child.H
	}
	return constraints.Constrain(size)
}

func (b *Box) Arrange(bounds Rect) {
	b.SetBounds(bounds)
	b.attach()
	if b.Child == nil {
		return
	}
	x, y, width, height := bounds.X, bounds.Y, bounds.W, bounds.H
	if b.Border != nil {
		x, y, width, height = x+1, y+1, width-2, height-2
	}
	x += max(0, b.Padding.Left)
	y += max(0, b.Padding.Top)
	width -= max(0, b.Padding.Left) + max(0, b.Padding.Right)
	height -= max(0, b.Padding.Top) + max(0, b.Padding.Bottom)
	width, height = max(0, width), max(0, height)
	if b.FitContent && width > 0 && height > 0 {
		size := b.Child.Measure(Loose(width, height))
		childWidth, childHeight := min(width, size.W), min(height, size.H)
		switch b.Align {
		case AlignCenter:
			x += (width - childWidth) / 2
		case AlignRight:
			x += width - childWidth
		}
		switch b.VerticalAlign {
		case AlignMiddle:
			y += (height - childHeight) / 2
		case AlignBottom:
			y += height - childHeight
		}
		width, height = childWidth, childHeight
	}
	b.Child.Arrange(NewRect(x, y, width, height))
}

func (b *Box) Render(screen tcell.Screen) {
	if b.Hidden() || b.Bounds().Empty() {
		return
	}
	fill(screen, b.Bounds(), ' ', b.Style)
	if b.Border != nil {
		drawBorder(screen, b.Bounds(), *b.Border, b.BorderStyle)
		if b.Title != "" && b.Bounds().W > 4 {
			drawText(screen, b.Bounds().X+2, b.Bounds().Y, b.Bounds().W-4, b.Title, b.BorderStyle)
		}
	}
	if b.Child != nil {
		b.Child.Render(screen)
	}
}

func (b *Box) HandlerEvent(*Event) {}

func drawBorder(screen tcell.Screen, bounds Rect, border Border, style tcell.Style) {
	if bounds.W < 2 || bounds.H < 2 {
		return
	}
	x0, y0 := bounds.X, bounds.Y
	x1, y1 := bounds.X+bounds.W-1, bounds.Y+bounds.H-1
	for x := x0 + 1; x < x1; x++ {
		screen.SetContent(x, y0, border.Top, nil, style)
		screen.SetContent(x, y1, border.Bottom, nil, style)
	}
	for y := y0 + 1; y < y1; y++ {
		screen.SetContent(x0, y, border.Left, nil, style)
		screen.SetContent(x1, y, border.Right, nil, style)
	}
	screen.SetContent(x0, y0, border.TopLeft, nil, style)
	screen.SetContent(x1, y0, border.TopRight, nil, style)
	screen.SetContent(x0, y1, border.BottomLeft, nil, style)
	screen.SetContent(x1, y1, border.BottomRight, nil, style)
}
