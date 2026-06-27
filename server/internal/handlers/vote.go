package handlers

import "errors"

type voteWebSocketMessage struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"`
}

func (m voteWebSocketMessage) validate() error {
	if m.Type != websocketMessageTypeVote {
		return errors.New("type must be vote")
	}
	return nil
}
