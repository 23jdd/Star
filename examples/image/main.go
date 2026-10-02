package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	star "github.com/23jdd/Star"
	"github.com/23jdd/Star/widgets"
	"github.com/gdamore/tcell/v2"
)

type imageApp struct {
	image  *widgets.Image
	status *widgets.Text
	root   widgets.Widget
}

func newImageApp() *imageApp {
	theme := widgets.DefaultTheme()
	return buildImageApp(theme, theme.NewImage(makeDemoImage(96, 48)))
}

func newImageAppFromPath(path string) (*imageApp, error) {
	theme := widgets.DefaultTheme()
	preview, err := theme.NewImageFromPath(path)
	if err != nil {
		return nil, err
	}
	return buildImageApp(theme, preview), nil
}

func buildImageApp(theme widgets.Theme, preview *widgets.Image) *imageApp {
	status := theme.NewText("")
	status.Align = widgets.AlignCenter

	frame := theme.NewBox(preview)
	frame.Border = &widgets.RoundedBorder
	frame.Title = " Half-block true-colour image "
	frame.Padding = widgets.UniformInsets(1)
	frame.Style = tcell.StyleDefault.Background(tcell.NewRGBColor(8, 18, 36))
	frame.BorderStyle = tcell.StyleDefault.Foreground(tcell.ColorLightSkyBlue)

	title := theme.NewText("STAR IMAGE WIDGET")
	title.Align = widgets.AlignCenter
	title.Style = theme.Primary.Bold(true)
	help := theme.NewText("1 Contain  ·  2 Cover  ·  3 Fill  ·  4 Native  ·  Q/Esc Quit")
	help.Align = widgets.AlignCenter

	root := widgets.NewVStack()
	root.Add(title, 1, 0)
	root.Add(frame, 0, 1)
	root.Add(status, 1, 0)
	root.Add(help, 1, 0)

	app := &imageApp{image: preview, status: status, root: root}
	app.setFit(widgets.ImageFitContain)
	return app
}

func (a *imageApp) Init() star.Cmd { return nil }

func (a *imageApp) Update(message star.Message) (star.Model, star.Cmd) {
	key, ok := message.(star.KeyMsg)
	if !ok {
		return a, nil
	}
	switch {
	case key.Key == tcell.KeyEscape || key.Key == tcell.KeyCtrlC || key.Rune == 'q' || key.Rune == 'Q':
		return a, star.QuitCmd()
	case key.Rune == '1':
		a.setFit(widgets.ImageFitContain)
	case key.Rune == '2':
		a.setFit(widgets.ImageFitCover)
	case key.Rune == '3':
		a.setFit(widgets.ImageFitFill)
	case key.Rune == '4':
		a.setFit(widgets.ImageFitNone)
	}
	return a, nil
}

func (a *imageApp) View() widgets.Widget { return a.root }

func (a *imageApp) setFit(fit widgets.ImageFit) {
	a.image.Fit = fit
	names := map[widgets.ImageFit]string{
		widgets.ImageFitContain: "Contain — 完整显示并保持比例",
		widgets.ImageFitCover:   "Cover — 填满区域并裁剪",
		widgets.ImageFitFill:    "Fill — 拉伸到可用区域",
		widgets.ImageFitNone:    "Native — 每个源像素对应半个单元格",
	}
	a.status.Content = fmt.Sprintf("当前模式：%s", names[fit])
}

func makeDemoImage(width, height int) image.Image {
	result := image.NewNRGBA(image.Rect(0, 0, width, height))
	cx, cy := float64(width-1)/2, float64(height-1)/2
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx, dy := (float64(x)-cx)/cx, (float64(y)-cy)/cy
			distance := math.Sqrt(dx*dx + dy*dy)
			if distance > 1 {
				continue // Transparent corners expose the Box background.
			}
			angle := math.Atan2(dy, dx)
			wave := (math.Sin(angle*6-distance*14) + 1) / 2
			result.SetNRGBA(x, y, color.NRGBA{
				R: uint8(35 + 210*wave),
				G: uint8(70 + 170*(1-distance)),
				B: uint8(120 + 120*(1-wave)),
				A: uint8(255 * math.Min(1, (1-distance)*8)),
			})
		}
	}
	return result
}

func main() {
	app := newImageApp()
	app, err := newImageAppFromPath("img.png")
	if err != nil {
		panic(err)
	}
	if err := star.NewProgram().Run(app); err != nil {
		panic(err)
	}
}
