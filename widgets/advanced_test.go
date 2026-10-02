package widgets

import (
	"errors"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestTextAreaMultilineEditing(t *testing.T) {
	area := NewTextArea("ab")
	area.Arrange(NewRect(0, 0, 10, 3))
	changes := 0
	area.On(Changed, false, func(*Event) { changes++ })
	Dispatch(area, &Event{Kind: KeyDown, KeyCode: tcell.KeyEnter})
	Dispatch(area, &Event{Kind: KeyDown, KeyCode: tcell.KeyRune, Rune: '界'})
	Dispatch(area, &Event{Kind: KeyDown, KeyCode: tcell.KeyBackspace2})
	Dispatch(area, &Event{Kind: KeyDown, KeyCode: tcell.KeyBackspace2})
	if area.Value != "ab" || area.CursorRow != 0 || area.CursorCol != 2 || changes != 4 {
		t.Fatalf("value=%q cursor=%d:%d changes=%d", area.Value, area.CursorRow, area.CursorCol, changes)
	}
}

func TestTableNavigationAndSubmission(t *testing.T) {
	table := NewTable(TableColumn{Title: "Name"}, TableColumn{Title: "State", Width: 8})
	table.Rows = [][]string{{"api", "up"}, {"db", "up"}, {"worker", "down"}}
	table.Arrange(NewRect(0, 0, 30, 3))
	submitted := 0
	table.On(Submitted, false, func(*Event) { submitted++ })
	Dispatch(table, &Event{Kind: KeyDown, KeyCode: tcell.KeyDown})
	Dispatch(table, &Event{Kind: KeyDown, KeyCode: tcell.KeyDown})
	Dispatch(table, &Event{Kind: KeyDown, KeyCode: tcell.KeyEnter})
	if table.Selected != 2 || table.Offset != 1 || submitted != 1 {
		t.Fatalf("selected=%d offset=%d submitted=%d", table.Selected, table.Offset, submitted)
	}
}

func TestTreeExpansionAndParentNavigation(t *testing.T) {
	child := &TreeNode{Label: "child"}
	root := &TreeNode{Label: "root", Children: []*TreeNode{child}}
	tree := NewTree(root)
	tree.Arrange(NewRect(0, 0, 20, 4))
	Dispatch(tree, &Event{Kind: KeyDown, KeyCode: tcell.KeyRight})
	Dispatch(tree, &Event{Kind: KeyDown, KeyCode: tcell.KeyDown})
	if tree.SelectedNode() != child {
		t.Fatal("child was not selected")
	}
	Dispatch(tree, &Event{Kind: KeyDown, KeyCode: tcell.KeyLeft})
	if tree.SelectedNode() != root {
		t.Fatal("Left did not select parent")
	}
}

func TestTabsSwitchContent(t *testing.T) {
	first, second := NewText("first"), NewText("second")
	tabs := NewTabs(Tab{Title: "One", Content: first}, Tab{Title: "Two", Content: second})
	tabs.Arrange(NewRect(0, 0, 20, 5))
	Dispatch(tabs, &Event{Kind: KeyDown, KeyCode: tcell.KeyRight})
	if tabs.Active != 1 || len(tabs.EventChildren()) != 1 || tabs.EventChildren()[0] != second {
		t.Fatal("right key did not activate second tab")
	}
}

func TestModalLayoutAndDismiss(t *testing.T) {
	modal := NewModal(NewText("content"))
	modal.Width, modal.Height = 20, 8
	modal.Arrange(NewRect(0, 0, 80, 24))
	if got, want := modal.PanelBounds(), NewRect(30, 8, 20, 8); got != want {
		t.Fatalf("panel=%#v want=%#v", got, want)
	}
	dismissed := 0
	modal.On(Dismissed, false, func(*Event) { dismissed++ })
	Dispatch(modal, &Event{Kind: KeyDown, KeyCode: tcell.KeyEscape})
	Dispatch(modal, &Event{Kind: Click, X: 0, Y: 0})
	if dismissed != 2 {
		t.Fatalf("dismissed=%d", dismissed)
	}
}

func TestFormValidationAndSubmission(t *testing.T) {
	name := NewInput("")
	form := NewForm(FormField{Label: "Name", Input: name, Validate: func(value string) error {
		if value == "" {
			return errors.New("required")
		}
		return nil
	}})
	form.Arrange(NewRect(0, 0, 30, 4))
	submitted := 0
	form.On(Submitted, false, func(*Event) { submitted++ })
	Dispatch(form.Submit, &Event{Kind: Click})
	if submitted != 0 || form.Errors[0] != "required" {
		t.Fatalf("invalid form submitted=%d errors=%v", submitted, form.Errors)
	}
	name.SetValue("Star")
	Dispatch(form.Submit, &Event{Kind: Click})
	if submitted != 1 || len(form.Errors) != 0 || form.Values()[0] != "Star" {
		t.Fatalf("valid form submitted=%d errors=%v values=%v", submitted, form.Errors, form.Values())
	}
}

func TestAdvancedComponentsRenderOnSimulationScreen(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)

	table := NewTable(TableColumn{Title: "A"}, TableColumn{Title: "B", Width: 8})
	table.Rows = [][]string{{"wide 界", "ok"}}
	tree := NewTree(&TreeNode{Label: "root", Expanded: true, Children: []*TreeNode{{Label: "child"}}})
	area := NewTextArea("one\ntwo")
	form := NewForm(FormField{Label: "Name", Input: NewInput("Star")})
	tabs := NewTabs(
		Tab{Title: "Table", Content: table},
		Tab{Title: "Tree", Content: tree},
		Tab{Title: "Editor", Content: area},
		Tab{Title: "Form", Content: form},
	)
	for index := range tabs.Tabs {
		tabs.SetActive(index)
		tabs.Measure(Tight(80, 24))
		tabs.Arrange(NewRect(0, 0, 80, 24))
		tabs.Render(screen)
	}
	modal := NewModal(tabs)
	modal.Arrange(NewRect(0, 0, 80, 24))
	modal.Render(screen)
}
