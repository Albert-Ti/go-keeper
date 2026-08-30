package repository

import (
	"context"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStorage struct {
	pool *pgxpool.Pool
}

func NewPGStorage(connString string) (*PGStorage, error) {
	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	poolCfg.MinConns = 4                       // держим минимум 4 живых соединения постоянно
	poolCfg.MaxConns = 20                      // верхний предел под нагрузку
	poolCfg.MaxConnIdleTime = 5 * time.Minute  // не закрывать соединения слишком агрессивно
	poolCfg.MaxConnLifetime = 30 * time.Minute // периодическая ротация соединений (защита от "протухания")

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, err
	}

	return &PGStorage{
		pool: pool,
	}, nil
}

func (pg *PGStorage) AddUser(ctx context.Context, email, password string) (int, error) {
	sql := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`

	var userID int
	errUser := pg.pool.QueryRow(ctx, sql, email, password).Scan(&userID)

	if errUser != nil {
		return 0, errUser
	}

	return userID, nil
}

func (pg *PGStorage) GetUser(ctx context.Context, email string) (models.User, error) {
	sql := `SELECT * FROM users WHERE email = $1`

	res, err := pg.pool.Query(ctx, sql, email)
	if err != nil {
		return models.User{}, err
	}

	var user models.User
	for res.Next() {
		if err := res.Scan(&user.Email, &user.Password); err != nil {
			return models.User{}, err
		}
	}

	return user, nil
}
