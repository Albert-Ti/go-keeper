package repository

import (
	"context"
	"fmt"
	"strings"
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

func (pg *PGStorage) AddUser(ctx context.Context, email, code, password string) error {
	sql := `
	INSERT INTO users (email, email_code, password) 
	VALUES ($1, $2, $3)
	`

	_, err := pg.pool.Exec(ctx, sql, email, code, password)

	if err != nil {
		return err
	}

	return nil
}

func (pg *PGStorage) GetUser(ctx context.Context, email string) (models.User, error) {
	sql := `
	SELECT uuid, email, email_code, is_confirm_email, password 
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
			&user.EmailCode,
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

func (pg *PGStorage) UpdateUser(ctx context.Context, p models.UpdateUserParams) error {
	setParts := make([]string, 0, 4)
	args := make([]any, 0, 5)
	argIdx := 1

	if p.EmailCode != nil {
		setParts = append(setParts, fmt.Sprintf("email_code = $%d", argIdx))
		args = append(args, *p.EmailCode)
		argIdx++
	}
	if p.IsConfirmEmail != nil {
		setParts = append(setParts, fmt.Sprintf("is_confirm_email = $%d", argIdx))
		args = append(args, *p.IsConfirmEmail)
		argIdx++
	}
	if p.Password != nil {
		setParts = append(setParts, fmt.Sprintf("password = $%d", argIdx))
		args = append(args, *p.Password)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	args = append(args, p.Email)
	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE email = $%d",
		strings.Join(setParts, ", "),
		argIdx,
	)

	_, err := pg.pool.Exec(ctx, query, args...)
	return err
}
