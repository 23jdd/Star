package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// Validator returns nil when a field value is valid.
type Validator func(value string) error

// FormField binds a label and validation function to an Input.
type FormField struct {
	Label    string
	Input    *Input
	Validate Validator
}

// Form lays out labeled inputs, displays validation errors, and emits
// Submitted after a valid submit-button activation.
type Form struct {
	WidgetBase
	Fields     []FormField
	Submit     *Button
	Errors     map[int]string
	Style      tcell.Style
	LabelStyle tcell.Style
	ErrorStyle tcell.Style
	FieldGap   int
	Disabled   bool
}

// NewForm creates a form and supplies missing Input fields.
func NewForm(fields ...FormField) *Form {
	f := &Form{
		Fields:     append([]FormField(nil), fields...),
		Submit:     NewButton(" Submit "),
		Errors:     make(map[int]string),
		Style:      tcell.StyleDefault,
		LabelStyle: tcell.StyleDefault.Bold(true),
		ErrorStyle: tcell.StyleDefault.Foreground(tcell.ColorRed),
	}
	for index := range f.Fields {
		if f.Fields[index].Input == nil {
			f.Fields[index].Input = NewInput("")
		}
	}
	f.attach()
	return f
}

func (f *Form) EventChildren() []Widget {
	if f.Disabled {
		return nil
	}
	children := make([]Widget, 0, len(f.Fields)+1)
	for _, field := range f.Fields {
		if field.Input != nil {
			children = append(children, field.Input)
		}
	}
	if f.Submit != nil {
		children = append(children, f.Submit)
	}
	return children
}

func (f *Form) Measure(c Constraints) Size {
	width, height := 0, 0
	for _, field := range f.Fields {
		width = max(width, uniseg.StringWidth(field.Label))
		if field.Input != nil {
			width = max(width, field.Input.Measure(Loose(c.MaxW, 1)).W)
		}
		height += 2 + max(0, f.FieldGap)
	}
	if len(f.Fields) > 0 {
		height -= max(0, f.FieldGap)
	}
	if f.Submit != nil {
		size := f.Submit.Measure(Loose(c.MaxW, 1))
		width, height = max(width, size.W), height+1
		if len(f.Fields) > 0 {
			height += max(0, f.FieldGap)
		}
	}
	return c.Constrain(Size{W: width, H: height})
}

func (f *Form) Arrange(bounds Rect) {
	f.SetBounds(bounds)
	f.attach()
	y := bounds.Y
	bottom := bounds.Y + bounds.H
	for _, field := range f.Fields {
		if field.Input != nil {
			height := 0
			if y+1 < bottom {
				height = 1
			}
			field.Input.Arrange(NewRect(bounds.X, y+1, bounds.W, height))
		}
		y += 2 + max(0, f.FieldGap)
	}
	if f.Submit != nil {
		height := 0
		if y < bottom {
			height = 1
		}
		f.Submit.Arrange(NewRect(bounds.X, y, bounds.W, height))
	}
}

func (f *Form) Render(screen tcell.Screen) {
	if f.Hidden() || f.Bounds().Empty() {
		return
	}
	fill(screen, f.Bounds(), ' ', f.Style)
	y := f.Bounds().Y
	bottom := f.Bounds().Y + f.Bounds().H
	for index, field := range f.Fields {
		if y >= bottom {
			break
		}
		drawText(screen, f.Bounds().X, y, f.Bounds().W, field.Label, f.LabelStyle)
		if message := f.Errors[index]; message != "" {
			prefix := field.Label + ": "
			x := f.Bounds().X + min(f.Bounds().W, uniseg.StringWidth(prefix))
			drawText(screen, x, y, max(0, f.Bounds().W-(x-f.Bounds().X)), message, f.ErrorStyle)
		}
		if field.Input != nil {
			field.Input.Render(screen)
		}
		y += 2 + max(0, f.FieldGap)
	}
	if f.Submit != nil {
		f.Submit.Render(screen)
	}
}

func (f *Form) HandlerEvent(event *Event) {
	if f.Disabled || event.Phase != Bubble || event.Kind != Click || f.Submit == nil || !sameWidget(event.Target, f.Submit) {
		return
	}
	if f.Validate() {
		Dispatch(f, &Event{Kind: Submitted})
	}
}

// Validate replaces Errors and reports whether every field is valid.
func (f *Form) Validate() bool {
	f.Errors = make(map[int]string)
	for index, field := range f.Fields {
		if field.Validate == nil || field.Input == nil {
			continue
		}
		if err := field.Validate(field.Input.Value); err != nil {
			f.Errors[index] = err.Error()
		}
	}
	return len(f.Errors) == 0
}

// Values returns field values in display order.
func (f *Form) Values() []string {
	values := make([]string, len(f.Fields))
	for index, field := range f.Fields {
		if field.Input != nil {
			values[index] = field.Input.Value
		}
	}
	return values
}

func (f *Form) attach() {
	for _, child := range f.EventChildren() {
		if base := widgetBase(child); base != nil {
			base.setParent(f)
		}
	}
}
