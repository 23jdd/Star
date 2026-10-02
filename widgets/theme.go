package widgets

import (
	"image"

	"github.com/gdamore/tcell/v2"
)

// Theme groups semantic styles for applications that want consistent widget
// construction. Widgets still expose their styles for local overrides.
type Theme struct {
	Text      tcell.Style
	Surface   tcell.Style
	Border    tcell.Style
	Primary   tcell.Style
	Focused   tcell.Style
	Selected  tcell.Style
	Disabled  tcell.Style
	Scrollbar tcell.Style
	Progress  tcell.Style
	Error     tcell.Style
	Backdrop  tcell.Style
}

// DefaultTheme returns a readable dark terminal theme.
func DefaultTheme() Theme {
	return Theme{
		Text:      tcell.StyleDefault.Foreground(tcell.ColorWhite),
		Surface:   tcell.StyleDefault.Background(tcell.ColorBlack),
		Border:    tcell.StyleDefault.Foreground(tcell.ColorGray),
		Primary:   tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue),
		Focused:   tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorLightBlue).Bold(true),
		Selected:  tcell.StyleDefault.Reverse(true),
		Disabled:  tcell.StyleDefault.Foreground(tcell.ColorGray),
		Scrollbar: tcell.StyleDefault.Foreground(tcell.ColorGray),
		Progress:  tcell.StyleDefault.Foreground(tcell.ColorGreen),
		Error:     tcell.StyleDefault.Foreground(tcell.ColorRed),
		Backdrop:  tcell.StyleDefault.Background(tcell.ColorBlack),
	}
}

func (t Theme) NewText(content string) *Text {
	widget := NewText(content)
	widget.Style = t.Text
	return widget
}

// NewImage creates an image that uses the theme surface when alpha blending
// cannot obtain a concrete background colour from the terminal cell.
func (t Theme) NewImage(source image.Image) *Image {
	widget := NewImage(source)
	_, widget.FallbackBackground, _ = t.Surface.Decompose()
	return widget
}

// NewImageFromPath decodes an image file and applies the theme surface as its
// fallback alpha-compositing background.
func (t Theme) NewImageFromPath(path string) (*Image, error) {
	widget, err := NewImageFromPath(path)
	if err != nil {
		return nil, err
	}
	_, widget.FallbackBackground, _ = t.Surface.Decompose()
	return widget, nil
}

func (t Theme) NewButton(content string) *Button {
	widget := NewButton(content)
	widget.Style, widget.FocusedStyle, widget.DisabledStyle = t.Primary, t.Focused, t.Disabled
	return widget
}

func (t Theme) NewInput(value string) *Input {
	widget := NewInput(value)
	widget.Style, widget.FocusStyle = t.Surface, t.Focused
	return widget
}

func (t Theme) NewList(items ...string) *List {
	widget := NewList(items...)
	widget.Style, widget.SelectedStyle = t.Surface, t.Selected
	return widget
}

// NewScrollView creates a scrollable text viewport using theme styles.
func (t Theme) NewScrollView(content string) *ScrollView {
	widget := NewScrollView(content)
	widget.Style, widget.ScrollbarStyle = t.Surface, t.Scrollbar
	return widget
}

func (t Theme) NewBox(child Widget) *Box {
	widget := NewBox(child)
	widget.Style, widget.BorderStyle = t.Surface, t.Border
	return widget
}

// NewTextArea creates a multiline input using theme styles.
func (t Theme) NewTextArea(value string) *TextArea {
	widget := NewTextArea(value)
	widget.Style, widget.FocusStyle = t.Surface, t.Focused
	return widget
}

// NewTable creates a table using theme styles.
func (t Theme) NewTable(columns ...TableColumn) *Table {
	widget := NewTable(columns...)
	widget.Style, widget.HeaderStyle, widget.SelectedStyle = t.Surface, t.Primary, t.Selected
	return widget
}

// NewTree creates a tree using theme styles.
func (t Theme) NewTree(roots ...*TreeNode) *Tree {
	widget := NewTree(roots...)
	widget.Style, widget.SelectedStyle = t.Surface, t.Selected
	return widget
}

// NewTabs creates tabs using theme styles.
func (t Theme) NewTabs(tabs ...Tab) *Tabs {
	widget := NewTabs(tabs...)
	widget.Style, widget.ActiveStyle, widget.DisabledStyle = t.Surface, t.Primary, t.Disabled
	return widget
}

// NewModal creates a modal using theme styles.
func (t Theme) NewModal(child Widget) *Modal {
	widget := NewModal(child)
	widget.Style, widget.BorderStyle, widget.BackdropStyle = t.Surface, t.Border, t.Backdrop
	return widget
}

// NewForm creates a validated form using theme styles.
func (t Theme) NewForm(fields ...FormField) *Form {
	widget := NewForm(fields...)
	widget.Style, widget.LabelStyle, widget.ErrorStyle = t.Surface, t.Text.Bold(true), t.Error
	widget.Submit.Style, widget.Submit.FocusedStyle, widget.Submit.DisabledStyle = t.Primary, t.Focused, t.Disabled
	for _, field := range widget.Fields {
		if field.Input != nil {
			field.Input.Style, field.Input.FocusStyle = t.Surface, t.Focused
		}
	}
	return widget
}
