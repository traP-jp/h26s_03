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
	Title    string `json:"title"`
	MemberID int64  `json:"member_id"`
}

func (h *Handler) CreateTask(ctx context.Context, req *openapi.CreateTaskRequest) error {
	if req == nil || req.Title == "" || req.MemberID == 0 {
		return fmt.Errorf("title and member_id are required")
	}

	query := `INSERT INTO tasks (project_id, assignee_member_id, title, status) VALUES (?, ?, ?, 'todo')`
	if _, err := h.db.ExecContext(ctx, query, 1, req.MemberID, req.Title); err != nil {
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
		Title:    req.Title,
		MemberID: req.MemberID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "title and member_id are required") {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}
