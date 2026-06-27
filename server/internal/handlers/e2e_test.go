package handlers_test

import (
	"context"
	"encoding/json"
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
)

func TestAPIEndToEndWithMySQLContainer(t *testing.T) {
	t.Parallel()

	baseURL, db := startTestServer(t)

	testCases := []struct {
		name string
		run  func(*testing.T, string, *sqlx.DB)
	}{
		{
			name: "initialize succeeds",
			run:  scenarioInitializeSucceeds,
		},
		{
			name: "get poll by id returns poll",
			run:  scenarioGetPollByIDReturnsPoll,
		},
		{
			name: "get poll by id returns not found",
			run:  scenarioGetPollByIDReturnsNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.run(t, baseURL, db)
		})
	}
}

func startTestServer(t *testing.T) (string, *sqlx.DB) {
	t.Helper()

	ctx := context.Background()
	dsn := startMySQL(t, ctx)
	db := connectDBWithRetry(t, dsn)
	t.Cleanup(func() { _ = db.Close() })
	applyMigrations(t, db)

	e := echo.New()
	h := handlers.New(db)
	apiServer, err := openapi.NewServer(h)
	if err != nil {
		t.Fatalf("create api server: %v", err)
	}
	e.Any("/api/*", echo.WrapHandler(apiServer))

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	return srv.URL, db
}

func scenarioInitializeSucceeds(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
}

func scenarioGetPollByIDReturnsPoll(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	seedPoll(t, db)

	resp, err := http.Get(baseURL + "/api/polls/1")
	if err != nil {
		t.Fatalf("get poll: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected poll status: got=%d want=%d body=%s", resp.StatusCode, http.StatusOK, string(raw))
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("unexpected poll content-type: got=%s want application/json", contentType)
	}

	var out pollResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode poll: %v", err)
	}

	if out.ID != 1 {
		t.Fatalf("unexpected poll id: got=%d want=1", out.ID)
	}
	if out.Name != "きのこ派？たけのこ派？" {
		t.Fatalf("unexpected poll name: got=%s", out.Name)
	}
	if out.Choice1 != "きのこ" || out.Choice2 != "たけのこ" {
		t.Fatalf("unexpected choices: choice1=%s choice2=%s", out.Choice1, out.Choice2)
	}
	if out.CreatedBy != "traq_user" {
		t.Fatalf("unexpected created_by: got=%s", out.CreatedBy)
	}
	if out.Result != nil {
		t.Fatalf("unexpected result: got=%v want nil", *out.Result)
	}
	if out.Due != nil {
		t.Fatalf("unexpected due: got=%v want nil", *out.Due)
	}
	wantCreatedAt := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	if !out.CreatedAt.Equal(wantCreatedAt) {
		t.Fatalf("unexpected created_at: got=%s want=%s", out.CreatedAt, wantCreatedAt)
	}
}

func scenarioGetPollByIDReturnsNotFound(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)

	resp, err := http.Get(baseURL + "/api/polls/999")
	if err != nil {
		t.Fatalf("get missing poll: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected missing poll status: got=%d want=%d body=%s", resp.StatusCode, http.StatusNotFound, string(raw))
	}
}

type pollResponse struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Choice1   string     `json:"choice1"`
	Choice2   string     `json:"choice2"`
	Result    *int64     `json:"result"`
	Due       *time.Time `json:"due"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
}

func seedPoll(t *testing.T, db *sqlx.DB) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO polls (id, name, choice1, choice2, result, due, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		1,
		"きのこ派？たけのこ派？",
		"きのこ",
		"たけのこ",
		nil,
		nil,
		"traq_user",
		time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("seed poll: %v", err)
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
