package main

import (
	"testing"

	star "github.com/23jdd/Star"
	"github.com/23jdd/Star/widgets"
)

func TestImageExampleChangesFit(t *testing.T) {
	app := newImageApp()
	app.Update(star.KeyMsg{Rune: '2'})
	if app.image.Fit != widgets.ImageFitCover {
		t.Fatalf("fit=%v want cover", app.image.Fit)
	}
	app.Update(star.KeyMsg{Rune: '4'})
	if app.image.Fit != widgets.ImageFitNone {
		t.Fatalf("fit=%v want native", app.image.Fit)
	}
}
