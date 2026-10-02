package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

func fill(screen tcell.Screen, bounds Rect, ch rune, style tcell.Style) {
	if bounds.Empty() {
		return
	}
	width, height := screen.Size()
	x0, y0 := max(0, bounds.X), max(0, bounds.Y)
	x1, y1 := min(width, bounds.X+bounds.W), min(height, bounds.Y+bounds.H)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			screen.SetContent(x, y, ch, nil, style)
		}
	}
}

func drawText(screen tcell.Screen, x, y, limit int, value string, style tcell.Style) {
	if limit <= 0 {
		return
	}
	screenWidth, screenHeight := screen.Size()
	if y < 0 || y >= screenHeight {
		return
	}
	used := 0
	graphemes := uniseg.NewGraphemes(value)
	for graphemes.Next() {
		cluster := graphemes.Str()
		clusterWidth := uniseg.StringWidth(cluster)
		if used+clusterWidth > limit {
			break
		}
		runes := []rune(cluster)
		if len(runes) == 0 {
			continue
		}
		cellX := x + used
		if cellX >= 0 && cellX < screenWidth {
			screen.SetContent(cellX, y, runes[0], runes[1:], style)
		}
		used += max(1, clusterWidth)
	}
}
