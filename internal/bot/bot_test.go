package bot

import (
	"os"
	"github.com/arjunmoola/nba-daily-fantasy-bot/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	//"github.com/jackc/pgx/v5"
	"testing"
	"context"
	//"time"
	"github.com/joho/godotenv"
)

var testPool *pgxpool.Pool

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

	testPool = p

	code := m.Run()
	p.Close()

	os.Exit(code)
}

func TestBotInit(t *testing.T) {
	config := config.New()
	bot := New( config, WithPool(testPool))

	if err := bot.Init(t.Context()); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if bot.lockTime.IsZero() {
		t.Errorf("expected lock time to be non zero")
		t.FailNow()
	}

	if len(bot.todaysPlayers) == 0 {
		t.Errorf("expected todays players to be non zero")
		t.FailNow()
	}
}
