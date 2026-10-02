package star

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
)

// KeyMsg represents a key press. Rune is set for printable keys.
type KeyMsg struct {
	Key       tcell.Key
	Rune      rune
	Modifiers tcell.ModMask
}

// String returns the tcell name of the key event.
func (m KeyMsg) String() string {
	return tcell.NewEventKey(m.Key, m.Rune, m.Modifiers).Name()
}

// MouseMsg represents a mouse event at terminal cell coordinates.
type MouseMsg struct {
	X, Y      int
	Buttons   tcell.ButtonMask
	Modifiers tcell.ModMask
}

// Position returns the mouse cell coordinates.
func (m MouseMsg) Position() (int, int) { return m.X, m.Y }

// WindowSizeMsg is sent at startup and whenever the terminal is resized.
type WindowSizeMsg struct{ Width, Height int }

// PasteMsg reports bracketed-paste start/end state.
type PasteMsg struct{ Start bool }

// FocusMsg reports whether the terminal gained or lost focus.
type FocusMsg struct{ Focused bool }

// RawEventMsg preserves tcell events that Star does not recognize yet.
type RawEventMsg struct{ Event tcell.Event }

type quitMsg struct{}
type batchMsg []Cmd
type sequenceMsg []Cmd
type contextCmdMsg struct {
	run func(context.Context) Message
}

// Quit requests a clean shutdown of the program.
func Quit() Message { return quitMsg{} }

// QuitCmd requests a clean shutdown from Model.Init or Model.Update.
func QuitCmd() Cmd { return func() Message { return Quit() } }

// Batch starts all non-nil commands concurrently.
func Batch(cmds ...Cmd) Cmd {
	filtered := make(batchMsg, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			filtered = append(filtered, cmd)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return func() Message { return filtered }
}

// Sequence runs commands one at a time in the supplied order. Nil commands
// are ignored. Messages remain ordered, although terminal input may interleave.
func Sequence(cmds ...Cmd) Cmd {
	filtered := make(sequenceMsg, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			filtered = append(filtered, cmd)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return func() Message { return filtered }
}

// WithContext creates a command whose work receives the Program lifecycle
// context. The context is cancelled when Run exits for any reason.
func WithContext(fn func(context.Context) Message) Cmd {
	if fn == nil {
		return nil
	}
	return func() Message { return contextCmdMsg{run: fn} }
}

// After returns a command that produces msg after d.
func After(d time.Duration, msg Message) Cmd {
	return func() Message {
		if d > 0 {
			time.Sleep(d)
		}
		return msg
	}
}

// Tick invokes fn once after d with the actual firing time.
func Tick(d time.Duration, fn func(time.Time) Message) Cmd {
	return func() Message {
		if d > 0 {
			time.Sleep(d)
		}
		if fn == nil {
			return nil
		}
		return fn(time.Now())
	}
}
