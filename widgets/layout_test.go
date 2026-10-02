package widgets

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestStackAllocatesFixedAndFlexibleSpace(t *testing.T) {
	one, two, fixed := NewText("1"), NewText("2"), NewText("fixed")
	stack := NewHStack()
	stack.Add(one, 0, 1).Add(two, 0, 2).Add(fixed, 3, 0)
	stack.Gap = 1
	stack.Arrange(NewRect(0, 0, 20, 2))

	if got, want := one.Bounds(), NewRect(0, 0, 5, 2); got != want {
		t.Fatalf("first bounds = %#v, want %#v", got, want)
	}
	if got, want := two.Bounds(), NewRect(6, 0, 10, 2); got != want {
		t.Fatalf("second bounds = %#v, want %#v", got, want)
	}
	if got, want := fixed.Bounds(), NewRect(17, 0, 3, 2); got != want {
		t.Fatalf("fixed bounds = %#v, want %#v", got, want)
	}
}

func TestTextRendersUnicodeAndClips(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(4, 1)
	text := NewText("A界B")
	text.Arrange(NewRect(0, 0, 3, 1))
	text.Render(screen)
	main, _, _, _ := screen.GetContent(0, 0)
	wide, _, _, _ := screen.GetContent(1, 0)
	clipped, _, _, _ := screen.GetContent(3, 0)
	if main != 'A' || wide != '界' || clipped == 'B' {
		t.Fatalf("unexpected cells: %q %q %q", main, wide, clipped)
	}
}

