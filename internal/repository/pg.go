package repository

import (
	"context"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/utils"
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

func (pg *PGStorage) AddUser(ctx context.Context, email, password string) (string, error) {
	emailToken := utils.GenerateUUID()
	sql := `
	INSERT INTO users (email, email_token, password) 
	VALUES ($1, $2, $3)
	`

	_, err := pg.pool.Exec(ctx, sql, email, emailToken, password)

	if err != nil {
		return "", err
	}

	return emailToken, nil
}

func (pg *PGStorage) GetUser(ctx context.Context, email string) (models.User, error) {
	sql := `
	SELECT uuid, email, email_token, is_confirm_email, password 
	FROM users 
	WHERE email = $1
	`

	rows, err := pg.pool.Query(ctx, sql, email)
	if err != nil {
		return models.User{}, err
	}

	var user models.User
	for rows.Next() {
		err := rows.Scan(
			&user.UUID,
			&user.Email,
			&user.EmailToken,
			&user.IsConfirmEmail,
			&user.Password,
		)

		if err != nil {
			return models.User{}, err
		}
	}

	if err = rows.Err(); err != nil {
		return models.User{}, err
	}

	return user, nil
}
