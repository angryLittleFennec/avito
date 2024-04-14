package server

import (
	"avito/internal/db"
	"log/slog"
	"os"

	"fmt"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Logger *slog.Logger
}

func NewServer(dbUrl string, redisUrl string) (*Server, error) {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	logger := slog.New(handler)

	database, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	err = db.Migrate(database)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})

	return &Server{DB: database, Redis: rdb, Logger: logger}, nil
}
