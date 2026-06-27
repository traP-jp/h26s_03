package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
)

const anonymousUser = "anonymous"

func (h *Handler) CreatePoll(ctx context.Context, req *openapi.CreatePollRequest) (*openapi.Poll, error) {
	createdBy, ok := authx.UserFromRequestContext(ctx)
	if !ok {
		createdBy = anonymousUser
	}

	createdAt := time.Now().UTC()
	due := sql.NullTime{}
	if v, ok := req.Due.Get(); ok {
		due = sql.NullTime{Time: v, Valid: true}
	}

	result, err := h.db.ExecContext(ctx, `
		INSERT INTO polls (name, choice1, choice2, result, due, created_by, created_at)
		VALUES (?, ?, ?, NULL, ?, ?, ?)
	`, req.Name, req.Choice1, req.Choice2, due, createdBy, createdAt)
	if err != nil {
		return nil, fmt.Errorf("create poll: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get created poll id: %w", err)
	}

	poll := &openapi.Poll{
		ID:        id,
		Name:      req.Name,
		Choice1:   req.Choice1,
		Choice2:   req.Choice2,
		Result:    nilInt(),
		Due:       nilDateTime(due),
		CreatedBy: createdBy,
		CreatedAt: createdAt,
	}

	return poll, nil
}

func nilInt() openapi.NilInt {
	v := openapi.NilInt{}
	v.SetToNull()
	return v
}

func nilDateTime(t sql.NullTime) openapi.NilDateTime {
	v := openapi.NilDateTime{}
	if t.Valid {
		v.SetTo(t.Time)
		return v
	}
	v.SetToNull()
	return v
}
