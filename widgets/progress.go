package widgets

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Progress renders a horizontal progress bar. Value is clamped to [0,1].
type Progress struct {
	WidgetBase
	Value          float64
	Style          tcell.Style
	FilledStyle    tcell.Style
	ShowPercentage bool
}

// NewProgress creates a progress bar with the supplied initial value.
func NewProgress(value float64) *Progress {
	return &Progress{
		Value:       value,
		Style:       tcell.StyleDefault.Foreground(tcell.ColorGray),
		FilledStyle: tcell.StyleDefault.Foreground(tcell.ColorGreen),
	}
}

func (p *Progress) Measure(c Constraints) Size { return c.Constrain(Size{W: 10, H: 1}) }
func (p *Progress) Arrange(bounds Rect)        { p.SetBounds(bounds) }
func (p *Progress) HandlerEvent(*Event)        {}

func (p *Progress) Render(screen tcell.Screen) {
	if p.Hidden() || p.Bounds().Empty() {
		return
	}
	value := p.Value
	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}
	filled := int(value*float64(p.Bounds().W) + 0.5)
	fill(screen, NewRect(p.Bounds().X, p.Bounds().Y, filled, 1), '█', p.FilledStyle)
	if p.ShowPercentage {
		label := fmt.Sprintf("%3.0f%%", value*100)
		x := p.Bounds().X + max(0, (p.Bounds().W-len(label))/2)
		p.drawPercentage(screen, x, label, filled)
	}
}

// drawPercentage keeps the filled section visible behind the label. Text over
// the filled section uses the fill colour as its background; text over the
// remaining track keeps the normal style and therefore adds no background.
func (p *Progress) drawPercentage(screen tcell.Screen, x int, label string, filled int) {
	bounds := p.Bounds()
	screenWidth, screenHeight := screen.Size()
	if bounds.Y < 0 || bounds.Y >= screenHeight {
		return
	}

	for offset, ch := range label {
		cellX := x + offset
		if cellX < bounds.X || cellX >= bounds.X+bounds.W || cellX < 0 || cellX >= screenWidth {
			continue
		}

		// Preserve the background already painted by the parent. Applying the
		// progress style directly would replace it with the terminal default.
		_, _, parentStyle, _ := screen.GetContent(cellX, bounds.Y)
		_, parentBackground, _ := parentStyle.Decompose()
		style := p.Style.Background(parentBackground)
		if cellX < bounds.X+filled {
			style = p.FilledStyle.Reverse(true)
		}
		screen.SetContent(cellX, bounds.Y, ch, nil, style)
	}
}
