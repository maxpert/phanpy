package core

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"os"
)

var connectionPool *pgxpool.Pool = nil

func InitDB(ctx context.Context, logger *zap.SugaredLogger) error {
	connectionString := os.Getenv("DB_URL")
	pool, err := pgxpool.Connect(ctx, connectionString)
	if err != nil {
		return err
	}

	logger.Info("pgx connection initialized...")
	connectionPool = pool

	return nil
}

func CloseDB() {
	if connectionPool != nil {
		connectionPool.Close()
	}
}
