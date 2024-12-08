package datasource

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

// TODO drop in a config
func NewPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	logrus.Info("initializing database ...")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logrus.Fatalf("DATABASE_URL environment variable is not set")
	}

	var dbPool *pgxpool.Pool
	var err error
	maxRetries := 10
	initialBackoff := time.Second * 1

	for retries := 0; retries < maxRetries; retries++ {
		config, configErr := pgxpool.ParseConfig(dbURL)
		if configErr != nil {
			logrus.Errorf("error parsing database URL: %v", configErr)
			return nil, configErr
		}

		//TODO config
		config.MaxConns = 50
		config.MaxConnIdleTime = 5 * time.Minute

		dbPool, err = pgxpool.ConnectConfig(ctx, config)
		if err == nil {
			logrus.Info("successfully connected to PostgreSQL")
			return dbPool, nil
		}

		backoff := initialBackoff * time.Duration(1<<retries)
		logrus.Warnf("failed to connect to the database, retrying in %v... (%d/%d)", backoff, retries+1, maxRetries)
		time.Sleep(backoff)
	}

	return nil, err
}
