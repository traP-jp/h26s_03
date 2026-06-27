package authx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSoftWithForwardedUserStoresUserInRequestContext(t *testing.T) {
	t.Parallel()

	called := false
	err := runMiddleware(t, Soft(), HeaderForwardedUser, " alice ", func(c echo.Context) error {
		called = true

		user, ok := UserFromRequestContext(c.Request().Context())
		if !ok {
			t.Fatalf("expected user in request context")
		}
		if user != "alice" {
			t.Fatalf("unexpected user: got=%s want=alice", user)
		}

		return c.NoContent(http.StatusNoContent)
	})
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	if !called {
		t.Fatalf("expected next handler to be called")
	}
}

func TestSoftWithoutForwardedUserPassesWithoutUser(t *testing.T) {
	t.Parallel()

	called := false
	err := runMiddleware(t, Soft(), "", "", func(c echo.Context) error {
		called = true

		if user, ok := UserFromRequestContext(c.Request().Context()); ok {
			t.Fatalf("unexpected user: %s", user)
		}

		return c.NoContent(http.StatusNoContent)
	})
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	if !called {
		t.Fatalf("expected next handler to be called")
	}
}

func TestHardWithoutForwardedUserReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	called := false
	err := runMiddleware(t, Hard(), "", "", func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})

	var httpErr *echo.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected echo.HTTPError, got=%T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got=%d want=%d", httpErr.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatalf("expected next handler not to be called")
	}
}

func runMiddleware(t *testing.T, middleware echo.MiddlewareFunc, headerName, headerValue string, next echo.HandlerFunc) error {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if headerName != "" {
		req.Header.Set(headerName, headerValue)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return middleware(next)(c)
}
