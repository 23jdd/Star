package main

import (
	"fmt"

	star "github.com/23jdd/Star"
	"github.com/23jdd/Star/widgets"
	"github.com/gdamore/tcell/v2"
)

type app struct {
	input  *widgets.Input
	list   *widgets.List
	status *widgets.Text
	router *widgets.Router
	root   widgets.Widget
}

func newApp() *app {
	a := &app{
		input:  widgets.NewInput(""),
		list:   widgets.NewList("Development", "Design", "Operations"),
		status: widgets.NewText("Tab changes focus; Esc quits"),
	}
	a.input.Placeholder = "Project name"
	save := widgets.NewButton(" Save ")
	inputBox := widgets.NewBox(a.input)
	inputBox.Border, inputBox.Title = &widgets.RoundedBorder, " Name "
	listBox := widgets.NewBox(a.list)
	listBox.Border, listBox.Title = &widgets.RoundedBorder, " Team "

	root := widgets.NewVStack()
	root.Gap = 1
	root.Add(widgets.NewText("Create project"), 1, 0)
	root.Add(inputBox, 3, 0)
	root.Add(listBox, 0, 1)
	root.Add(save, 1, 0)
	root.Add(a.status, 1, 0)
	a.root = root
	a.router = widgets.NewRouter(root)

	save.On(widgets.Click, false, func(*widgets.Event) {
		team := ""
		if len(a.list.Items) > 0 {
			team = a.list.Items[a.list.Selected]
		}
		a.status.Content = fmt.Sprintf("Saved %q for %s", a.input.Value, team)
	})
	return a
}

func (a *app) Init() star.Cmd { return nil }

func (a *app) Update(message star.Message) (star.Model, star.Cmd) {
	if key, ok := message.(star.KeyMsg); ok && (key.Key == tcell.KeyEscape || key.Key == tcell.KeyCtrlC) {
		return a, star.QuitCmd()
	}
	star.RouteInput(a.router, message)
	return a, nil
}

func (a *app) View() widgets.Widget {
	a.router.SetRoot(a.root)
	return a.root
}

func main() {
	if err := star.NewProgram(star.WithMouse(true)).Run(newApp()); err != nil {
		panic(err)
	}
}
