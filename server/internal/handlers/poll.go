package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

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

func (h *Handler) GetPollsID(c echo.Context) error {
	pollId := c.Param("id")

	if len(pollId) == 0 {
		return c.String(http.StatusBadRequest, "Poll ID is required")
	}

	var poll pollRow
	if err := h.db.GetContext(c.Request().Context(), &poll, "SELECT * FROM polls WHERE id = ?", pollId); err != nil {
		if err == sql.ErrNoRows {
			return c.String(http.StatusNotFound, "Poll not found")
		}
		return c.String(http.StatusInternalServerError, "Failed to fetch poll")
	}

	return c.JSON(http.StatusOK, poll)
}
