// Package star provides a small Elm-style terminal UI runtime backed by tcell.
//
// A Model owns application state. Star serializes terminal events and command
// results through Model.Update, then measures, arranges, and renders the Widget
// returned by Model.View. Commands may perform blocking work because they run
// outside the update loop.
package star
