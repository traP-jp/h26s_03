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
	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
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
		{
			name: "get poll votes returns votes",
			run:  scenarioGetPollVotesReturnsVotes,
		},
		{
			name: "get poll votes returns not found",
			run:  scenarioGetPollVotesReturnsNotFound,
		},
		{
			name: "create poll succeeds",
			run:  scenarioCreatePollSucceeds,
		},
		{
			name: "create vote succeeds",
			run:  scenarioCreateVoteSucceeds,
		},
		{
			name: "create vote returns not found",
			run:  scenarioCreateVoteReturnsNotFound,
		},
		{
			name: "create vote returns conflict",
			run:  scenarioCreateVoteReturnsConflict,
		},
		{
			name: "create vote returns conflict when balance is insufficient",
			run:  scenarioCreateVoteReturnsConflictWhenBalanceIsInsufficient,
		},
		{
			name: "delete poll succeeds",
			run:  scenarioDeletePollSucceeds,
		},
		{
			name: "delete poll returns forbidden",
			run:  scenarioDeletePollReturnsForbidden,
		},
		{
			name: "delete poll returns not found",
			run:  scenarioDeletePollReturnsNotFound,
		},
		{
			name: "patch poll updates selected fields",
			run:  scenarioPatchPollUpdatesSelectedFields,
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
	e.Use(authx.Soft())
	h := handlers.New(db)
	e.PATCH("/api/polls/:id", h.UpdatePollEcho)
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

func scenarioGetPollVotesReturnsVotes(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	seedPoll(t, db)
	seedVotes(t, db)

	resp, err := http.Get(baseURL + "/api/polls/1/votes")
	if err != nil {
		t.Fatalf("get poll votes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected poll votes status: got=%d want=%d body=%s", resp.StatusCode, http.StatusOK, string(raw))
	}

	var out votesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode poll votes: %v", err)
	}
	if len(out.Data) != 2 {
		t.Fatalf("unexpected votes length: got=%d want=2", len(out.Data))
	}
	if out.Data[0].ID != 1 || out.Data[0].Username != "alice" || out.Data[0].Choice != 1 || out.Data[0].Bet != 100 {
		t.Fatalf("unexpected first vote: %+v", out.Data[0])
	}
	if out.Data[1].ID != 2 || out.Data[1].Username != "bob" || out.Data[1].Choice != 2 || out.Data[1].Bet != 200 {
		t.Fatalf("unexpected second vote: %+v", out.Data[1])
	}
}

func scenarioGetPollVotesReturnsNotFound(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	mustRequestNoBody(t, http.MethodGet, baseURL+"/api/polls/999/votes", http.StatusNotFound)
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

type votesResponse struct {
	Data []voteResponse `json:"data"`
}

type voteResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Choice    int       `json:"choice"`
	Bet       int       `json:"bet"`
	CreatedAt time.Time `json:"created_at"`
}

func seedPoll(t *testing.T, db *sqlx.DB) {
	t.Helper()

	seedPollCreatedBy(t, db, "traq_user")
}

