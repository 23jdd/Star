package main

import (
	"fmt"
	"time"

	Star "github.com/23jdd/Star"
	"github.com/23jdd/Star/widgets"
	"github.com/gdamore/tcell/v2"
)

type Sleep struct {
	d time.Duration
}

type App struct {
	counter int
}

func (a App) Init() Star.Cmd {
	return func() Star.Message {
		time.Sleep(time.Second)
		return Sleep{d: time.Second}
	}
}

func (a App) Update(message Star.Message) (Star.Model, Star.Cmd) {
	if key, ok := message.(Star.KeyMsg); ok && (key.Key == tcell.KeyEscape || key.Rune == 'q') {
		return a, Star.QuitCmd()
	}
	if s, ok := message.(Sleep); ok {
		a.counter += 1
		return a, func() Star.Message {
			time.Sleep(s.d)
			return Sleep{d: s.d}
		}
	}
	return a, nil
}

func (a App) View() widgets.Widget {
	title := widgets.NewText("Star — Elm-style TUI")
	counter := widgets.NewButton(fmt.Sprintf(" ticks: %d ", a.counter))
	help := widgets.NewText("Press q or Esc to quit")
	return widgets.NewVStack(title, counter, help)
}
func main() {
	if err := Star.NewProgram().Run(App{counter: 0}); err != nil {
		panic(err)
	}
}
