package handlers

import "errors"

const (
	websocketMessageTypeReaction = "reaction"
	websocketMessageTypeVote     = "vote"
)

type reactionWebSocketMessage struct {
	Type     string `json:"type"`
	PollID   string `json:"poll_id"`
	Username string `json:"username"`
	Reaction string `json:"reaction"`
}

func (m reactionWebSocketMessage) validate() error {
	if m.Type != websocketMessageTypeReaction {
		return errors.New("type must be reaction")
	}
	if m.PollID == "" {
		return errors.New("poll_id is required")
	}
	if m.Username == "" {
		return errors.New("username is required")
	}
	if m.Reaction == "" {
		return errors.New("reaction is required")
	}
	return nil
}
