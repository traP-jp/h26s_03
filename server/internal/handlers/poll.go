package handlers

import (
	"context"
	"database/sql"
	"time"

	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
)

func (h *Handler) GetPoll(ctx context.Context, params openapi.GetPollParams) (openapi.GetPollRes, error) {
	var (
		poll      openapi.Poll
		result    sql.NullInt64
		due       sql.NullTime
		createdAt time.Time
	)

	if err := h.db.QueryRowxContext(
		ctx,
		"SELECT id, name, choice1, choice2, result, due, created_by, created_at FROM polls WHERE id = ?",
		params.ID,
	).Scan(
		&poll.ID,
		&poll.Name,
		&poll.Choice1,
		&poll.Choice2,
		&result,
		&due,
		&poll.CreatedBy,
		&createdAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return &openapi.GetPollNotFound{}, nil
		}
		return nil, err
	}

	if result.Valid {
		poll.Result.SetTo(int(result.Int64))
	} else {
		poll.Result.SetToNull()
	}

	if due.Valid {
		poll.Due.SetTo(due.Time)
	} else {
		poll.Due.SetToNull()
	}

	poll.CreatedAt = createdAt

	return &poll, nil
}
