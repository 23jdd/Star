package star

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/23jdd/Star/widgets"
	"github.com/gdamore/tcell/v2"
)

var (
	ErrNilModel   = errors.New("star: nil model")
	ErrNotRunning = errors.New("star: program is not running")
	ErrRunning    = errors.New("star: program is already running")
)

// PanicError wraps a panic raised by Model or Cmd code. The terminal is
// restored before Run returns this error.
type PanicError struct{ Value any }

func (e *PanicError) Error() string { return fmt.Sprintf("star: panic: %v", e.Value) }

type programConfig struct {
	screen    tcell.Screen
	queueSize int
	mouse     bool
	paste     bool
	focus     bool
}

// Option configures a Program.
type Option func(*programConfig)

// WithScreen supplies a screen, most commonly tcell.NewSimulationScreen in
// tests. Star still initializes and finalizes it.
func WithScreen(screen tcell.Screen) Option {
	return func(c *programConfig) { c.screen = screen }
}

// WithQueueSize changes the internal message buffer. The default is 256.
func WithQueueSize(size int) Option {
	return func(c *programConfig) {
		if size > 0 {
			c.queueSize = size
		}
	}
}

// WithMouse controls whether the terminal reports mouse activity.
func WithMouse(enabled bool) Option { return func(c *programConfig) { c.mouse = enabled } }

// WithPaste controls whether bracketed-paste boundary events are reported.
func WithPaste(enabled bool) Option { return func(c *programConfig) { c.paste = enabled } }

// WithFocus controls whether terminal focus changes are reported.
func WithFocus(enabled bool) Option { return func(c *programConfig) { c.focus = enabled } }

// Program owns a single terminal session.
type Program struct {
	config programConfig

	mu      sync.RWMutex
	running bool
	input   chan Message
	done    chan struct{}
}

// NewProgram constructs a reusable, but not concurrently runnable, Program.
func NewProgram(options ...Option) *Program {
	c := programConfig{queueSize: 256}
	for _, option := range options {
		if option != nil {
			option(&c)
		}
	}
	return &Program{config: c}
}

// Send safely injects a message from another goroutine. It may wait for queue
// capacity; use SendContext or TrySend when the caller must not block.
func (p *Program) Send(msg Message) error {
	return p.SendContext(context.Background(), msg)
}

// SendContext injects a message or returns when ctx is cancelled.
func (p *Program) SendContext(ctx context.Context, msg Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	p.mu.RLock()
	input, done, running := p.input, p.done, p.running
	p.mu.RUnlock()
	if !running {
		return ErrNotRunning
	}
	select {
	case input <- msg:
		return nil
	case <-done:
		return ErrNotRunning
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TrySend attempts to inject a message without blocking.
func (p *Program) TrySend(msg Message) bool {
	p.mu.RLock()
	input, done, running := p.input, p.done, p.running
	p.mu.RUnlock()
	if !running {
		return false
	}
	select {
	case input <- msg:
		return true
	case <-done:
		return false
	default:
		return false
	}
}

func (p *Program) initScreen() (tcell.Screen, error) {
	if p.config.screen != nil {
		if err := p.config.screen.Init(); err != nil {
			return nil, err
		}
		return p.config.screen, nil
	}
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	return screen, nil
}

// Run executes model until QuitCmd, Quit, context cancellation, or a fatal
// screen error.
func (p *Program) Run(model Model) error { return p.RunContext(context.Background(), model) }

// RunContext executes model and always restores the screen before returning.
func (p *Program) RunContext(ctx context.Context, model Model) (err error) {
	if model == nil {
		return ErrNilModel
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan struct{})
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return ErrRunning
	}
	p.running = true
	p.input = make(chan Message, p.config.queueSize)
	p.done = done
	input := p.input
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		p.running = false
		p.input = nil
		p.done = nil
		p.mu.Unlock()
		if value := recover(); value != nil {
			err = &PanicError{Value: value}
		}
	}()
	defer close(done)

	screen, err := p.initScreen()
	if err != nil {
		return err
	}
	defer screen.Fini()
	screen.HideCursor()
	if p.config.mouse {
		screen.EnableMouse()
	}
	if p.config.paste {
		screen.EnablePaste()
	}
	if p.config.focus {
		screen.EnableFocus()
	}

	events := make(chan tcell.Event, 64)
	go pollEvents(screen, events, done)

	w, h := screen.Size()
	model, cmd := model.Update(WindowSizeMsg{Width: w, Height: h})
	if model == nil {
		return ErrNilModel
	}
	if err := render(screen, model.View()); err != nil {
		return err
	}
	launch(input, done, cmd)
	launch(input, done, model.Init())

	for {
		select {
		case <-runCtx.Done():
			screen.PostEvent(tcell.NewEventInterrupt(nil))
			return runCtx.Err()
		case event := <-events:
			if _, resized := event.(*tcell.EventResize); resized {
				screen.Sync()
			}
			msg, eventErr := translateEvent(event)
			if eventErr != nil {
				return eventErr
			}
			if msg == nil {
				continue
			}
			var quit bool
			model, cmd, quit, err = update(model, msg)
			if err != nil || quit {
				return err
			}
			if model == nil {
				return ErrNilModel
			}
			launch(input, done, cmd)
			if err := render(screen, model.View()); err != nil {
				return err
			}
		case msg := <-input:
			if cmds, ok := msg.(batchMsg); ok {
				for _, batchCmd := range cmds {
					launch(input, done, batchCmd)
				}
				continue
			}
			if cmds, ok := msg.(sequenceMsg); ok {
				launchSequence(input, done, runCtx, cmds)
				continue
			}
			if command, ok := msg.(contextCmdMsg); ok {
				launchContextCommand(input, done, runCtx, command.run)
				continue
			}
			var quit bool
			model, cmd, quit, err = update(model, msg)
			if err != nil || quit {
				return err
			}
			if model == nil {
				return ErrNilModel
			}
			launch(input, done, cmd)
			if err := render(screen, model.View()); err != nil {
				return err
			}
		}
	}
}

