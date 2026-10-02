package widgets

import "github.com/gdamore/tcell/v2"

// Axis is the main layout direction of a Stack.
type Axis uint8

const (
	// Vertical lays children out from top to bottom.
	Vertical Axis = iota
	// Horizontal lays children out from left to right.
	Horizontal
)

// Stack lays out children along one axis. A child with Basis > 0 receives a
// fixed number of cells; remaining space is divided by Flex (default 1).
type Stack struct {
	WidgetBase
	Axis       Axis
	Gap        int
	Children   []StackChild
	Background tcell.Style
}

// StackChild describes one child and its main-axis allocation.
type StackChild struct {
	Widget Widget
	Basis  int
	Flex   int
}

// NewStack creates a stack on axis with equally flexible initial children.
func NewStack(axis Axis, children ...Widget) *Stack {
	stack := &Stack{Axis: axis, Background: tcell.StyleDefault}
	for _, child := range children {
		stack.Add(child, 0, 1)
	}
	return stack
}

// NewVStack creates a vertical stack.
func NewVStack(children ...Widget) *Stack { return NewStack(Vertical, children...) }

// NewHStack creates a horizontal stack.
func NewHStack(children ...Widget) *Stack { return NewStack(Horizontal, children...) }

// Add appends child. basis > 0 is a fixed main-axis size; otherwise flex is
// its share of the remaining space.
func (s *Stack) Add(child Widget, basis, flex int) *Stack {
	if child == nil {
		return s
	}
	if basis <= 0 && flex <= 0 {
		flex = 1
	}
	s.Children = append(s.Children, StackChild{Widget: child, Basis: max(0, basis), Flex: max(0, flex)})
	if base := widgetBase(child); base != nil {
		base.setParent(s)
	}
	return s
}

func (s *Stack) EventChildren() []Widget {
	children := s.nonNilChildren()
	result := make([]Widget, 0, len(children))
	for _, child := range children {
		result = append(result, child.Widget)
	}
	return result
}

func (s *Stack) Measure(constraints Constraints) Size {
	children := s.nonNilChildren()
	mainSize, crossSize := 0, 0
	for _, child := range children {
		size := child.Widget.Measure(Loose(constraints.MaxW, constraints.MaxH))
		main, cross := size.H, size.W
		if s.Axis == Horizontal {
			main, cross = size.W, size.H
		}
		if child.Basis > 0 {
			main = child.Basis
		}
		mainSize += main
		crossSize = max(crossSize, cross)
	}
	if len(children) > 1 {
		mainSize += max(0, s.Gap) * (len(children) - 1)
	}
	if s.Axis == Horizontal {
		return constraints.Constrain(Size{W: mainSize, H: crossSize})
	}
	return constraints.Constrain(Size{W: crossSize, H: mainSize})
}

func (s *Stack) Arrange(bounds Rect) {
	s.SetBounds(bounds)
	children := s.nonNilChildren()
	if len(children) == 0 {
		return
	}
	extent := bounds.H
	if s.Axis == Horizontal {
		extent = bounds.W
	}
	available := extent
	available -= max(0, s.Gap) * (len(children) - 1)
	available = max(0, available)
	fixed, totalFlex := 0, 0
	for _, child := range children {
		if child.Basis > 0 {
			fixed += child.Basis
		} else {
			totalFlex += max(1, child.Flex)
		}
	}
	remaining := max(0, available-fixed)
	position, flexUsed := 0, 0
	for i, child := range children {
		if base := widgetBase(child.Widget); base != nil {
			base.setParent(s)
		}
		main := child.Basis
		if main <= 0 {
			weight := max(1, child.Flex)
			before := remaining * flexUsed / max(1, totalFlex)
			flexUsed += weight
			main = remaining*flexUsed/max(1, totalFlex) - before
		}
		if position+main > extent {
			main = max(0, extent-position)
		}
		childBounds := Rect{X: bounds.X, Y: bounds.Y + position, W: bounds.W, H: main}
		if s.Axis == Horizontal {
			childBounds = Rect{X: bounds.X + position, Y: bounds.Y, W: main, H: bounds.H}
		}
		child.Widget.Arrange(childBounds)
		position += main
		if i < len(children)-1 {
			position += max(0, s.Gap)
		}
	}
}

func (s *Stack) Render(screen tcell.Screen) {
	if s.Hidden() {
		return
	}
	fill(screen, s.Bounds(), ' ', s.Background)
	for _, child := range s.Children {
		if child.Widget != nil {
			child.Widget.Render(screen)
		}
	}
}

func (s *Stack) HandlerEvent(*Event) {}

func (s *Stack) nonNilChildren() []StackChild {
	children := make([]StackChild, 0, len(s.Children))
	for _, child := range s.Children {
		if child.Widget != nil {
			children = append(children, child)
		}
	}
	return children
}