func seedPollCreatedBy(t *testing.T, db *sqlx.DB, createdBy string) int64 {
	t.Helper()

	_, err := db.Exec(`INSERT INTO polls (id, name, choice1, choice2, result, due, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		1,
		"きのこ派？たけのこ派？",
		"きのこ",
		"たけのこ",
		nil,
		nil,
		createdBy,
		time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("seed poll: %v", err)
	}

	return 1
}

func seedUser(t *testing.T, db *sqlx.DB, username string, balance int) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO users (username, balance) VALUES (?, ?)`, username, balance)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func userBalance(t *testing.T, db *sqlx.DB, username string) int {
	t.Helper()

	var balance int
	if err := db.QueryRow(`SELECT balance FROM users WHERE username = ?`, username).Scan(&balance); err != nil {
		t.Fatalf("get user balance: %v", err)
	}
	return balance
}

func scenarioPatchPollUpdatesSelectedFields(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	pollID := seedPollCreatedBy(t, db, "owner-user")

	body := strings.NewReader(`{"name":"after","result":1,"due":null}`)
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/polls/%d", baseURL, pollID), body)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(authx.HeaderForwardedUser, "owner-user")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request PATCH /api/polls/%d: %v", pollID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: got=%d want=%d body=%s", resp.StatusCode, http.StatusOK, string(raw))
	}

	var got struct {
		ID     int64   `json:"id"`
		Name   string  `json:"name"`
		Result *int    `json:"result"`
		Due    *string `json:"due"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != pollID {
		t.Fatalf("unexpected id: got=%d want=%d", got.ID, pollID)
	}
	if got.Name != "after" {
		t.Fatalf("unexpected name: got=%q want=%q", got.Name, "after")
	}
	if got.Result == nil || *got.Result != 1 {
		t.Fatalf("unexpected result: got=%v want=1", got.Result)
	}
	if got.Due != nil {
		t.Fatalf("unexpected due: got=%v want=nil", got.Due)
	}
}

func seedVotes(t *testing.T, db *sqlx.DB) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO votes (id, poll_id, username, choice, bet, created_at)
		VALUES (?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?)`,
		1,
		1,
		"alice",
		1,
		100,
		time.Date(2026, 6, 27, 13, 0, 0, 0, time.UTC),
		2,
		1,
		"bob",
		2,
		200,
		time.Date(2026, 6, 27, 14, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("seed votes: %v", err)
	}
}

func scenarioCreatePollSucceeds(t *testing.T, baseURL string, db *sqlx.DB) {
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

func scenarioCreateVoteSucceeds(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	seedPoll(t, db)
	seedUser(t, db, "alice", 100)

	resp := mustRequestJSONWithUser(t, http.MethodPost, baseURL+"/api/polls/1/votes", "alice", `{"choice":1,"bet":100}`, http.StatusCreated)

	var out voteResponse
	if err := json.Unmarshal(resp, &out); err != nil {
		t.Fatalf("decode created vote: %v", err)
	}
	if out.ID != 1 || out.Username != "alice" || out.Choice != 1 || out.Bet != 100 {
		t.Fatalf("unexpected created vote: %+v", out)
	}
	if out.CreatedAt.IsZero() {
		t.Fatalf("unexpected created_at: zero")
	}
	if got := userBalance(t, db, "alice"); got != 0 {
		t.Fatalf("unexpected balance: got=%d want=0", got)
	}
}

func scenarioCreateVoteReturnsNotFound(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	mustRequestJSONWithUser(t, http.MethodPost, baseURL+"/api/polls/999/votes", "alice", `{"choice":1,"bet":100}`, http.StatusNotFound)
}

func scenarioCreateVoteReturnsConflict(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	seedPoll(t, db)
	seedUser(t, db, "alice", 300)
	mustRequestJSONWithUser(t, http.MethodPost, baseURL+"/api/polls/1/votes", "alice", `{"choice":1,"bet":100}`, http.StatusCreated)
	mustRequestJSONWithUser(t, http.MethodPost, baseURL+"/api/polls/1/votes", "alice", `{"choice":2,"bet":200}`, http.StatusConflict)
	if got := userBalance(t, db, "alice"); got != 200 {
		t.Fatalf("unexpected balance: got=%d want=200", got)
	}
}

func scenarioCreateVoteReturnsConflictWhenBalanceIsInsufficient(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	seedPoll(t, db)
	seedUser(t, db, "alice", 99)
	mustRequestJSONWithUser(t, http.MethodPost, baseURL+"/api/polls/1/votes", "alice", `{"choice":1,"bet":100}`, http.StatusConflict)
	if got := userBalance(t, db, "alice"); got != 99 {
		t.Fatalf("unexpected balance: got=%d want=99", got)
	}
}

func scenarioDeletePollSucceeds(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	seedPoll(t, db)
	mustRequestNoBodyWithUser(t, http.MethodDelete, baseURL+"/api/polls/1", "traq_user", http.StatusNoContent)
	mustRequestNoBody(t, http.MethodGet, baseURL+"/api/polls/1", http.StatusNotFound)
}

func scenarioDeletePollReturnsForbidden(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	seedPoll(t, db)
	mustRequestNoBodyWithUser(t, http.MethodDelete, baseURL+"/api/polls/1", "alice", http.StatusForbidden)
	mustRequestNoBody(t, http.MethodGet, baseURL+"/api/polls/1", http.StatusOK)
}

func scenarioDeletePollReturnsNotFound(t *testing.T, baseURL string, db *sqlx.DB) {
	t.Helper()

	mustRequestNoBody(t, http.MethodPost, baseURL+"/api/initialize", http.StatusNoContent)
	mustRequestNoBodyWithUser(t, http.MethodDelete, baseURL+"/api/polls/999", "traq_user", http.StatusNotFound)
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

func mustRequestNoBodyWithUser(t *testing.T, method, url, user string, expectedStatus int) {
	t.Helper()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("X-Forwarded-User", user)

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

func mustRequestJSONWithUser(t *testing.T, method, url, user, body string, expectedStatus int) []byte {
	t.Helper()

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-User", user)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, url, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != expectedStatus {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", resp.StatusCode, expectedStatus, string(raw))
	}
	return raw
}
