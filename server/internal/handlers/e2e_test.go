package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
	"github.com/traP-jp/h26s_03/server/internal/handlers"
	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
)

func TestAPIEndToEndWithMySQLContainer(t *testing.T) {
	t.Parallel()

	baseURL := startTestServer(t)

	testCases := []struct {
		name string
		run  func(*testing.T, string)
	}{
		{
			name: "initialize succeeds",
			run:  scenarioInitializeSucceeds,
		},
		{
			name: "create poll succeeds",
			run:  scenarioCreatePollSucceeds,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.run(t, baseURL)
		})
	}
}

func startTestServer(t *testing.T) string {
	t.Helper()

	ctx := context.Background()
	dsn := startMySQL(t, ctx)
	db := connectDBWithRetry(t, dsn)
	t.Cleanup(func() { _ = db.Close() })
	applyMigrations(t, db)

	e := echo.New()
	e.Use(authx.Soft())
	h := handlers.New(db)
	e.POST("/api/initialize", h.InitializeEcho)
	apiServer, err := openapi.NewServer(h)
	if err != nil {
		t.Fatalf("create openapi server: %v", err)
	}
	e.Any("/api/*", echo.WrapHandler(apiServer))

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	return srv.URL
}

func scenarioInitializeSucceeds(t *testing.T, baseURL string) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
}

func scenarioCreatePollSucceeds(t *testing.T, baseURL string) {
	t.Helper()

	body := `{"name":"昼食","choice1":"うどん","choice2":"カレー","due":null}`
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/polls", strings.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-User", "alice")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request POST /api/polls: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", resp.StatusCode, http.StatusCreated, string(raw))
	}

	if !strings.Contains(string(raw), `"created_by":"alice"`) {
		t.Fatalf("unexpected body: %s", string(raw))
	}
}

func startMySQL(t *testing.T, ctx context.Context) string {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.4",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "root",
			"MYSQL_DATABASE":      "app",
			"MYSQL_USER":          "app",
			"MYSQL_PASSWORD":      "app",
		},
		WaitingFor: wait.ForListeningPort("3306/tcp").WithStartupTimeout(2 * time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start mysql container: %v", err)
	}

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("mysql container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "3306/tcp")
	if err != nil {
		t.Fatalf("mysql container port: %v", err)
	}

	return fmt.Sprintf("app:app@tcp(%s:%s)/app?parseTime=true&multiStatements=true", host, port.Port())
}

func connectDBWithRetry(t *testing.T, dsn string) *sqlx.DB {
	t.Helper()

	const attempts = 60
	var lastErr error
	for i := 0; i < attempts; i++ {
		db, err := sqlx.Connect("mysql", dsn)
		if err == nil {
			return db
		}
		lastErr = err
		time.Sleep(1 * time.Second)
	}

	t.Fatalf("connect db: %v", lastErr)
	return nil
}

func applyMigrations(t *testing.T, db *sqlx.DB) {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("get current filename")
	}

	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	migrationPaths, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		t.Fatalf("find migrations: %v", err)
	}
	if len(migrationPaths) == 0 {
		t.Fatalf("no migration files found in %s", migrationsDir)
	}
	sort.Strings(migrationPaths)

	for _, migrationPath := range migrationPaths {
		raw, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read migration %s: %v", filepath.Base(migrationPath), err)
		}
		if _, err := db.Exec(string(raw)); err != nil {
			t.Fatalf("apply migration %s: %v", filepath.Base(migrationPath), err)
		}
	}
}

func mustRequestNoBody(t *testing.T, method, url string, expectedStatus int) {
	t.Helper()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got=%d want=%d body=%s", resp.StatusCode, expectedStatus, string(raw))
	}
}
