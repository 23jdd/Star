package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	star "github.com/23jdd/Star"
	"github.com/23jdd/Star/widgets"
	"github.com/gdamore/tcell/v2"
)

type tickMsg time.Time

type showcase struct {
	theme widgets.Theme

	clock    *widgets.Text
	nav      *widgets.Tree
	tabs     *widgets.Tabs
	cpu      *widgets.Progress
	memory   *widgets.Progress
	chart    *widgets.Canvas
	logs     *widgets.ScrollView
	services *widgets.Table
	notes    *widgets.TextArea
	settings *widgets.Form
	footer   *widgets.Text

	base   widgets.Widget
	modal  *widgets.Modal
	router *widgets.Router
	ticks  int
	series []float64
}

func newShowcase() *showcase {
	theme := widgets.DefaultTheme()
	a := &showcase{
		theme:  theme,
		clock:  theme.NewText(""),
		cpu:    widgets.NewProgress(0.42),
		memory: widgets.NewProgress(0.61),
		chart:  widgets.NewCanvas(64, 5),
		logs:   widgets.NewScrollView(""),
		notes:  theme.NewTextArea("# Release notes\n\n- Review service health\n- Validate deployment plan"),
		footer: theme.NewText("F1 Help · F2 Commands · Tab Focus · Ctrl+C Quit"),
		series: make([]float64, 64),
	}
	a.cpu.ShowPercentage = true
	a.memory.ShowPercentage = true
	a.logs.Wrap, a.logs.FollowEnd, a.logs.ShowScrollbar = true, true, true

	a.nav = theme.NewTree(&widgets.TreeNode{Label: "Workspace", Expanded: true, Children: []*widgets.TreeNode{
		{Label: "Dashboard"},
		{Label: "Services"},
		{Label: "Notes"},
		{Label: "Settings"},
		{Label: "Image"},
	}})
	a.services = theme.NewTable(
		widgets.TableColumn{Title: "Service"},
		widgets.TableColumn{Title: "Region", Width: 12},
		widgets.TableColumn{Title: "Version", Width: 10},
		widgets.TableColumn{Title: "State", Width: 11},
	)
	a.services.Rows = [][]string{
		{"api-gateway", "us-east", "v2.8.1", "healthy"},
		{"billing", "eu-west", "v1.14.0", "healthy"},
		{"notifications", "us-west", "v3.2.4", "healthy"},
		{"search", "ap-east", "v4.0.2", "degraded"},
		{"worker", "eu-central", "v2.3.7", "healthy"},
	}

	endpoint := theme.NewInput("https://api.example.com")
	token := theme.NewInput("")
	token.Placeholder = "Required access token"
	a.settings = theme.NewForm(
		widgets.FormField{Label: "API endpoint", Input: endpoint, Validate: nonEmpty},
		widgets.FormField{Label: "Access token", Input: token, Validate: nonEmpty},
	)
	a.settings.Submit.Content = " Save settings "

	a.tabs = theme.NewTabs(
		widgets.Tab{Title: "Overview", Content: a.overviewPage()},
		widgets.Tab{Title: "Services", Content: a.servicesPage()},
		widgets.Tab{Title: "Notes", Content: a.notesPage()},
		widgets.Tab{Title: "Settings", Content: a.settingsPage()},
	)

	header := widgets.NewHStack()
	title := theme.NewText(" ★ STAR CONTROL CENTER ")
	title.Style = theme.Primary.Bold(true)
	a.clock.Align = widgets.AlignRight
	header.Add(title, 32, 0)
	header.Add(a.clock, 0, 1)

	navBox := theme.NewBox(a.nav)
	navBox.Border, navBox.Title = &widgets.RoundedBorder, " Navigation "
	body := widgets.NewHStack()
	body.Gap = 1
	body.Add(navBox, 24, 0)
	body.Add(a.tabs, 0, 1)

	root := widgets.NewVStack()
	root.Add(header, 1, 0)
	root.Add(body, 0, 1)
	root.Add(a.footer, 1, 0)
	a.base = root
	a.router = widgets.NewRouter(root)
	a.router.DragThreshold = 1
	a.router.FocusNext()

	a.bindEvents()
	a.appendLog("control center initialized")
	a.updateMetrics(time.Now())
	return a
}

