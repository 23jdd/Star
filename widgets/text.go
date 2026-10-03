package widgets

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// Text renders newline-separated text clipped to its arranged bounds.
type Text struct {
	WidgetBase
	Content string
	Style   tcell.Style
	Wrap    bool
	Align   TextAlign
}

// TextAlign controls horizontal alignment inside Text bounds.
type TextAlign uint8

const (
	// AlignLeft aligns text to the left edge.
	AlignLeft TextAlign = iota
	// AlignCenter centers text horizontally.
	AlignCenter
	// AlignRight aligns text to the right edge.
	AlignRight
)

// VerticalAlign controls vertical placement inside an available area.
type VerticalAlign uint8

const (
	// AlignTop places content at the top edge.
	AlignTop VerticalAlign = iota
	// AlignMiddle centres content vertically.
	AlignMiddle
	// AlignBottom places content at the bottom edge.
	AlignBottom
)

// NewText creates left-aligned, unwrapped text.
func NewText(content string) *Text {
	return &Text{Content: content, Style: tcell.StyleDefault}
}

func (t *Text) Measure(constraints Constraints) Size {
	lines := t.lines(constraints.MaxW)
	width := 0
	for _, line := range lines {
		width = max(width, uniseg.StringWidth(line))
	}
	return constraints.Constrain(Size{W: width, H: len(lines)})
}

func (t *Text) Arrange(bounds Rect) { t.SetBounds(bounds) }

func (t *Text) Render(screen tcell.Screen) {
	if t.Hidden() {
		return
	}
	for row, line := range t.lines(t.Bounds().W) {
		if row >= t.Bounds().H {
			break
		}
		x := t.Bounds().X
		lineWidth := uniseg.StringWidth(line)
		if t.Align == AlignCenter {
			x += max(0, (t.Bounds().W-lineWidth)/2)
		} else if t.Align == AlignRight {
			x += max(0, t.Bounds().W-lineWidth)
		}
		drawText(screen, x, t.Bounds().Y+row, t.Bounds().W, line, t.Style)
	}
}

func (t *Text) HandlerEvent(*Event) {}

func (t *Text) lines(width int) []string {
	lines := strings.Split(t.Content, "\n")
	if !t.Wrap || width <= 0 {
		return lines
	}
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			result = append(result, "")
			continue
		}
		graphemes := uniseg.NewGraphemes(line)
		current, currentWidth := "", 0
		for graphemes.Next() {
			cluster := graphemes.Str()
			clusterWidth := max(1, uniseg.StringWidth(cluster))
			if current != "" && currentWidth+clusterWidth > width {
				result = append(result, current)
				current, currentWidth = "", 0
			}
			current += cluster
			currentWidth += clusterWidth
		}
		if current != "" {
			result = append(result, current)
		}
	}
	return result
}
