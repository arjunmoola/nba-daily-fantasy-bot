package dailyfantasybotdb

import (
	"github.com/arjunmoola/nba-daily-fantasy-bot/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"fmt"

	//"database/sql"

	"time"
	"context"
	"os"
	"log/slog"
)

type Tx struct {
	db *DB
}

type DBOption func(d *DB)

func WithLogger(l *slog.Logger) DBOption {
	return func(d *DB) {
		d.logger = l
	}
}

func WithPool(pool *pgxpool.Pool) DBOption {
	return func(d *DB) {
		d.pool = pool
	}
}

func WithDbUrl(s string) DBOption {
	return func(d *DB) {
		d.dbUrl = s
	}
}

type DB struct {
	dbUrl string
	pool *pgxpool.Pool
	logger *slog.Logger

	Tx *Tx
}

func New(ctx context.Context, opts ...DBOption) (*DB, error) {
	db := new(DB)

	db.Tx = &Tx{db: db}

	for _, opt := range opts {
		opt(db)
	}

	if db.pool != nil {
		return db, nil
	}

	if db.dbUrl == "" {
		db.dbUrl = os.Getenv("DATABASE_URL")
	}

	config, err := pgxpool.ParseConfig(db.dbUrl)

	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	db.pool = pool

	return db, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) BeginFunc(ctx context.Context, f func(tx pgx.Tx) error) (err error) {
	tx, err := db.pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			if err != nil {
				err = fmt.Errorf("encontered error during rollback %v: %v", rollbackErr, err)
			} else {
				err = rollbackErr
			}
		}
	}()

	return f(tx)
}

func (t *Tx) GetLockTime(ctx context.Context, date time.Time) (time.Time, error) {
	pgxTx, err := t.db.pool.Begin(ctx)

	if err != nil {
		return time.Time{}, err
	}

	defer pgxTx.Rollback(ctx)

	return getLockTime(ctx, pgxTx, date)
}

func (db *DB) GetLockTime(ctx context.Context, date time.Time) (time.Time, error) {
	return time.Now(), nil
}

func (d *DB) Get(ctx context.Context) {
	database.Queries
}


func getLockTime(ctx context.Context, db database.DBTX, date time.Time) (time.Time, error) {
	params := pgtype.Date{
		Time:date,
		InfinityModifier: pgtype.Finite,
		Valid: true,
	}

	result, err := database.New(db).IsLocked(ctx, params)

	if err != nil {
		return result.LockTime.Time, err
	}

	return result.LockTime.Time, nil
}
