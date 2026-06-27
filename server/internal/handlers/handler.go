package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
)

type Handler struct {
	openapi.UnimplementedHandler
	db    *sqlx.DB
	wsHub *websocketHub
}

var _ openapi.Handler = (*Handler)(nil)

func New(db *sqlx.DB) *Handler {
	return &Handler{db: db, wsHub: newWebsocketHub()}
}

type pollRow struct {
	ID        int64         `db:"id"`
	Name      string        `db:"name"`
	Choice1   string        `db:"choice1"`
	Choice2   string        `db:"choice2"`
	Result    sql.NullInt64 `db:"result"`
	Due       sql.NullTime  `db:"due"`
	CreatedBy string        `db:"created_by"`
	CreatedAt time.Time     `db:"created_at"`
}

func (h *Handler) GetPolls(ctx context.Context) (*openapi.PollsResponse, error) {
	const query = `
SELECT
	id,
	name,
	choice1,
	choice2,
	result,
	due,
	created_by,
	created_at
FROM polls
ORDER BY created_at DESC, id DESC`

	var rows []pollRow
	if err := h.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("select polls: %w", err)
	}

	polls := make([]openapi.Poll, 0, len(rows))
	for _, row := range rows {
		poll := openapi.Poll{
			ID:        row.ID,
			Name:      row.Name,
			Choice1:   row.Choice1,
			Choice2:   row.Choice2,
			CreatedBy: row.CreatedBy,
		}

		if row.Result.Valid {
			poll.Result.SetTo(int(row.Result.Int64))
		} else {
			poll.Result.SetToNull()
		}

		if row.Due.Valid {
			poll.Due.SetTo(row.Due.Time)
		} else {
			poll.Due.SetToNull()
		}

		poll.CreatedAt = row.CreatedAt

		polls = append(polls, poll)
	}

	return &openapi.PollsResponse{Data: polls}, nil
}

func (h *Handler) GetPollsEcho(c echo.Context) error {
	response, err := h.GetPolls(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, response)
}
