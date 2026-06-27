package authx

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	HeaderForwardedUser = "X-Forwarded-User"
	ContextUserKey      = "forwardedUser"
)

type contextUserKey struct{}

type Mode string

const (
	ModeSoft Mode = "SOFT"
	ModeHard Mode = "HARD"
)

func ParseMode(raw string) Mode {
	if strings.EqualFold(raw, string(ModeHard)) {
		return ModeHard
	}
	return ModeSoft
}

func Soft() echo.MiddlewareFunc {
	return New(ModeSoft)
}

func Hard() echo.MiddlewareFunc {
	return New(ModeHard)
}

func New(mode Mode) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := strings.TrimSpace(c.Request().Header.Get(HeaderForwardedUser))
			if user == "" {
				if mode == ModeHard {
					return echo.NewHTTPError(http.StatusUnauthorized, HeaderForwardedUser+" is required")
				}
				return next(c)
			}

			c.Set(ContextUserKey, user)
			req := c.Request()
			c.SetRequest(req.WithContext(context.WithValue(req.Context(), contextUserKey{}, user)))
			return next(c)
		}
	}
}

func UserFromRequestContext(ctx context.Context) (string, bool) {
	v := ctx.Value(contextUserKey{})
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", false
	}
	return s, true
}