func update(model Model, msg Message) (Model, Cmd, bool, error) {
	switch msg := msg.(type) {
	case quitMsg:
		return model, nil, true, nil
	case cmdPanicMsg:
		return model, nil, false, &PanicError{Value: msg.value}
	default:
		next, cmd := model.Update(msg)
		return next, cmd, false, nil
	}
}

type cmdPanicMsg struct{ value any }

func launch(input chan<- Message, done <-chan struct{}, cmd Cmd) {
	if cmd == nil {
		return
	}
	go func() {
		defer func() {
			if value := recover(); value != nil {
				sendCommandResult(input, done, cmdPanicMsg{value})
			}
		}()
		if msg := cmd(); msg != nil {
			sendCommandResult(input, done, msg)
		}
	}()
}

func sendCommandResult(input chan<- Message, done <-chan struct{}, msg Message) {
	select {
	case input <- msg:
	case <-done:
	}
}

func launchSequence(input chan<- Message, done <-chan struct{}, ctx context.Context, commands sequenceMsg) {
	go func() {
		for _, command := range commands {
			if command == nil {
				continue
			}
			var result Message
			func() {
				defer func() {
					if value := recover(); value != nil {
						result = cmdPanicMsg{value}
					}
				}()
				result = command()
			}()
			if contextual, ok := result.(contextCmdMsg); ok {
				result = runContextCommand(ctx, contextual.run)
			}
			if result != nil {
				sendCommandResult(input, done, result)
			}
			switch result.(type) {
			case cmdPanicMsg, quitMsg:
				return
			}
			select {
			case <-done:
				return
			default:
			}
		}
	}()
}

func launchContextCommand(input chan<- Message, done <-chan struct{}, ctx context.Context, command func(context.Context) Message) {
	go func() {
		result := runContextCommand(ctx, command)
		if result != nil {
			sendCommandResult(input, done, result)
		}
	}()
}

func runContextCommand(ctx context.Context, command func(context.Context) Message) (result Message) {
	defer func() {
		if value := recover(); value != nil {
			result = cmdPanicMsg{value}
		}
	}()
	if command == nil {
		return nil
	}
	return command(ctx)
}

func pollEvents(screen tcell.Screen, events chan<- tcell.Event, done <-chan struct{}) {
	for {
		event := screen.PollEvent()
		select {
		case events <- event:
		case <-done:
			return
		}
		if event == nil {
			return
		}
	}
}

func translateEvent(event tcell.Event) (Message, error) {
	switch event := event.(type) {
	case *tcell.EventKey:
		return KeyMsg{Key: event.Key(), Rune: event.Rune(), Modifiers: event.Modifiers()}, nil
	case *tcell.EventMouse:
		x, y := event.Position()
		return MouseMsg{X: x, Y: y, Buttons: event.Buttons(), Modifiers: event.Modifiers()}, nil
	case *tcell.EventResize:
		w, h := event.Size()
		return WindowSizeMsg{Width: w, Height: h}, nil
	case *tcell.EventPaste:
		return PasteMsg{Start: event.Start()}, nil
	case *tcell.EventFocus:
		return FocusMsg{Focused: event.Focused}, nil
	case *tcell.EventError:
		return nil, event
	case *tcell.EventInterrupt:
		return nil, nil
	case nil:
		return nil, errors.New("star: screen event stream closed")
	default:
		return RawEventMsg{Event: event}, nil
	}
}

func render(screen tcell.Screen, view widgets.Widget) error {
	screen.Clear()
	if view != nil {
		w, h := screen.Size()
		view.Measure(widgets.Tight(w, h))
		view.Arrange(widgets.NewRect(0, 0, w, h))
		view.Render(screen)
	}
	screen.Show()
	return nil
}
