package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
)

func (h *Handler) GetPollsID(c echo.Context) error {
	pollId := c.Param("id")

	if len(pollId) == 0 {
		return c.String(http.StatusBadRequest, "Poll ID is required")
	}

	var (
		poll      openapi.Poll
		result    sql.NullInt64
		due       sql.NullTime
		createdAt time.Time
	)

	if err := h.db.QueryRowxContext(
		c.Request().Context(),
		"SELECT id, name, choice1, choice2, result, due, created_by, created_at FROM polls WHERE id = ?",
		pollId,
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
			return c.String(http.StatusNotFound, "Poll not found")
		}
		return c.String(http.StatusInternalServerError, "Failed to fetch poll")
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

	return c.JSON(http.StatusOK, poll)
}
