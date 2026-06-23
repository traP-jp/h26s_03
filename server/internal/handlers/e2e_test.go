package handlers_test

import (
	"bytes"
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

	"github.com/traP-jp/h26s_03/server/internal/handlers"
)

func TestAPIEndToEndWithMySQLContainer(t *testing.T) {
	t.Parallel()

	baseURL := startTestServer(t)

	testCases := []struct {
		name string
		run  func(*testing.T, string)
	}{
		{
			name: "initialize seeds expected records",
			run:  scenarioInitializeSeedsExpectedRecords,
		},
		{
			name: "create task adds new feed item",
			run:  scenarioCreateTaskAddsFeedItem,
		},
		{
			name: "create task validates required fields",
			run:  scenarioCreateTaskValidatesRequiredFields,
		},
		{
			name: "create task with unknown member fails",
			run:  scenarioCreateTaskUnknownMemberFails,
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
	h := handlers.New(db)
	e.POST("/api/initialize", h.InitializeEcho)
	e.GET("/api/feed", h.GetFeedEcho)
	e.GET("/api/members", h.GetMembersEcho)
	e.POST("/api/tasks", h.CreateTaskEcho)

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	return srv.URL
}

func scenarioInitializeSeedsExpectedRecords(t *testing.T, baseURL string) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)

	feed := fetchFeed(t, baseURL+"/api/feed")
	if got, want := len(feed.Data), 3; got != want {
		t.Fatalf("unexpected feed length: got=%d want=%d", got, want)
	}

	members := fetchMembers(t, baseURL+"/api/members")
	if got, want := len(members.Data), 3; got != want {
		t.Fatalf("unexpected members length: got=%d want=%d", got, want)
	}
}

func scenarioCreateTaskAddsFeedItem(t *testing.T, baseURL string) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)

	createTaskAndExpectStatus(t, baseURL+"/api/tasks", map[string]any{
		"title":     "Testcontainers task",
		"member_id": 1,
	}, http.StatusCreated)

	feed := fetchFeed(t, baseURL+"/api/feed")
	if got, want := len(feed.Data), 4; got != want {
		t.Fatalf("unexpected feed length after create: got=%d want=%d", got, want)
	}

	var found bool
	for _, item := range feed.Data {
		if item.TaskTitle == "Testcontainers task" {
			found = true
			if item.TaskStatus != "todo" {
				t.Fatalf("unexpected created task status: got=%s", item.TaskStatus)
			}
			break
		}
	}
	if !found {
		t.Fatalf("created task not found in feed")
	}
}

func scenarioCreateTaskValidatesRequiredFields(t *testing.T, baseURL string) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)

	body := createTaskAndExpectStatus(t, baseURL+"/api/tasks", map[string]any{
		"member_id": 1,
	}, http.StatusBadRequest)

	if !strings.Contains(body, "title and member_id are required") {
		t.Fatalf("unexpected validation message: %s", body)
	}
}

func scenarioCreateTaskUnknownMemberFails(t *testing.T, baseURL string) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)

	createTaskAndExpectStatus(t, baseURL+"/api/tasks", map[string]any{
		"title":     "No member",
		"member_id": 99999,
	}, http.StatusInternalServerError)
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

type feedResponse struct {
	Data []struct {
		TaskID     int64  `json:"task_id"`
		TaskTitle  string `json:"task_title"`
		TaskStatus string `json:"task_status"`
	} `json:"data"`
}

func createTaskAndExpectStatus(t *testing.T, url string, payload map[string]any, expectedStatus int) string {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post /api/tasks: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != expectedStatus {
		t.Fatalf("unexpected status from create task: got=%d want=%d body=%s", resp.StatusCode, expectedStatus, string(raw))
	}

	return string(raw)
}

func fetchFeed(t *testing.T, url string) feedResponse {
	t.Helper()

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("get feed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected feed status: %d body=%s", resp.StatusCode, string(raw))
	}

	var out feedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode feed: %v", err)
	}
	return out
}

type membersResponse struct {
	Data []struct {
		ID int64 `json:"id"`
	} `json:"data"`
}

func fetchMembers(t *testing.T, url string) membersResponse {
	t.Helper()

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("get members: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected members status: %d body=%s", resp.StatusCode, string(raw))
	}

	var out membersResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode members: %v", err)
	}
	return out
}
