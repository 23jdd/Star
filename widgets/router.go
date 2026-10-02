package widgets

import (
	"reflect"

	"github.com/gdamore/tcell/v2"
)

// Focusable is implemented by widgets that participate in keyboard focus.
type Focusable interface {
	CanFocus() bool
	SetFocus(bool)
}

type keyboardActivatable interface{ KeyboardActivatable() bool }

// Router translates terminal input into routed widget events, maintains focus,
// captures the pointer during a press, and synthesizes Click events.
type Router struct {
	Root Widget

	Focused         Widget
	PointerCaptured Widget
	PressedTarget   Widget
	// DragThreshold is the number of cells the pointer may move while still
	// producing a click. The default is zero.
	DragThreshold int

	buttons tcell.ButtonMask
	pressX  int
	pressY  int
	dragged bool
}

// NewRouter creates an input router for root.
func NewRouter(root Widget) *Router { return &Router{Root: root} }

// SetRoot changes the routed tree and clears focus if it no longer exists.
func (r *Router) SetRoot(root Widget) {
	r.Root = root
	if r.Focused != nil && !containsWidget(root, r.Focused) {
		r.SetFocus(nil)
	}
}

// SetFocus changes focus and dispatches Blurred/Focused events.
func (r *Router) SetFocus(widget Widget) bool {
	if widget != nil {
		focusable, ok := widget.(Focusable)
		if !ok || !focusable.CanFocus() {
			return false
		}
	}
	if sameWidget(widget, r.Focused) {
		return true
	}
	previous := r.Focused
	r.Focused = widget
	if focusable, ok := previous.(Focusable); ok {
		focusable.SetFocus(false)
		Dispatch(previous, &Event{Kind: Blurred})
	}
	if focusable, ok := widget.(Focusable); ok {
		focusable.SetFocus(true)
		Dispatch(widget, &Event{Kind: Focused})
	}
	return true
}

// FocusNext advances focus in depth-first order, wrapping at the end.
func (r *Router) FocusNext() bool { return r.moveFocus(1) }

// FocusPrevious moves focus backward, wrapping at the beginning.
func (r *Router) FocusPrevious() bool { return r.moveFocus(-1) }

func (r *Router) moveFocus(direction int) bool {
	order := focusOrder(r.Root, nil)
	if len(order) == 0 {
		return r.SetFocus(nil)
	}
	index := -1
	for i, widget := range order {
		if sameWidget(widget, r.Focused) {
			index = i
			break
		}
	}
	if direction < 0 {
		if index < 0 {
			index = 0
		}
		index = (index - 1 + len(order)) % len(order)
	} else {
		index = (index + 1) % len(order)
	}
	return r.SetFocus(order[index])
}

// HandleKey dispatches a key and performs default Tab navigation and keyboard
// activation. It returns true when a widget received the key.
func (r *Router) HandleKey(key tcell.Key, char rune, modifiers tcell.ModMask) bool {
	target := r.Focused
	if target == nil {
		target = r.Root
	}
	if target == nil {
		return false
	}
	event := &Event{Kind: KeyDown, Key: tcell.KeyNames[key], KeyCode: key, Rune: char, Modifiers: modifiers}
	Dispatch(target, event)
	if event.DefaultPrevented() {
		return true
	}
	if key == tcell.KeyTAB || key == tcell.KeyBacktab {
		if key == tcell.KeyBacktab || modifiers&tcell.ModShift != 0 {
			r.FocusPrevious()
		} else {
			r.FocusNext()
		}
	} else if activation, ok := target.(keyboardActivatable); ok && activation.KeyboardActivatable() && (key == tcell.KeyEnter || char == ' ') {
		Dispatch(target, &Event{Kind: Click, KeyCode: key, Rune: char, Modifiers: modifiers})
	}
	return true
}

// HandleMouse routes movement, presses, releases, wheel input, and clicks.
func (r *Router) HandleMouse(x, y int, buttons tcell.ButtonMask, modifiers tcell.ModMask) bool {
	hit := HitTest(r.Root, x, y)
	target := hit
	if r.PointerCaptured != nil {
		target = r.PointerCaptured
	}
	if target == nil {
		r.buttons = buttons
		return false
	}

	for mask, deltaY := range map[tcell.ButtonMask]int{
		tcell.WheelUp: -1, tcell.WheelDown: 1,
	} {
		if buttons&mask != 0 {
			Dispatch(target, &Event{Kind: PointerScroll, X: x, Y: y, Buttons: buttons, Modifiers: modifiers, DeltaY: deltaY})
		}
	}
	Dispatch(target, &Event{Kind: PointerMove, X: x, Y: y, Buttons: buttons, Modifiers: modifiers})

	pressed := buttons&tcell.Button1 != 0 && r.buttons&tcell.Button1 == 0
	released := buttons&tcell.Button1 == 0 && r.buttons&tcell.Button1 != 0
	if pressed {
		r.PressedTarget = hit
		r.PointerCaptured = hit
		r.pressX, r.pressY, r.dragged = x, y, false
		r.focusNearest(hit)
		Dispatch(hit, &Event{Kind: PointerDown, X: x, Y: y, Buttons: buttons, Modifiers: modifiers})
	}
	if r.PointerCaptured != nil && (abs(x-r.pressX) > max(0, r.DragThreshold) || abs(y-r.pressY) > max(0, r.DragThreshold)) {
		r.dragged = true
	}
	if released {
		captured, pressedTarget := r.PointerCaptured, r.PressedTarget
		Dispatch(captured, &Event{Kind: PointerUp, X: x, Y: y, Buttons: buttons, Modifiers: modifiers})
		if !r.dragged && hit != nil && sameWidget(hit, pressedTarget) {
			Dispatch(pressedTarget, &Event{Kind: Click, X: x, Y: y, Buttons: buttons, Modifiers: modifiers})
		}
		r.PointerCaptured, r.PressedTarget = nil, nil
		r.dragged = false
	}
	r.buttons = buttons
	return true
}

func (r *Router) focusNearest(widget Widget) {
	for widget != nil {
		if focusable, ok := widget.(Focusable); ok && focusable.CanFocus() {
			r.SetFocus(widget)
			return
		}
		base := widgetBase(widget)
		if base == nil {
			return
		}
		widget = base.parent
	}
}

func focusOrder(root Widget, result []Widget) []Widget {
	if root == nil {
		return result
	}
	if focusable, ok := root.(Focusable); ok && focusable.CanFocus() {
		result = append(result, root)
	}
	if provider, ok := root.(childProvider); ok {
		for _, child := range provider.EventChildren() {
			result = focusOrder(child, result)
		}
	}
	return result
}

func containsWidget(root, target Widget) bool {
	if root == nil || target == nil {
		return false
	}
	if sameWidget(root, target) {
		return true
	}
	if provider, ok := root.(childProvider); ok {
		for _, child := range provider.EventChildren() {
			if containsWidget(child, target) {
				return true
			}
		}
	}
	return false
}

func sameWidget(left, right Widget) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	leftType, rightType := reflect.TypeOf(left), reflect.TypeOf(right)
	if leftType != rightType || !leftType.Comparable() {
		return false
	}
	return left == right
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
