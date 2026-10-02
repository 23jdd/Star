package widgets

import "github.com/gdamore/tcell/v2"

// Overlay arranges all children into the same bounds and renders later
// children above earlier ones.
type Overlay struct {
	WidgetBase
	Children []Widget
}

// NewOverlay creates a layered copy of children.
func NewOverlay(children ...Widget) *Overlay {
	o := &Overlay{Children: append([]Widget(nil), children...)}
	o.attach()
	return o
}

func (o *Overlay) attach() {
	for _, child := range o.Children {
		if base := widgetBase(child); base != nil {
			base.setParent(o)
		}
	}
}

func (o *Overlay) EventChildren() []Widget { return append([]Widget(nil), o.Children...) }

func (o *Overlay) Measure(constraints Constraints) Size {
	size := Size{}
	for _, child := range o.Children {
		if child != nil {
			measured := child.Measure(constraints)
			size.W, size.H = max(size.W, measured.W), max(size.H, measured.H)
		}
	}
	return constraints.Constrain(size)
}

func (o *Overlay) Arrange(bounds Rect) {
	o.SetBounds(bounds)
	o.attach()
	for _, child := range o.Children {
		if child != nil {
			child.Arrange(bounds)
		}
	}
}

func (o *Overlay) Render(screen tcell.Screen) {
	if o.Hidden() {
		return
	}
	for _, child := range o.Children {
		if child != nil {
			child.Render(screen)
		}
	}
}

func (o *Overlay) HandlerEvent(*Event) {}
