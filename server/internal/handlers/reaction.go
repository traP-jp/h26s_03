package handlers

import "errors"

type reactionWebSocketMessage struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"`
	Reaction string `json:"reaction"`
}

func (m reactionWebSocketMessage) validate() error {
	if m.Type != websocketMessageTypeReaction {
		return errors.New("type must be reaction")
	}
	if m.Reaction == "" {
		return errors.New("reaction is required")
	}
	return nil
}