func (a *showcase) overviewPage() widgets.Widget {
	cpuBox := a.theme.NewBox(a.cpu)
	cpuBox.Border, cpuBox.Title = &widgets.RoundedBorder, " CPU "
	memoryBox := a.theme.NewBox(a.memory)
	memoryBox.Border, memoryBox.Title = &widgets.RoundedBorder, " Memory "
	metrics := widgets.NewHStack()
	metrics.Gap = 1
	metrics.Add(cpuBox, 0, 1)
	metrics.Add(memoryBox, 0, 1)

	chartBox := a.theme.NewBox(a.chart)
	chartBox.Border, chartBox.Title = &widgets.RoundedBorder, " Throughput history "
	logBox := a.theme.NewBox(a.logs)
	logBox.Border, logBox.Title = &widgets.RoundedBorder, " Live activity "

	page := widgets.NewVStack()
	page.Gap = 1
	page.Add(metrics, 3, 0)
	page.Add(chartBox, 7, 0)
	page.Add(logBox, 0, 1)
	return page
}

func (a *showcase) servicesPage() widgets.Widget {
	box := a.theme.NewBox(a.services)
	box.Border, box.Title = &widgets.RoundedBorder, " Deployed services "
	return box
}

func (a *showcase) notesPage() widgets.Widget {
	box := a.theme.NewBox(a.notes)
	box.Border, box.Title = &widgets.RoundedBorder, " Editable release notes "
	return box
}

func (a *showcase) settingsPage() widgets.Widget {
	box := a.theme.NewBox(a.settings)
	box.Border, box.Title = &widgets.RoundedBorder, " Connection settings "
	box.Padding = widgets.UniformInsets(1)
	return box
}

func (a *showcase) bindEvents() {
	a.nav.On(widgets.Submitted, false, func(*widgets.Event) {
		node := a.nav.SelectedNode()
		if node == nil {
			return
		}
		pages := map[string]int{"Dashboard": 0, "Services": 1, "Notes": 2, "Settings": 3}
		if page, ok := pages[node.Label]; ok {
			a.tabs.SetActive(page)
			a.footer.Content = "Opened " + node.Label
		}
	})
	a.tabs.On(widgets.Changed, false, func(*widgets.Event) {
		a.footer.Content = "Active page: " + a.tabs.Tabs[a.tabs.Active].Title
	})
	a.services.On(widgets.Submitted, false, func(*widgets.Event) {
		row := a.services.Rows[a.services.Selected]
		a.openTextModal(" Service details ", strings.Join(row, "\n"), 44, 11)
	})
	a.notes.On(widgets.Changed, false, func(*widgets.Event) {
		a.footer.Content = fmt.Sprintf("Notes: %d characters", len([]rune(a.notes.Value)))
	})
	a.settings.On(widgets.Submitted, false, func(*widgets.Event) {
		a.appendLog("connection settings saved")
		a.openTextModal(" Settings saved ", "The validated connection settings were saved.", 52, 9)
	})
}

func (a *showcase) Init() star.Cmd { return nextTick() }

func (a *showcase) Update(message star.Message) (star.Model, star.Cmd) {
	switch message := message.(type) {
	case star.KeyMsg:
		if message.Key == tcell.KeyCtrlC || (message.Key == tcell.KeyEscape && a.modal == nil) {
			return a, star.QuitCmd()
		}
		if a.modal == nil && message.Key == tcell.KeyF1 {
			a.openHelp()
			return a, nil
		}
		if a.modal == nil && message.Key == tcell.KeyF2 {
			a.openCommands()
			return a, nil
		}
		star.RouteInput(a.router, message)
	case star.MouseMsg:
		star.RouteInput(a.router, message)
	case star.WindowSizeMsg:
		a.footer.Content = fmt.Sprintf("%dx%d · F1 Help · F2 Commands · Ctrl+C Quit", message.Width, message.Height)
	case tickMsg:
		a.updateMetrics(time.Time(message))
		return a, nextTick()
	}
	return a, nil
}

