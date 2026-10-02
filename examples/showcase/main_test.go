package main

import (
	"testing"
	"time"

	star "github.com/23jdd/Star"
	"github.com/gdamore/tcell/v2"
)

func TestShowcaseTickAndModalLifecycle(t *testing.T) {
	app := newShowcase()
	previousTicks := app.ticks
	model, command := app.Update(tickMsg(time.Now()))
	if model != app || command == nil || app.ticks != previousTicks+1 {
		t.Fatal("tick did not update metrics or schedule the next tick")
	}

	app.Update(star.KeyMsg{Key: tcell.KeyF1})
	if app.modal == nil || app.View() != app.modal {
		t.Fatal("F1 did not open the help modal")
	}
	app.Update(star.KeyMsg{Key: tcell.KeyEscape})
	if app.modal != nil || app.View() != app.base {
		t.Fatal("Escape did not dismiss the modal")
	}
}

func TestShowcaseCommandPalette(t *testing.T) {
	app := newShowcase()
	app.Update(star.KeyMsg{Key: tcell.KeyF2})
	if app.modal == nil {
		t.Fatal("F2 did not open command palette")
	}
	app.Update(star.KeyMsg{Key: tcell.KeyEscape})
	if app.modal != nil {
		t.Fatal("command palette did not close")
	}
}
