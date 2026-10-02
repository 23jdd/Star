package star

import "github.com/23jdd/Star/widgets"

// Message is a value delivered to Model.Update. Messages are processed one at
// a time, so model updates never run concurrently.
type Message = any

// Cmd performs asynchronous work and returns its result as a message. Returning
// nil means that no message should be delivered.
type Cmd func() Message

// Model is the Elm architecture used by a Star application.
type Model interface {
	Init() Cmd
	Update(Message) (Model, Cmd)
	View() widgets.Widget
}
