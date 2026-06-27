package handlers

import (
	"github.com/jmoiron/sqlx"
	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
)

type Handler struct {
	openapi.UnimplementedHandler
	db *sqlx.DB
}

var _ openapi.Handler = (*Handler)(nil)

func New(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}
