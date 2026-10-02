package main

import (
	"errors"
	"fmt"
	"strings"

	star "github.com/23jdd/Star"
	"github.com/23jdd/Star/widgets"
	"github.com/gdamore/tcell/v2"
)

type componentApp struct {
	theme  widgets.Theme
	table  *widgets.Table
	tree   *widgets.Tree
	editor *widgets.TextArea
	form   *widgets.Form
	tabs   *widgets.Tabs
	status *widgets.Text
	base   widgets.Widget
	modal  *widgets.Modal
	router *widgets.Router
}

func newComponentApp() *componentApp {
	theme := widgets.DefaultTheme()
	a := &componentApp{theme: theme, status: theme.NewText("←/→ switches tabs · Tab changes focus · F1 opens help · Esc quits")}

	a.table = theme.NewTable(
		widgets.TableColumn{Title: "Service"},
		widgets.TableColumn{Title: "Region", Width: 12},
		widgets.TableColumn{Title: "State", Width: 10},
	)
	a.table.Rows = [][]string{
		{"api", "us-east", "healthy"},
		{"billing", "eu-west", "healthy"},
		{"search", "ap-east", "degraded"},
		{"worker", "us-west", "healthy"},
	}

	a.tree = theme.NewTree(&widgets.TreeNode{Label: "production", Expanded: true, Children: []*widgets.TreeNode{
		{Label: "frontend", Children: []*widgets.TreeNode{{Label: "web"}, {Label: "assets"}}},
		{Label: "backend", Expanded: true, Children: []*widgets.TreeNode{{Label: "api"}, {Label: "jobs"}}},
	}})
	a.editor = theme.NewTextArea("Star advanced widgets\n\nEdit this text with normal cursor keys.")
	a.editor.Placeholder = "Write release notes"

	name := theme.NewInput("")
	name.Placeholder = "Project name"
	description := theme.NewInput("")
	description.Placeholder = "Short description"
	a.form = theme.NewForm(
		widgets.FormField{Label: "Name", Input: name, Validate: required},
		widgets.FormField{Label: "Description", Input: description, Validate: required},
	)

	a.tabs = theme.NewTabs(
		widgets.Tab{Title: "Table", Content: a.table},
		widgets.Tab{Title: "Tree", Content: a.tree},
		widgets.Tab{Title: "Editor", Content: a.editor},
		widgets.Tab{Title: "Form", Content: a.form},
	)
	root := widgets.NewVStack()
	root.Add(a.tabs, 0, 1)
	root.Add(a.status, 1, 0)
	a.base = root
	a.router = widgets.NewRouter(root)
	a.router.FocusNext()

	a.tabs.On(widgets.Changed, false, func(*widgets.Event) {
		a.status.Content = fmt.Sprintf("Active tab: %s", a.tabs.Tabs[a.tabs.Active].Title)
	})
	a.table.On(widgets.Submitted, false, func(*widgets.Event) {
		a.status.Content = "Table row: " + strings.Join(a.table.Rows[a.table.Selected], " / ")
	})
	a.tree.On(widgets.Submitted, false, func(*widgets.Event) {
		if node := a.tree.SelectedNode(); node != nil {
			a.status.Content = "Tree node: " + node.Label
		}
	})
	a.form.On(widgets.Submitted, false, func(*widgets.Event) {
		a.status.Content = "Form submitted: " + strings.Join(a.form.Values(), " / ")
	})
	return a
}

func required(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("required")
	}
	return nil
}

func (a *componentApp) Init() star.Cmd { return nil }

func (a *componentApp) Update(message star.Message) (star.Model, star.Cmd) {
	if key, ok := message.(star.KeyMsg); ok {
		if key.Key == tcell.KeyCtrlC || (key.Key == tcell.KeyEscape && a.modal == nil) {
			return a, star.QuitCmd()
		}
		if key.Key == tcell.KeyF1 && a.modal == nil {
			a.openHelp()
			return a, nil
		}
	}
	star.RouteInput(a.router, message)
	return a, nil
}

func (a *componentApp) View() widgets.Widget {
	if a.modal != nil {
		a.router.SetRoot(a.modal)
		return a.modal
	}
	a.router.SetRoot(a.base)
	return a.base
}

func (a *componentApp) openHelp() {
	content := a.theme.NewText("Advanced components\n\nTable: ↑/↓ selects rows\nTree: ←/→ collapses and expands\nEditor: multiline text editing\nForm: Tab moves through fields\n\nPress Esc or click outside to close")
	content.Wrap = true
	a.modal = a.theme.NewModal(content)
	a.modal.Title = " Help "
	a.modal.Width, a.modal.Height = 54, 12
	a.modal.On(widgets.Dismissed, false, func(*widgets.Event) {
		a.modal = nil
		a.router.SetRoot(a.base)
		a.router.FocusNext()
	})
	a.router.SetRoot(a.modal)
}

func main() {
	program := star.NewProgram(star.WithMouse(true), star.WithFocus(true))
	if err := program.Run(newComponentApp()); err != nil {
		panic(err)
	}
}
