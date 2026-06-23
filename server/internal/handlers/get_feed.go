package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
)

type feedRow struct {
	ProjectID   int64  `db:"project_id" json:"project_id"`
	ProjectName string `db:"project_name" json:"project_name"`
	OwnerName   string `db:"owner_name" json:"owner_name"`
	TaskID      int64  `db:"task_id" json:"task_id"`
	TaskTitle   string `db:"task_title" json:"task_title"`
	TaskStatus  string `db:"task_status" json:"task_status"`
}

func (h *Handler) GetFeed(ctx context.Context) (*openapi.FeedResponse, error) {
	query := `
		SELECT
			p.id AS project_id,
			p.name AS project_name,
			owner.name AS owner_name,
			t.id AS task_id,
			t.title AS task_title,
			t.status AS task_status
		FROM tasks t
		INNER JOIN projects p ON p.id = t.project_id
		INNER JOIN members owner ON owner.id = p.owner_member_id
		INNER JOIN members assignee ON assignee.id = t.assignee_member_id
		ORDER BY p.id, t.id
	`

	rows := []feedRow{}
	if err := h.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("select feed: %w", err)
	}

	out := make([]openapi.FeedItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, openapi.FeedItem{
			ProjectID:   row.ProjectID,
			ProjectName: row.ProjectName,
			OwnerName:   row.OwnerName,
			TaskID:      row.TaskID,
			TaskTitle:   row.TaskTitle,
			TaskStatus:  row.TaskStatus,
		})
	}

	return &openapi.FeedResponse{Data: out}, nil
}

func (h *Handler) GetFeedEcho(c echo.Context) error {
	res, err := h.GetFeed(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