func (a *showcase) View() widgets.Widget {
	if a.modal != nil {
		a.router.SetRoot(a.modal)
		return a.modal
	}
	a.router.SetRoot(a.base)
	return a.base
}

func nextTick() star.Cmd {
	return star.Tick(400*time.Millisecond, func(now time.Time) star.Message { return tickMsg(now) })
}

func (a *showcase) updateMetrics(now time.Time) {
	a.ticks++
	a.clock.Content = now.Format("2006-01-02 15:04:05 ")
	a.cpu.Value = 0.48 + 0.34*math.Sin(float64(a.ticks)/5)
	a.memory.Value = 0.62 + 0.12*math.Sin(float64(a.ticks)/11)
	value := 0.5 + 0.45*math.Sin(float64(a.ticks)/3)
	copy(a.series, a.series[1:])
	a.series[len(a.series)-1] = value
	a.drawSeries()
	if a.ticks%12 == 0 {
		a.appendLog(fmt.Sprintf("health check completed · cpu %.0f%%", a.cpu.Value*100))
	}
}

func (a *showcase) drawSeries() {
	levels := []rune("▁▂▃▄▅▆▇█")
	a.chart.Resize(len(a.series), 5)
	for x, value := range a.series {
		level := clampInt(int(value*float64(len(levels))), 0, len(levels)-1)
		for y := 0; y < a.chart.Height; y++ {
			a.chart.SetCell(x, y, widgets.Cell{Main: ' ', Style: a.theme.Surface})
		}
		a.chart.SetCell(x, a.chart.Height-1, widgets.Cell{Main: levels[level], Style: a.theme.Progress})
	}
}

func (a *showcase) appendLog(message string) {
	line := time.Now().Format("15:04:05") + "  " + message
	if a.logs.Content != "" {
		a.logs.Content += "\n"
	}
	a.logs.SetContent(a.logs.Content + line)
}

func (a *showcase) openHelp() {
	text := "Keyboard\n\nF1  help\nF2  command palette\nTab / Shift-Tab  focus\nArrow keys  navigate widgets\nEnter  activate\nEsc  close modal or quit\n\nMouse clicking, wheel scrolling, forms, tables, trees, tabs and editing are enabled."
	a.openTextModal(" Help ", text, 62, 17)
}

func (a *showcase) openCommands() {
	commands := a.theme.NewList("Open overview", "Open services", "Clear activity log", "Reset metrics", "Close")
	modal := a.theme.NewModal(commands)
	modal.Title, modal.Width, modal.Height = " Command palette ", 42, 11
	a.modal = modal
	commands.On(widgets.Submitted, false, func(*widgets.Event) {
		switch commands.Selected {
		case 0:
			a.tabs.SetActive(0)
		case 1:
			a.tabs.SetActive(1)
		case 2:
			a.logs.SetContent("")
		case 3:
			a.ticks = 0
			a.series = make([]float64, len(a.series))
			a.updateMetrics(time.Now())
		}
		a.closeModal()
	})
	a.bindModalDismiss(modal)
	a.router.SetRoot(modal)
	a.router.FocusNext()
}

func (a *showcase) openTextModal(title, text string, width, height int) {
	content := a.theme.NewScrollView(text)
	content.Wrap = true
	modal := a.theme.NewModal(content)
	modal.Title, modal.Width, modal.Height = title, width, height
	a.modal = modal
	a.bindModalDismiss(modal)
	a.router.SetRoot(modal)
	a.router.FocusNext()
}

func (a *showcase) bindModalDismiss(modal *widgets.Modal) {
	modal.On(widgets.Dismissed, false, func(*widgets.Event) { a.closeModal() })
}

func (a *showcase) closeModal() {
	a.modal = nil
	a.router.SetRoot(a.base)
	a.router.FocusNext()
}

func nonEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("required")
	}
	return nil
}

func clampInt(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func main() {
	program := star.NewProgram(
		star.WithMouse(true),
		star.WithPaste(true),
		star.WithFocus(true),
	)
	if err := program.Run(newShowcase()); err != nil {
		panic(err)
	}
}
