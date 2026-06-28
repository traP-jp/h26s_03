package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
	"github.com/traP-jp/h26s_03/server/internal/handlers"
	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
)

const defaultMigrationsDir = "migrations"

func main() {
	dsn := getenv("DB_DSN", "app:app@tcp(localhost:3306)/app?parseTime=true&multiStatements=true")
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	migrationsDir := getenv("MIGRATIONS_DIR", defaultMigrationsDir)
	if err := runMigrations(migrationsDir, dsn); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

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
	e.GET("/api/ws", h.WebSocket)
	e.GET("/api/polls", h.GetPollsEcho)
	e.PATCH("/api/polls/:id", h.UpdatePollEcho)

	apiServer, err := openapi.NewServer(h)
	if err != nil {
		log.Fatalf("failed to create openapi server: %v", err)
	}
	e.Any("/api/*", echo.WrapHandler(apiServer))

	assetsDir := getenv("ASSETS_DIR", "")
	if assetsDir != "" {
		indexPath := filepath.Join(assetsDir, "index.html")
		e.GET("/*", func(c echo.Context) error {
			requestPath := strings.TrimPrefix(filepath.Clean("/"+c.Param("*")), "/")
			filePath := filepath.Join(assetsDir, requestPath)

			relPath, err := filepath.Rel(assetsDir, filePath)
			if err != nil || relPath == ".." || strings.HasPrefix(relPath, "../") {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid asset path")
			}

			info, err := os.Stat(filePath)
			if err == nil && !info.IsDir() {
				return c.File(filePath)
			}
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			return c.File(indexPath)
		})
		log.Printf("serving assets from %s", assetsDir)
	}

	addr := getenv("API_ADDR", ":8080")
	log.Printf("api listening on %s", addr)
	if err := e.Start(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func runMigrations(migrationsDir, dsn string) error {
	migrationsDir = resolveMigrationsDir(migrationsDir)

	absMigrationsDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return err
	}

	m, err := migrate.New("file://"+absMigrationsDir, "mysql://"+dsn)
	if err != nil {
		return err
	}
	defer func() {
		sourceErr, databaseErr := m.Close()
		if sourceErr != nil {
			log.Printf("failed to close migration source: %v", sourceErr)
		}
		if databaseErr != nil {
			log.Printf("failed to close migration database: %v", databaseErr)
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("database migrations are up to date")
			return nil
		}
		return err
	}

	log.Printf("database migrations applied")
	return nil
}

func resolveMigrationsDir(migrationsDir string) string {
	if migrationsDir != defaultMigrationsDir {
		return migrationsDir
	}
	if _, err := os.Stat(migrationsDir); err == nil {
		return migrationsDir
	}
	if _, err := os.Stat(filepath.Join("server", defaultMigrationsDir)); err == nil {
		return filepath.Join("server", defaultMigrationsDir)
	}
	return migrationsDir
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
