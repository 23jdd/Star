package widgets

// Constraints defines the inclusive minimum and maximum widget size.
type Constraints struct {
	MinW, MinH int
	MaxW, MaxH int
}

// Rect is an absolute terminal-cell rectangle.
type Rect struct {
	X, Y, W, H int
}

// Size is a width and height measured in terminal cells.
type Size struct {
	W, H int
}

// NewRect constructs a rectangle and normalizes negative dimensions to zero.
func NewRect(x, y, w, h int) Rect {
	return Rect{X: x, Y: y, W: max(0, w), H: max(0, h)}
}

// Tight creates constraints that require exactly w by h cells.
func Tight(w, h int) Constraints {
	w, h = max(0, w), max(0, h)
	return Constraints{MinW: w, MinH: h, MaxW: w, MaxH: h}
}

// Loose creates constraints with no minimum size.
func Loose(w, h int) Constraints {
	return Constraints{MaxW: max(0, w), MaxH: max(0, h)}
}

// Constrain clamps size to the normalized constraint range.
func (c Constraints) Constrain(size Size) Size {
	minW, minH := max(0, c.MinW), max(0, c.MinH)
	maxW, maxH := max(0, c.MaxW), max(0, c.MaxH)
	if maxW < minW {
		maxW = minW
	}
	if maxH < minH {
		maxH = minH
	}
	return Size{W: clamp(size.W, minW, maxW), H: clamp(size.H, minH, maxH)}
}

// Contains reports whether a cell lies inside the rectangle.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// Empty reports whether the rectangle has no drawable cells.
func (r Rect) Empty() bool { return r.W <= 0 || r.H <= 0 }

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
