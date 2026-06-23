package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
)

type memberRow struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name"`
}

func (h *Handler) GetMembers(ctx context.Context) (*openapi.MembersResponse, error) {
	query := `SELECT id, name FROM members ORDER BY id`
	rows := []memberRow{}
	if err := h.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("select members: %w", err)
	}

	out := make([]openapi.Member, 0, len(rows))
	for _, row := range rows {
		out = append(out, openapi.Member{ID: row.ID, Name: row.Name})
	}

	return &openapi.MembersResponse{Data: out}, nil
}

func (h *Handler) GetMembersEcho(c echo.Context) error {
	res, err := h.GetMembers(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
