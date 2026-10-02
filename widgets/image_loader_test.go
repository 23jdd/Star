package widgets

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestNewImageFromPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pixel.png")
	source := image.NewNRGBA(image.Rect(0, 0, 3, 5))
	source.SetNRGBA(1, 2, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	writePNG(t, path, source)

	widget, err := NewImageFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if widget.Source == nil || widget.Source.Bounds() != source.Bounds() {
		t.Fatalf("bounds=%v want=%v", widget.Source.Bounds(), source.Bounds())
	}
	if got := color.NRGBAModel.Convert(widget.Source.At(1, 2)).(color.NRGBA); got != (color.NRGBA{R: 10, G: 20, B: 30, A: 255}) {
		t.Fatalf("pixel=%#v", got)
	}
}

func TestSetImagePathKeepsSourceOnError(t *testing.T) {
	original := image.NewRGBA(image.Rect(0, 0, 2, 2))
	widget := NewImage(original)
	if err := widget.SetImagePath(filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("missing image returned no error")
	}
	if widget.Source != original {
		t.Fatal("source changed after a failed load")
	}
}

func writePNG(t *testing.T, path string, source image.Image) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, source); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
