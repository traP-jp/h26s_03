package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) Initialize(ctx context.Context) error {
	conn, err := h.db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `SET FOREIGN_KEY_CHECKS = 0;`); err != nil {
		return fmt.Errorf("disable foreign key checks: %w", err)
	}
	defer conn.ExecContext(ctx, `SET FOREIGN_KEY_CHECKS = 1;`)

	queries := []string{
		`TRUNCATE TABLE tasks;`,
		`TRUNCATE TABLE projects;`,
		`TRUNCATE TABLE members;`,
		`INSERT INTO members (id, name) VALUES
			(1, 'Sakura'),
			(2, 'Haru'),
			(3, 'Mio');`,
		`INSERT INTO projects (id, name, owner_member_id) VALUES
			(1, 'Landing Page', 1),
			(2, 'Admin API', 2);`,
		`INSERT INTO tasks (id, project_id, assignee_member_id, title, status) VALUES
			(1, 1, 2, 'Design hero section', 'todo'),
			(2, 1, 3, 'Implement CTA animation', 'doing'),
			(3, 2, 1, 'Build initialize endpoint', 'done');`,
	}

	for _, q := range queries {
		if _, err := conn.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("initialize failed: %w", err)
		}
	}

	return nil
}

func (h *Handler) InitializeEcho(c echo.Context) error {
	if err := h.Initialize(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
