package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestButtonContentAlignment(t *testing.T) {
	tests := []struct {
		name  string
		align TextAlign
		wantX int
	}{
		{name: "left", align: AlignLeft, wantX: 1},
		{name: "center", align: AlignCenter, wantX: 5},
		{name: "right", align: AlignRight, wantX: 9},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			screen := tcell.NewSimulationScreen("UTF-8")
			if err := screen.Init(); err != nil {
				t.Fatal(err)
			}
			defer screen.Fini()
			screen.SetSize(12, 1)

			button := NewButton("OK")
			button.Align = test.align
			button.Arrange(NewRect(0, 0, 12, 1))
			button.Render(screen)

			main, _, _, _ := screen.GetContent(test.wantX, 0)
			if main != 'O' {
				t.Fatalf("cell %d=%q want O", test.wantX, main)
			}
		})
	}
}

func TestNewButtonIsCenteredByDefault(t *testing.T) {
	if button := NewButton("OK"); button.Align != AlignCenter {
		t.Fatalf("align=%v want AlignCenter", button.Align)
	}
}
