package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestProgressPercentageFollowsBarBackground(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(10, 1)

	progress := NewProgress(0.5)
	progress.ShowPercentage = true
	progress.Style = tcell.StyleDefault.Foreground(tcell.ColorGray)
	progress.FilledStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	progress.Arrange(NewRect(0, 0, 10, 1))
	parentStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	for x := 0; x < 10; x++ {
		screen.SetContent(x, 0, ' ', nil, parentStyle)
	}
	untouchedStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	screen.SetContent(7, 0, 'x', nil, untouchedStyle)
	progress.Render(screen)

	// " 50%" is centred at columns 3..6. Columns 3 and 4 cover the filled
	// section, while columns 5 and 6 cover the remaining track.
	wantRunes := []rune{' ', '5', '0', '%'}
	for index, wantRune := range wantRunes {
		x := 3 + index
		gotRune, _, gotStyle, _ := screen.GetContent(x, 0)
		if gotRune != wantRune {
			t.Fatalf("cell %d rune=%q want=%q", x, gotRune, wantRune)
		}

		wantStyle := progress.Style.Background(tcell.ColorDarkBlue)
		if x < 5 {
			wantStyle = progress.FilledStyle.Reverse(true)
		}
		if gotStyle != wantStyle {
			t.Fatalf("cell %d style=%#v want=%#v", x, gotStyle, wantStyle)
		}
	}

	if _, _, style, _ := screen.GetContent(2, 0); style != progress.FilledStyle {
		t.Fatalf("filled cell outside label style=%#v want=%#v", style, progress.FilledStyle)
	}
	if ch, _, style, _ := screen.GetContent(7, 0); ch != 'x' || style != untouchedStyle {
		t.Fatalf("unfilled cell was changed: rune=%q style=%#v", ch, style)
	}
}
