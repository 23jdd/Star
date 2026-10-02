package main

import (
	"fmt"
	"strings"
	"time"

	star "github.com/23jdd/Star"
	"github.com/23jdd/Star/widgets"
	"github.com/gdamore/tcell/v2"
)

type pulseMsg time.Time

type dashboard struct {
	theme    widgets.Theme
	input    *widgets.Input
	services *widgets.List
	progress *widgets.Progress
	logs     *widgets.ScrollView
	status   *widgets.Text
	deploy   *widgets.Button
	root     widgets.Widget
	router   *widgets.Router
	build    int
}

func newDashboard() *dashboard {
	theme := widgets.DefaultTheme()
	d := &dashboard{
		theme:    theme,
		input:    theme.NewInput(""),
		services: theme.NewList("api-gateway", "billing", "notifications", "search"),
		progress: widgets.NewProgress(0),
		logs:     widgets.NewScrollView("[ready] dashboard initialized"),
		status:   theme.NewText("Starting…"),
		deploy:   theme.NewButton(" Deploy "),
	}
	d.input.Placeholder = "New service name"
	d.progress.ShowPercentage = true
	d.logs.FollowEnd = true
	d.logs.ShowScrollbar = true
	d.logs.Wrap = true

	servicesBox := theme.NewBox(d.services)
	servicesBox.Border = &widgets.RoundedBorder
	servicesBox.Title = " Services "

	inputBox := theme.NewBox(d.input)
	inputBox.Border = &widgets.RoundedBorder
	inputBox.Title = " Add service "

	logsBox := theme.NewBox(d.logs)
	logsBox.Border = &widgets.RoundedBorder
	logsBox.Title = " Activity "

	left := widgets.NewVStack()
	left.Gap = 1
	left.Add(inputBox, 3, 0)
	left.Add(servicesBox, 0, 1)

	right := widgets.NewVStack()
	right.Gap = 1
	right.Add(d.progress, 1, 0)
	right.Add(logsBox, 0, 1)
	right.Add(d.deploy, 1, 0)

	body := widgets.NewHStack()
	body.Gap = 1
	body.Add(left, 30, 0)
	body.Add(right, 0, 1)

	header := theme.NewText(" Star Operations Dashboard ")
	header.Style = theme.Primary.Bold(true)
	d.status.Style = theme.Disabled

	root := widgets.NewVStack()
	root.Add(header, 1, 0)
	root.Add(body, 0, 1)
	root.Add(d.status, 1, 0)
	d.root = root
	d.router = widgets.NewRouter(root)
	d.router.FocusNext()

	d.input.On(widgets.Submitted, false, func(*widgets.Event) {
		name := strings.TrimSpace(d.input.Value)
		if name == "" {
			d.status.Content = "Service name cannot be empty"
			return
		}
		d.services.Items = append(d.services.Items, name)
		d.services.SetSelected(len(d.services.Items) - 1)
		d.input.SetValue("")
		d.appendLog(fmt.Sprintf("[service] added %s", name))
	})
	d.services.On(widgets.Changed, false, func(*widgets.Event) {
		d.status.Content = "Selected " + d.selectedService()
	})
	d.deploy.On(widgets.Click, false, func(*widgets.Event) {
		d.build++
		d.progress.Value = 0
		d.appendLog(fmt.Sprintf("[deploy #%d] started %s", d.build, d.selectedService()))
	})
	return d
}

func (d *dashboard) Init() star.Cmd {
	return star.Tick(120*time.Millisecond, func(now time.Time) star.Message {
		return pulseMsg(now)
	})
}

func (d *dashboard) Update(message star.Message) (star.Model, star.Cmd) {
	switch message := message.(type) {
	case star.KeyMsg:
		if message.Key == tcell.KeyEscape || message.Key == tcell.KeyCtrlC {
			return d, star.QuitCmd()
		}
		star.RouteInput(d.router, message)
	case star.MouseMsg:
		star.RouteInput(d.router, message)
	case star.WindowSizeMsg:
		d.status.Content = fmt.Sprintf("%dx%d · Tab/Shift-Tab: focus · Enter: activate · Esc: quit", message.Width, message.Height)
	case pulseMsg:
		d.progress.Value += 0.025
		if d.progress.Value >= 1 {
			d.progress.Value = 0
			d.appendLog("[health] all services responding")
		}
		return d, star.Tick(120*time.Millisecond, func(now time.Time) star.Message {
			return pulseMsg(now)
		})
	}
	return d, nil
}

func (d *dashboard) View() widgets.Widget {
	d.router.SetRoot(d.root)
	return d.root
}

func (d *dashboard) selectedService() string {
	if len(d.services.Items) == 0 {
		return "no service"
	}
	index := d.services.Selected
	if index < 0 || index >= len(d.services.Items) {
		index = 0
	}
	return d.services.Items[index]
}

func (d *dashboard) appendLog(line string) {
	stamp := time.Now().Format("15:04:05")
	if d.logs.Content != "" {
		d.logs.Content += "\n"
	}
	d.logs.SetContent(d.logs.Content + stamp + " " + line)
	d.status.Content = line
}

func main() {
	program := star.NewProgram(
		star.WithMouse(true),
		star.WithPaste(true),
		star.WithFocus(true),
	)
	if err := program.Run(newDashboard()); err != nil {
		panic(err)
	}
}
