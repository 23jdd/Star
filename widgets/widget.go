package widgets

import "github.com/gdamore/tcell/v2"

// Widget is the minimal contract implemented by all visual components.
type Widget interface {
	Render(screen tcell.Screen)
	Measure(constraints Constraints) Size
	Arrange(bounds Rect)
	HandlerEvent(e *Event)
}

// WidgetBase stores the common layout and event state. Custom widgets may
// embed it and expose it through EventBase to participate in event routing.
type WidgetBase struct {
	bounds    Rect
	parent    Widget
	hidden    bool
	listeners map[EventKind][]listener

	// DefaultAction runs after propagation unless PreventDefault was called.
	DefaultAction Handler
}

func (b *WidgetBase) Bounds() Rect     { return b.bounds }
func (b *WidgetBase) SetBounds(r Rect) { b.bounds = r }
func (b *WidgetBase) Parent() Widget   { return b.parent }

// SetParent connects a custom child widget to its event parent. Built-in
// containers call this automatically; custom containers should call it from
// construction or Arrange.
func (b *WidgetBase) SetParent(parent Widget) { b.parent = parent }
func (b *WidgetBase) SetHidden(v bool)        { b.hidden = v }
func (b *WidgetBase) Hidden() bool            { return b.hidden }
func (b *WidgetBase) EventBase() *WidgetBase  { return b }

func (b *WidgetBase) setParent(parent Widget) { b.SetParent(parent) }

// On appends an event listener. Capture listeners run on the way down the
// tree; non-capture listeners run at the target and while bubbling.
func (b *WidgetBase) On(kind EventKind, capture bool, handler Handler) {
	if handler == nil {
		return
	}
	if b.listeners == nil {
		b.listeners = make(map[EventKind][]listener)
	}
	b.listeners[kind] = append(b.listeners[kind], listener{capture: capture, handler: handler})
}

func (b *WidgetBase) invoke(owner Widget, e *Event, capture bool) {
	e.CurrentTarget = owner
	for _, item := range append([]listener(nil), b.listeners[e.Kind]...) {
		if item.capture == capture {
			item.handler(e)
			if e.immediateStopped {
				return
			}
		}
	}
}

type baseProvider interface{ EventBase() *WidgetBase }
type childProvider interface{ EventChildren() []Widget }

func widgetBase(widget Widget) *WidgetBase {
	if provider, ok := widget.(baseProvider); ok {
		return provider.EventBase()
	}
	return nil
}
