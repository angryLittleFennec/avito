package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("token")

		if isValidUserToken(token) {
			return echo.NewHTTPError(http.StatusForbidden, "No access")
		}

		if !isValidAdminToken(token) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		return next(c)
	}
}

func UserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("token")

		if !isValidUserToken(token) && !isValidAdminToken(token) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		return next(c)
	}
}

var validAdminTokens = []string{"admin1", "admin2", "admin3"}

var validUserTokens = []string{"user1", "user2", "user3"}

func isValidAdminToken(token string) bool {
	for _, t := range validAdminTokens {
		if t == token {
			return true
		}
	}
	return false
}

func isValidUserToken(token string) bool {
	for _, t := range validUserTokens {
		if t == token {
			return true
		}
	}
	return false
}
