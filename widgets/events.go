package widgets

import "github.com/gdamore/tcell/v2"

// EventKind identifies a routed widget event.
type EventKind uint8

const (
	// KeyDown is a keyboard press.
	KeyDown EventKind = iota
	// PointerDown is a primary-pointer press.
	PointerDown
	// PointerMove is pointer movement or a changed button state.
	PointerMove
	// PointerUp is a primary-pointer release.
	PointerUp
	// Click is a synthesized activation.
	Click
	// PointerScroll is mouse-wheel input.
	PointerScroll
	// Focused is emitted after a widget gains focus.
	Focused
	// Blurred is emitted after a widget loses focus.
	Blurred
	// Changed reports an interactive value change.
	Changed
	// Submitted reports explicit value activation or submission.
	Submitted
	// Dismissed requests that a transient widget be closed.
	Dismissed
)

// EventPhase is the current stage of routed event propagation.
type EventPhase uint8

const (
	// Capture travels from the root toward the target parent.
	Capture EventPhase = iota
	// Target invokes listeners on the original target.
	Target
	// Bubble travels from the target parent toward the root.
	Bubble
)

// Event contains keyboard, pointer, target, and propagation state.
type Event struct {
	Kind  EventKind
	Phase EventPhase

	Target        Widget
	CurrentTarget Widget

	X, Y int
	Key  string
	Rune rune

	KeyCode   tcell.Key
	Buttons   tcell.ButtonMask
	Modifiers tcell.ModMask
	DeltaX    int
	DeltaY    int

	stopped          bool
	immediateStopped bool
	defaultPrevented bool
}

// Handler processes one routed event.
type Handler func(*Event)

type listener struct {
	capture bool
	handler Handler
}

// StopPropagation prevents the event from visiting subsequent widgets.
func (e *Event) StopPropagation() {
	e.stopped = true
}

// StopImmediatePropagation also skips remaining listeners on the current widget.
func (e *Event) StopImmediatePropagation() {
	e.stopped = true
	e.immediateStopped = true
}

// PreventDefault cancels the target widget's default action.
func (e *Event) PreventDefault() {
	e.defaultPrevented = true
}

// PropagationStopped reports whether propagation has been stopped.
func (e *Event) PropagationStopped() bool { return e.stopped }

// DefaultPrevented reports whether the default action was cancelled.
func (e *Event) DefaultPrevented() bool { return e.defaultPrevented }

// Dispatch sends an event through capture, target, and bubble phases. Widgets
// embedding WidgetBase get registered listeners; HandlerEvent is also invoked
// once at each visited phase.
func Dispatch(target Widget, event *Event) {
	if target == nil || event == nil {
		return
	}
	event.Target = target
	path := []Widget{target}
	for current := target; ; {
		base := widgetBase(current)
		if base == nil || base.parent == nil {
			break
		}
		current = base.parent
		path = append(path, current)
	}

	event.Phase = Capture
	for i := len(path) - 1; i >= 1 && !event.stopped; i-- {
		invokeWidget(path[i], event, true, true)
	}
	if !event.stopped {
		event.Phase = Target
		event.immediateStopped = false
		if base := widgetBase(target); base != nil {
			base.invoke(target, event, true)
		}
		if !event.immediateStopped {
			if base := widgetBase(target); base != nil {
				base.invoke(target, event, false)
			}
		}
		if !event.immediateStopped {
			event.CurrentTarget = target
			target.HandlerEvent(event)
		}
	}
	if !event.stopped {
		event.Phase = Bubble
		for i := 1; i < len(path) && !event.stopped; i++ {
			invokeWidget(path[i], event, false, true)
		}
	}
	event.CurrentTarget = nil
	if base := widgetBase(target); !event.defaultPrevented && base != nil && base.DefaultAction != nil {
		base.DefaultAction(event)
	}
}

func invokeWidget(widget Widget, event *Event, capture, callHandler bool) {
	event.immediateStopped = false
	if base := widgetBase(widget); base != nil {
		base.invoke(widget, event, capture)
	}
	if callHandler && !event.immediateStopped {
		event.CurrentTarget = widget
		widget.HandlerEvent(event)
	}
}

// HitTest returns the topmost visible widget containing x,y. Later children
// are treated as visually above earlier children.
func HitTest(root Widget, x, y int) Widget {
	if root == nil {
		return nil
	}
	if base := widgetBase(root); base != nil {
		if base.hidden || !base.bounds.Contains(x, y) {
			return nil
		}
	}
	if provider, ok := root.(childProvider); ok {
		children := provider.EventChildren()
		for i := len(children) - 1; i >= 0; i-- {
			if target := HitTest(children[i], x, y); target != nil {
				return target
			}
		}
	}
	return root
}
