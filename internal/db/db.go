package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
)

type T struct {
	*pgxpool.Pool
}

type Entity struct {
	UserID  string `json:"user_id"`
	ShortID string `json:"id"`
	LongURL string `json:"url"`
}

type BatchInput []BatchInputItem
type BatchInputItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	ShortID       string `json:"-"`
}

var ErrUniqueViolation = errors.New("long URL already exist")

func New(ctx context.Context, url string) (T, error) {
	var pool T
	var err error
	pool.Pool, err = pgxpool.Connect(ctx, url)
	if err != nil {
		return pool, err
	}

	// создание таблицы
	sql1 := "create table if not exists urls (" +
		"id serial primary key, " +
		"user_id varchar(512) not null, " +
		"short_id varchar(512) not null, " +
		"long_url varchar(1024) not null unique)"
	_, err = pool.Exec(ctx, sql1)
	if err != nil {
		return pool, err
	}

	return pool, nil
}

func (d *T) AddEntity(ctx context.Context, e Entity) error {
	sql := "insert into urls values (default, $1, $2, $3)"
	_, err := d.Pool.Exec(ctx, sql, e.UserID, e.ShortID, e.LongURL)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrUniqueViolation
		}
	}
	return err
}

func (d *T) SelectByLongURL(ctx context.Context, longURL string) (string, error) {
	row := d.Pool.QueryRow(ctx, "select short_id from urls where long_url = $1", longURL)
	var shortPath string
	err := row.Scan(&shortPath)
	return shortPath, err
}

func (d *T) SelectByShortID(ctx context.Context, id string) (string, error) {
	rows, err := d.Pool.Query(ctx, "select long_url from urls where short_id = $1", id)
	if err != nil {
		return "", err
	}
	longURL := ""
	for rows.Next() {
		err = rows.Scan(&longURL)
		if err != nil {
			return "", err
		}
	}
	return longURL, nil
}

func (d *T) SelectByUser(ctx context.Context, userID string) ([]Entity, error) {
	rows, err := d.Pool.Query(ctx, "select * from urls where user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	var e Entity
	var rowID int
	eArray := make([]Entity, 0, 10)
	for rows.Next() {
		err = rows.Scan(&rowID, &e.UserID, &e.ShortID, &e.LongURL)
		if err != nil {
			return nil, err
		}
		eArray = append(eArray, e)
	}
	return eArray, nil
}

func (d *T) Flush(ctx context.Context, userID string, data BatchInput) error {
	tx, err := d.Begin(ctx)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(ctx, "batch", "insert into urls(user_id, short_id, long_url) VALUES($1, $2, $3)")
	if err != nil {
		return err
	}

	for _, v := range data {
		if _, err = tx.Exec(ctx, stmt.Name, userID, v.ShortID, v.OriginalURL); err != nil {
			if err = tx.Rollback(ctx); err != nil {
				return fmt.Errorf("unable to rollback: %w", err)
			}
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("unable to commit: %w", err)
	}
	return err
}