func TestDispatchOrderAndDefault(t *testing.T) {
	root := NewVStack()
	child := NewButton("child")
	root.Add(child, 1, 0)
	root.Arrange(NewRect(0, 0, 10, 2))
	var order []string
	root.On(Click, true, func(*Event) { order = append(order, "root-capture") })
	child.On(Click, true, func(*Event) { order = append(order, "target-capture") })
	child.On(Click, false, func(*Event) { order = append(order, "target") })
	root.On(Click, false, func(*Event) { order = append(order, "root-bubble") })
	child.DefaultAction = func(*Event) { order = append(order, "default") }

	if got := HitTest(root, 1, 0); got != child {
		t.Fatalf("HitTest = %T, want child", got)
	}
	Dispatch(child, &Event{Kind: Click, X: 1, Y: 0})
	want := []string{"root-capture", "target-capture", "target", "root-bubble", "default"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
}

func TestRouterFocusInputAndClick(t *testing.T) {
	input := NewInput("")
	button := NewButton("save")
	root := NewVStack()
	root.Add(input, 1, 0).Add(button, 1, 0)
	root.Arrange(NewRect(0, 0, 10, 2))
	router := NewRouter(root)

	if !router.FocusNext() || router.Focused != input {
		t.Fatal("first focus should select input")
	}
	router.HandleKey(tcell.KeyRune, '界', tcell.ModNone)
	if input.Value != "界" {
		t.Fatalf("input value = %q", input.Value)
	}
	router.HandleKey(tcell.KeyTAB, 0, tcell.ModNone)
	if router.Focused != button || !button.Focused {
		t.Fatal("Tab should focus button")
	}
	clicks := 0
	button.On(Click, false, func(*Event) { clicks++ })
	router.HandleKey(tcell.KeyEnter, 0, tcell.ModNone)
	if clicks != 1 {
		t.Fatalf("keyboard clicks = %d, want 1", clicks)
	}
	router.HandleMouse(1, 1, tcell.Button1, tcell.ModNone)
	router.HandleMouse(1, 1, tcell.ButtonNone, tcell.ModNone)
	if clicks != 2 {
		t.Fatalf("mouse clicks = %d, want 2", clicks)
	}
}

func TestBoxArrangesInsets(t *testing.T) {
	child := NewText("inside")
	box := NewBox(child)
	box.Border = &RoundedBorder
	box.Padding = UniformInsets(1)
	box.Arrange(NewRect(2, 3, 20, 8))
	if got, want := child.Bounds(), NewRect(4, 5, 16, 4); got != want {
		t.Fatalf("child bounds = %#v, want %#v", got, want)
	}
}

func TestScrollViewClampsOffset(t *testing.T) {
	view := NewScrollView("one\ntwo\nthree\nfour")
	view.Arrange(NewRect(0, 0, 10, 2))
	view.ScrollTo(99)
	if view.Offset != 2 {
		t.Fatalf("offset = %d, want 2", view.Offset)
	}
	Dispatch(view, &Event{Kind: PointerScroll, DeltaY: -1})
	if view.Offset != 1 {
		t.Fatalf("offset after wheel = %d, want 1", view.Offset)
	}
}

func TestInputEditingAndPreventDefault(t *testing.T) {
	input := NewInput("")
	changes := 0
	input.On(Changed, false, func(*Event) { changes++ })
	Dispatch(input, &Event{Kind: KeyDown, KeyCode: tcell.KeyRune, Rune: 'a'})
	Dispatch(input, &Event{Kind: KeyDown, KeyCode: tcell.KeyRune, Rune: '界'})
	Dispatch(input, &Event{Kind: KeyDown, KeyCode: tcell.KeyLeft})
	Dispatch(input, &Event{Kind: KeyDown, KeyCode: tcell.KeyBackspace2})
	if input.Value != "界" || input.Cursor != 0 || changes != 3 {
		t.Fatalf("input = %q cursor=%d changes=%d", input.Value, input.Cursor, changes)
	}
	input.On(KeyDown, false, func(event *Event) {
		if event.Rune == 'x' {
			event.PreventDefault()
		}
	})
	Dispatch(input, &Event{Kind: KeyDown, KeyCode: tcell.KeyRune, Rune: 'x'})
	if input.Value != "界" {
		t.Fatalf("prevented input changed to %q", input.Value)
	}
}

func TestListNavigationAndSubmission(t *testing.T) {
	list := NewList("a", "b", "c", "d")
	list.Arrange(NewRect(0, 0, 5, 2))
	submitted := 0
	list.On(Submitted, false, func(*Event) { submitted++ })
	Dispatch(list, &Event{Kind: KeyDown, KeyCode: tcell.KeyDown})
	Dispatch(list, &Event{Kind: KeyDown, KeyCode: tcell.KeyDown})
	Dispatch(list, &Event{Kind: KeyDown, KeyCode: tcell.KeyEnter})
	if list.Selected != 2 || list.Offset != 1 || submitted != 1 {
		t.Fatalf("selected=%d offset=%d submitted=%d", list.Selected, list.Offset, submitted)
	}
}

func TestTextWrapAndOverlayHitOrder(t *testing.T) {
	text := NewText("A界B")
	text.Wrap = true
	if size := text.Measure(Loose(2, 10)); size != (Size{W: 2, H: 3}) {
		t.Fatalf("wrapped size = %#v", size)
	}
	bottom, top := NewText("bottom"), NewButton("top")
	overlay := NewOverlay(bottom, top)
	overlay.Arrange(NewRect(0, 0, 10, 2))
	if got := HitTest(overlay, 1, 0); got != top {
		t.Fatalf("top hit = %T", got)
	}
}

func TestCanvasResizePreservesCells(t *testing.T) {
	canvas := NewCanvas(2, 2)
	canvas.SetCell(1, 1, Cell{Main: 'X'})
	canvas.Resize(3, 3)
	cell, ok := canvas.Cell(1, 1)
	if !ok || cell.Main != 'X' {
		t.Fatalf("cell = %#v, %v", cell, ok)
	}
	if canvas.SetCell(9, 9, Cell{}) {
		t.Fatal("out-of-range SetCell succeeded")
	}
}

type customContainer struct {
	WidgetBase
	child Widget
}

func (c *customContainer) Measure(constraints Constraints) Size {
	return c.child.Measure(constraints)
}
func (c *customContainer) Arrange(bounds Rect) {
	c.SetBounds(bounds)
	if provider, ok := c.child.(interface{ EventBase() *WidgetBase }); ok {
		provider.EventBase().SetParent(c)
	}
	c.child.Arrange(bounds)
}
func (c *customContainer) Render(screen tcell.Screen) { c.child.Render(screen) }
func (c *customContainer) HandlerEvent(*Event)        {}
func (c *customContainer) EventChildren() []Widget    { return []Widget{c.child} }

func TestCustomContainerCanBindEventParent(t *testing.T) {
	child := NewButton("child")
	container := &customContainer{child: child}
	container.Arrange(NewRect(0, 0, 10, 1))
	bubbled := false
	container.On(Click, false, func(*Event) { bubbled = true })
	Dispatch(child, &Event{Kind: Click})
	if !bubbled || child.Parent() != container {
		t.Fatal("custom parent did not participate in event propagation")
	}
}
