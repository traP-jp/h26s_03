package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
)

type createTaskRequest struct {
	Title string `json:"title"`
}

func (h *Handler) CreateTask(ctx context.Context, req *openapi.CreateTaskRequest) error {
	if req == nil || req.Title == "" {
		return fmt.Errorf("title is required")
	}

	query := `INSERT INTO tasks (title, status) VALUES (?, 'todo')`
	if _, err := h.db.ExecContext(ctx, query, req.Title); err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	return nil
}

func (h *Handler) CreateTaskEcho(c echo.Context) error {
	var req createTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json")
	}

	err := h.CreateTask(c.Request().Context(), &openapi.CreateTaskRequest{
		Title: req.Title,
	})
	if err != nil {
		if strings.Contains(err.Error(), "title is required") {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}
