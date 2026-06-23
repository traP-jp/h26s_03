package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
)

type taskRow struct {
	ID     int64  `db:"id" json:"id"`
	Title  string `db:"title" json:"title"`
	Status string `db:"status" json:"status"`
}

func (h *Handler) GetTasks(ctx context.Context) (*openapi.TasksResponse, error) {
	query := `SELECT id, title, status FROM tasks ORDER BY id`

	rows := []taskRow{}
	if err := h.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}

	out := make([]openapi.Task, 0, len(rows))
	for _, row := range rows {
		out = append(out, openapi.Task{
			ID:     row.ID,
			Title:  row.Title,
			Status: row.Status,
		})
	}

	return &openapi.TasksResponse{Data: out}, nil
}

func (h *Handler) GetTasksEcho(c echo.Context) error {
	res, err := h.GetTasks(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
