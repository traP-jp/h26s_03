package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
)

func (h *Handler) GetMe(ctx context.Context) (*openapi.Me, error) {
	username, ok := authx.UserFromRequestContext(ctx)
	if !ok {
		username = anonymousUser
	}

	var balance int
	if err := h.db.QueryRowxContext(ctx, `SELECT balance FROM users WHERE username = ?`, username).Scan(&balance); err != nil {
		if err == sql.ErrNoRows {
			if _, err := h.db.ExecContext(ctx, `INSERT INTO users (username, balance) VALUES (?, ?)`, username, initialUserBalance); err != nil {
				return nil, fmt.Errorf("create me: %w", err)
			}
			return &openapi.Me{Username: username, Balance: initialUserBalance}, nil
		}
		return nil, fmt.Errorf("get me: %w", err)
	}

	return &openapi.Me{Username: username, Balance: balance}, nil
}
