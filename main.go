package main

import (
	sv "avito/internal/server"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func main() {

	dbUrl := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	server, err := sv.NewServer(dbUrl, redisURL)

	if err != nil {
		return
	}

	e := echo.New()

	sv.RegisterHandlersWithAuth(e, server)
	e.Use()

	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal("Shutting down the server", err)
	}
}
