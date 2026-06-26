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
		`INSERT INTO tasks (id, title, status) VALUES
			(1, 'トップページの構成を考える', 'todo'),
			(2, 'タスク追加フォームを作る', 'doing'),
			(3, '初期化APIをつなぐ', 'done');`,
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
