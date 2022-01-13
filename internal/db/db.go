package db

import (
	"context"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/repository"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Entity repository.Entity

type T struct {
	*pgxpool.Pool
}

var Pool T

func (d *T) New(ctx context.Context, url string) error {
	var err error
	d.Pool, err = pgxpool.Connect(ctx, url)
	if err != nil {
		return err
	}

	// создание таблицы
	sql1 := "create table if not exists urls (user_id varchar(32) not null, short_path varchar(32) not null, long_url varchar(512) not null)"
	_, err = d.Exec(context.Background(), sql1)
	if err != nil {
		panic(err)
	}

	return err
}

func addEntity(ctx context.Context, e Entity) error {
	sql := "insert into urls1 values ($1, $2, $3)"
	_, err := Pool.Exec(ctx, sql, e.UserID, e.ID, e.URL)
	return err
}

func SelectByUser(ctx context.Context, userID string) ([]Entity, error) {
	rows, err := Pool.Query(ctx, "select * from urls where user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	e := Entity{}
	eArray := make([]Entity, 0, 10)
	for rows.Next() {
		err = rows.Scan(&e.UserID, &e.ID, &e.URL)
		if err != nil {
			return nil, err
		}
		eArray = append(eArray, e)
	}
	return eArray, nil
}
