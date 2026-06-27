package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/traP-jp/h26s_03/server/internal/handlers"
	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
)

func main() {
	dsn := getenv("DB_DSN", "app:app@tcp(localhost:3306)/app?parseTime=true&multiStatements=true")
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType, authx.HeaderForwardedUser},
		AllowOrigins: []string{"*"},
	}))
	e.Use(authx.New(authx.ParseMode(getenv("AUTH_MODE", string(authx.ModeSoft)))))

	h := handlers.New(db)
	e.POST("/api/initialize", h.InitializeEcho)
	e.GET("/api/polls", h.GetPollsEcho)
	e.GET("/polls", h.GetPollsChapterEcho)

	assetsDir := getenv("ASSETS_DIR", "")
	if assetsDir != "" {
		indexPath := filepath.Join(assetsDir, "index.html")
		e.Static("/", assetsDir)
		e.File("/*", indexPath)
		log.Printf("serving assets from %s", assetsDir)
	}

	addr := getenv("API_ADDR", ":8080")
	log.Printf("api listening on %s", addr)
	if err := e.Start(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
