package handlers

import "errors"

type voteWebSocketMessage struct {
	Type     string `json:"type"`
	PollID   string `json:"poll_id"`
	Username string `json:"username"`
}


func (m voteWebSocketMessage) validate() error {
	if m.Type != websocketMessageTypeVote {
		return errors.New("type must be vote")
	}
	if m.PollID == "" {
		return errors.New("poll_id is required")
	}
	if m.Username == "" {
		return errors.New("username is required")
	}
	return nil
}
