package star

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/23jdd/Star/widgets"
	"github.com/gdamore/tcell/v2"
)

type testModel struct {
	messages chan Message
	init     Cmd
}

func (m *testModel) Init() Cmd { return m.init }
func (m *testModel) Update(msg Message) (Model, Cmd) {
	m.messages <- msg
	if key, ok := msg.(KeyMsg); ok && key.Rune == 'q' {
		return m, QuitCmd()
	}
	return m, nil
}
func (m *testModel) View() widgets.Widget { return widgets.NewText("ready") }

func TestProgramTranslatesEventsAndQuits(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.SetSize(20, 4)
	model := &testModel{messages: make(chan Message, 4)}
	program := NewProgram(WithScreen(screen))
	done := make(chan error, 1)
	go func() { done <- program.Run(model) }()

	select {
	case msg := <-model.messages:
		if _, ok := msg.(WindowSizeMsg); !ok {
			t.Fatalf("first message is %T, want WindowSizeMsg", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("program did not start")
	}

	screen.PostEventWait(tcell.NewEventResize(30, 6))
	select {
	case msg := <-model.messages:
		resize, ok := msg.(WindowSizeMsg)
		if !ok || resize.Width != 30 || resize.Height != 6 {
			t.Fatalf("resize message = %#v", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("resize was not delivered")
	}

	screen.PostEventWait(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("program did not quit")
	}
}

func TestProgramReturnsCommandPanic(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	model := &testModel{
		messages: make(chan Message, 2),
		init: func() Message {
			panic("boom")
		},
	}
	err := NewProgram(WithScreen(screen)).Run(model)
	var panicErr *PanicError
	if !errors.As(err, &panicErr) || panicErr.Value != "boom" {
		t.Fatalf("Run returned %#v, want PanicError(boom)", err)
	}
}

func TestBatchFiltersNilCommands(t *testing.T) {
	cmd := Batch(nil, func() Message { return "ok" })
	if cmd == nil {
		t.Fatal("Batch returned nil")
	}
	commands, ok := cmd().(batchMsg)
	if !ok || len(commands) != 1 || commands[0]() != "ok" {
		t.Fatalf("unexpected batch: %#v", commands)
	}
	if Batch(nil) != nil {
		t.Fatal("empty Batch should return nil")
	}
}

func TestSequencePreservesCommandOrder(t *testing.T) {
	cmd := Sequence(
		func() Message { return 1 },
		nil,
		func() Message { return 2 },
	)
	commands, ok := cmd().(sequenceMsg)
	if !ok || len(commands) != 2 || commands[0]() != 1 || commands[1]() != 2 {
		t.Fatalf("unexpected sequence: %#v", commands)
	}
}

func TestSendWhenStopped(t *testing.T) {
	program := NewProgram()
	if !errors.Is(program.Send("message"), ErrNotRunning) {
		t.Fatal("Send should return ErrNotRunning")
	}
	if program.TrySend("message") {
		t.Fatal("TrySend should fail while stopped")
	}
}

func TestProgramExternalSend(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	model := &testModel{messages: make(chan Message, 4)}
	program := NewProgram(WithScreen(screen))
	done := make(chan error, 1)
	go func() { done <- program.Run(model) }()
	<-model.messages // initial size
	if !program.TrySend("external") {
		t.Fatal("TrySend failed while running")
	}
	if got := <-model.messages; got != "external" {
		t.Fatalf("message = %#v", got)
	}
	if err := program.Send(Quit()); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestRunContextAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	model := &testModel{messages: make(chan Message, 1)}
	err := NewProgram().RunContext(ctx, model)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunContext returned %v", err)
	}
}

func TestContextCommandCancelledWhenProgramQuits(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	started := make(chan struct{})
	cancelled := make(chan struct{})
	model := &testModel{
		messages: make(chan Message, 2),
		init: WithContext(func(ctx context.Context) Message {
			close(started)
			<-ctx.Done()
			close(cancelled)
			return nil
		}),
	}
	program := NewProgram(WithScreen(screen))
	done := make(chan error, 1)
	go func() { done <- program.Run(model) }()
	<-model.messages // initial size

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("context command did not start")
	}
	if err := program.Send(Quit()); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("context command was not cancelled")
	}
}
