package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// TreeNode is one expandable node in a Tree.
type TreeNode struct {
	Label    string
	Children []*TreeNode
	Expanded bool
}

type visibleTreeNode struct {
	node   *TreeNode
	parent *TreeNode
	depth  int
}

// Tree displays and navigates a hierarchy of nodes.
type Tree struct {
	WidgetBase
	Roots         []*TreeNode
	Selected      int
	Offset        int
	Indent        int
	Style         tcell.Style
	SelectedStyle tcell.Style
	Focused       bool
	Disabled      bool
}

// NewTree creates a hierarchy using roots without copying the nodes.
func NewTree(roots ...*TreeNode) *Tree {
	t := &Tree{Roots: append([]*TreeNode(nil), roots...), Indent: 2, Style: tcell.StyleDefault, SelectedStyle: tcell.StyleDefault.Reverse(true)}
	t.DefaultAction = t.defaultAction
	return t
}

func (t *Tree) CanFocus() bool      { return !t.Disabled && !t.Hidden() }
func (t *Tree) SetFocus(value bool) { t.Focused = value }
func (t *Tree) HandlerEvent(*Event) {}

// SelectedNode returns the visible selected node, or nil for an empty tree.
func (t *Tree) SelectedNode() *TreeNode {
	nodes := t.visibleNodes()
	if len(nodes) == 0 {
		return nil
	}
	t.Selected = clamp(t.Selected, 0, len(nodes)-1)
	return nodes[t.Selected].node
}

func (t *Tree) Measure(c Constraints) Size {
	nodes := t.visibleNodes()
	width := 0
	for _, item := range nodes {
		width = max(width, item.depth*max(0, t.Indent)+2+uniseg.StringWidth(item.node.Label))
	}
	return c.Constrain(Size{W: width, H: len(nodes)})
}

func (t *Tree) Arrange(bounds Rect) {
	t.SetBounds(bounds)
	t.ensureVisible()
}

func (t *Tree) Render(screen tcell.Screen) {
	if t.Hidden() || t.Bounds().Empty() {
		return
	}
	fill(screen, t.Bounds(), ' ', t.Style)
	nodes := t.visibleNodes()
	t.ensureVisibleFor(len(nodes))
	for row := 0; row < t.Bounds().H && t.Offset+row < len(nodes); row++ {
		index := t.Offset + row
		item := nodes[index]
		style := t.Style
		if index == t.Selected {
			style = t.SelectedStyle
			fill(screen, NewRect(t.Bounds().X, t.Bounds().Y+row, t.Bounds().W, 1), ' ', style)
		}
		prefix := "  "
		if len(item.node.Children) > 0 {
			if item.node.Expanded {
				prefix = "▾ "
			} else {
				prefix = "▸ "
			}
		}
		x := t.Bounds().X + item.depth*max(0, t.Indent)
		drawText(screen, x, t.Bounds().Y+row, max(0, t.Bounds().W-(x-t.Bounds().X)), prefix+item.node.Label, style)
	}
}

func (t *Tree) visibleNodes() []visibleTreeNode {
	var result []visibleTreeNode
	var visit func([]*TreeNode, *TreeNode, int)
	visit = func(nodes []*TreeNode, parent *TreeNode, depth int) {
		for _, node := range nodes {
			if node == nil {
				continue
			}
			result = append(result, visibleTreeNode{node: node, parent: parent, depth: depth})
			if node.Expanded {
				visit(node.Children, node, depth+1)
			}
		}
	}
	visit(t.Roots, nil, 0)
	return result
}

func (t *Tree) ensureVisible() { t.ensureVisibleFor(len(t.visibleNodes())) }

func (t *Tree) ensureVisibleFor(count int) {
	t.Selected = clamp(t.Selected, 0, max(0, count-1))
	if t.Selected < t.Offset {
		t.Offset = t.Selected
	} else if t.Bounds().H > 0 && t.Selected >= t.Offset+t.Bounds().H {
		t.Offset = t.Selected - t.Bounds().H + 1
	}
	t.Offset = clamp(t.Offset, 0, max(0, count-max(1, t.Bounds().H)))
}

func (t *Tree) defaultAction(event *Event) {
	if t.Disabled {
		return
	}
	nodes := t.visibleNodes()
	if len(nodes) == 0 {
		return
	}
	previous := t.Selected
	switch event.Kind {
	case KeyDown:
		switch event.KeyCode {
		case tcell.KeyUp:
			t.Selected--
		case tcell.KeyDown:
			t.Selected++
		case tcell.KeyHome:
			t.Selected = 0
		case tcell.KeyEnd:
			t.Selected = len(nodes) - 1
		case tcell.KeyRight:
			if len(nodes[t.Selected].node.Children) > 0 {
				nodes[t.Selected].node.Expanded = true
			}
		case tcell.KeyLeft:
			item := nodes[t.Selected]
			if item.node.Expanded {
				item.node.Expanded = false
			} else if item.parent != nil {
				for index, candidate := range nodes {
					if candidate.node == item.parent {
						t.Selected = index
						break
					}
				}
			}
		case tcell.KeyEnter:
			item := nodes[t.Selected].node
			if len(item.Children) > 0 {
				item.Expanded = !item.Expanded
			}
			Dispatch(t, &Event{Kind: Submitted})
		}
	case PointerScroll:
		t.Selected += event.DeltaY
	case Click:
		row := event.Y - t.Bounds().Y
		if row >= 0 && row < t.Bounds().H && t.Offset+row < len(nodes) {
			t.Selected = t.Offset + row
			item := nodes[t.Selected].node
			if len(item.Children) > 0 {
				item.Expanded = !item.Expanded
			}
			Dispatch(t, &Event{Kind: Submitted})
		}
	}
	t.ensureVisible()
	if previous != t.Selected {
		Dispatch(t, &Event{Kind: Changed})
	}
}
