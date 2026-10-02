package widgets

import (
	"image"
	"image/color"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestImageMeasureUsesHalfCells(t *testing.T) {
	widget := NewImage(image.NewRGBA(image.Rect(0, 0, 7, 5)))
	if got, want := widget.Measure(Loose(20, 20)), (Size{W: 7, H: 3}); got != want {
		t.Fatalf("size=%#v want=%#v", got, want)
	}
}

func TestImageRendersTwoPixelsPerCell(t *testing.T) {
	screen := newImageTestScreen(t, 1, 1)
	defer screen.Fini()

	source := image.NewNRGBA(image.Rect(0, 0, 1, 2))
	source.Set(0, 0, color.NRGBA{R: 255, A: 255})
	source.Set(0, 1, color.NRGBA{B: 255, A: 255})
	widget := NewImage(source)
	widget.Arrange(NewRect(0, 0, 1, 1))
	widget.Render(screen)

	main, _, style, _ := screen.GetContent(0, 0)
	foreground, background, _ := style.Decompose()
	if main != '▀' || foreground != tcell.NewRGBColor(255, 0, 0) || background != tcell.NewRGBColor(0, 0, 255) {
		t.Fatalf("cell=%q foreground=%v background=%v", main, foreground, background)
	}
}

func TestImageTransparencyPreservesParentBackground(t *testing.T) {
	screen := newImageTestScreen(t, 1, 1)
	defer screen.Fini()
	parent := tcell.StyleDefault.Background(tcell.ColorDarkBlue)
	screen.SetContent(0, 0, ' ', nil, parent)

	source := image.NewNRGBA(image.Rect(0, 0, 1, 2))
	source.Set(0, 0, color.NRGBA{G: 255, A: 255})
	widget := NewImage(source)
	widget.Arrange(NewRect(0, 0, 1, 1))
	widget.Render(screen)

	main, _, style, _ := screen.GetContent(0, 0)
	foreground, background, _ := style.Decompose()
	if main != '▀' || foreground != tcell.NewRGBColor(0, 255, 0) || background != tcell.ColorDarkBlue {
		t.Fatalf("cell=%q foreground=%v background=%v", main, foreground, background)
	}
}

func TestImageContainLeavesLetterboxUntouched(t *testing.T) {
	screen := newImageTestScreen(t, 4, 4)
	defer screen.Fini()
	untouched := tcell.StyleDefault.Background(tcell.ColorPurple)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			screen.SetContent(x, y, '.', nil, untouched)
		}
	}

	source := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			source.Set(x, y, color.NRGBA{R: 255, G: 255, A: 255})
		}
	}
	widget := NewImage(source)
	widget.Arrange(NewRect(0, 0, 4, 4))
	widget.Render(screen)

	if main, _, style, _ := screen.GetContent(0, 0); main != '.' || style != untouched {
		t.Fatalf("top letterbox changed: rune=%q style=%#v", main, style)
	}
	if main, _, _, _ := screen.GetContent(0, 1); main != '▀' {
		t.Fatalf("fitted image did not start on centred row: rune=%q", main)
	}
}

func TestImageBlendsPartialAlphaWithParent(t *testing.T) {
	screen := newImageTestScreen(t, 1, 1)
	defer screen.Fini()
	screen.SetContent(0, 0, ' ', nil, tcell.StyleDefault.Background(tcell.NewRGBColor(0, 0, 255)))

	source := image.NewNRGBA(image.Rect(0, 0, 1, 2))
	source.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 128})
	widget := NewImage(source)
	widget.Arrange(NewRect(0, 0, 1, 1))
	widget.Render(screen)

	_, _, style, _ := screen.GetContent(0, 0)
	foreground, _, _ := style.Decompose()
	if want := tcell.NewRGBColor(128, 0, 127); foreground != want {
		t.Fatalf("blended foreground=%v want=%v", foreground, want)
	}
}

func TestImageFitDimensions(t *testing.T) {
	tests := []struct {
		name       string
		fit        ImageFit
		wantWidth  int
		wantHeight int
	}{
		{name: "contain", fit: ImageFitContain, wantWidth: 20, wantHeight: 10},
		{name: "cover", fit: ImageFitCover, wantWidth: 40, wantHeight: 20},
		{name: "fill", fit: ImageFitFill, wantWidth: 20, wantHeight: 20},
		{name: "native", fit: ImageFitNone, wantWidth: 100, wantHeight: 50},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			width, height := imageFitSize(test.fit, 100, 50, 20, 20)
			if width != test.wantWidth || height != test.wantHeight {
				t.Fatalf("size=%dx%d want=%dx%d", width, height, test.wantWidth, test.wantHeight)
			}
		})
	}
}

func newImageTestScreen(t *testing.T, width, height int) tcell.SimulationScreen {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(width, height)
	return screen
}
