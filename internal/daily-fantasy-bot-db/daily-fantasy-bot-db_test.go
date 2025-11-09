package dailyfantasybotdb

import (
	"context"
	"testing"
	"time"
	"github.com/joho/godotenv"
	// "github.com/arjunmoola/nba-daily-fantasy-bot/internal/utils"
	"os"
	// "path"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

//var _ = loadEnv()

//var _ = godotenv.Load("./../../.env")

var pool *pgxpool.Pool

func beginTx(t *testing.T) pgx.Tx {
	tx, err := pool.Begin(t.Context())

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	return tx
}

func cleanup(t *testing.T, tx pgx.Tx) {
	t.Cleanup(func() {
		tx.Rollback(t.Context())
	})
}

func TestMain(m *testing.M) {
	godotenv.Load("./../../.env")

	dbUrl := os.Getenv("DATABASE_URL")

	config, err := pgxpool.ParseConfig(dbUrl)

	if err != nil {
		os.Exit(1)
	}

	ctx := context.Background()

	p, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		os.Exit(1)
	}

	if err := p.Ping(ctx); err != nil {
		p.Close()
		os.Exit(1)
	}

	pool = p

	code := m.Run()
	p.Close()

	os.Exit(code)
}

func TestPing(t *testing.T) {
	tx := beginTx(t)

	cleanup(t, tx)

}

func TestGetLockTime(t *testing.T) {
	tx := beginTx(t)

	cleanup(t, tx)

	lockTime, err := getLockTime(t.Context(), tx, time.Now())

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Log(lockTime)
}
