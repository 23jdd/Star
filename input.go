package star

import "github.com/23jdd/Star/widgets"

// RouteInput forwards a Star input message into a widget Router. It returns
// false for messages that are not keyboard or mouse input.
func RouteInput(router *widgets.Router, message Message) bool {
	if router == nil {
		return false
	}
	switch message := message.(type) {
	case KeyMsg:
		return router.HandleKey(message.Key, message.Rune, message.Modifiers)
	case MouseMsg:
		return router.HandleMouse(message.X, message.Y, message.Buttons, message.Modifiers)
	default:
		return false
	}
}
