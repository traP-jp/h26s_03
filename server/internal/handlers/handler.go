package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
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

func toOpenAPIPoll(row pollRow) openapi.Poll {
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

	return poll
}

func (h *Handler) getPollByID(ctx context.Context, id int64) (pollRow, error) {
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
WHERE id = ?`

	var row pollRow
	if err := h.db.GetContext(ctx, &row, query, id); err != nil {
		return pollRow{}, err
	}

	return row, nil
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
		polls = append(polls, toOpenAPIPoll(row))
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

func (h *Handler) UpdatePoll(ctx context.Context, req *openapi.UpdatePollRequest, params openapi.UpdatePollParams) (openapi.UpdatePollRes, error) {
	current, err := h.getPollByID(ctx, params.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &openapi.UpdatePollNotFound{}, nil
		}
		return nil, err
	}

	user, ok := authx.UserFromRequestContext(ctx)
	if !ok || user != current.CreatedBy {
		return &openapi.UpdatePollForbidden{}, nil
	}

	name := current.Name
	if v, ok := req.Name.Get(); ok {
		name = v
	}

	choice1 := current.Choice1
	if v, ok := req.Choice1.Get(); ok {
		choice1 = v
	}

	choice2 := current.Choice2
	if v, ok := req.Choice2.Get(); ok {
		choice2 = v
	}

	result := current.Result
	if req.Result.IsSet() {
		result = sql.NullInt64{Int64: int64(req.Result.Value), Valid: true}
	}
	if req.Result.IsNull() {
		result = sql.NullInt64{Valid: false}
	}

	due := current.Due
	if req.Due.IsSet() {
		due = sql.NullTime{Time: req.Due.Value, Valid: true}
	}
	if req.Due.IsNull() {
		due = sql.NullTime{Valid: false}
	}

	const updateQuery = `
UPDATE polls
SET name = ?, choice1 = ?, choice2 = ?, result = ?, due = ?
WHERE id = ?`

	if _, err := h.db.ExecContext(ctx, updateQuery, name, choice1, choice2, result, due, params.ID); err != nil {
		return nil, err
	}

	updated, err := h.getPollByID(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	poll := toOpenAPIPoll(updated)
	return &poll, nil
}

func (h *Handler) UpdatePollEcho(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid poll id")
	}

	var req openapi.UpdatePollRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	response, err := h.UpdatePoll(c.Request().Context(), &req, openapi.UpdatePollParams{ID: id})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	switch response := response.(type) {
	case *openapi.Poll:
		return c.JSON(http.StatusOK, response)
	case *openapi.UpdatePollForbidden:
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	case *openapi.UpdatePollNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "poll not found")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
