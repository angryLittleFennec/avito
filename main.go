package main

import (
	sv "avito/internal/server"
	"github.com/labstack/echo/v4"
	"os"
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

	e.Start(":8080")
}
