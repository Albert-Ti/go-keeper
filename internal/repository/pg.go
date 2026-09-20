package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRowAffected = errors.New("rows affected 0")

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

func (pg *PGStorage) AddUser(ctx context.Context, email, code, pass string) error {
	sql := `
	INSERT INTO users (email, email_code, pass) 
	VALUES ($1, $2, $3)
	`

	_, err := pg.pool.Exec(ctx, sql, email, code, pass)

	if err != nil {
		return err
	}

	return nil
}

func (pg *PGStorage) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	sql := `
	SELECT uuid, email, email_code, is_confirm_email, pass 
	FROM users 
	WHERE email = $1
	`
	var user models.User

	err := pg.pool.QueryRow(ctx, sql, email).Scan(
		&user.UUID,
		&user.Email,
		&user.EmailCode,
		&user.IsConfirmEmail,
		&user.Pass,
	)

	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (pg *PGStorage) GetUserByID(ctx context.Context, uuid string) (models.Profile, error) {
	sql := `
	SELECT email, pass, created_at
	FROM users 
	WHERE uuid = $1
	`
	var user models.Profile

	err := pg.pool.QueryRow(ctx, sql, uuid).Scan(
		&user.Email,
		&user.Pass,
		&user.CreatedAt,
	)

	if err != nil {
		return models.Profile{}, err
	}

	return user, nil
}

// UpdateUser - Универсальный метод db, который легко масштабируется при увеличении полей у таблицы users.
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

	if len(setParts) == 0 {
		return nil
	}

	var query string

	if p.Email != "" {
		args = append(args, p.Email)
		query = fmt.Sprintf(
			"UPDATE users SET %s WHERE email = $%d",
			strings.Join(setParts, ", "),
			argIdx,
		)
	}

	if p.UUID != "" {
		args = append(args, p.UUID)

		query = fmt.Sprintf(
			"UPDATE users SET %s WHERE uuid = $%d",
			strings.Join(setParts, ", "),
			argIdx,
		)
	}

	_, err := pg.pool.Exec(ctx, query, args...)
	return err
}

// ChangePass - отдельное обновление пароля с сохранением в истории.
func (pg *PGStorage) ChangePass(ctx context.Context, uuid, passOld, passNew string) error {
	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE users SET pass = $1 WHERE uuid = $2`, passNew, uuid)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO pass_list (old_pass, user_uuid) VALUES ($1, $2)`, passOld, uuid)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (pg *PGStorage) GetCards(ctx context.Context, uuid string) ([]models.Card, error) {
	sql := `
	SELECT id, card_number, expiry_date, active
	FROM bank_cards 
	WHERE user_uuid = $1
	`
	rows, err := pg.pool.Query(ctx, sql, uuid)
	if err != nil {
		return nil, err
	}

	var list []models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(&card.ID, &card.CardNumber, &card.ExpiryDate, &card.Active)
		if err != nil {
			return nil, err
		}
		list = append(list, card)
	}
	return list, nil
}

func (pg *PGStorage) CreateCard(ctx context.Context, uuid string, number string, expiry time.Time) error {
	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE bank_cards SET active = false WHERE user_uuid = $1 AND active = true`,
		uuid,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO bank_cards (user_uuid, card_number, expiry_date, active) VALUES ($1, $2, $3, $4)`,
		uuid, number, expiry, true,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (pg *PGStorage) DeleteCard(ctx context.Context, id int64) error {
	sql := `DELETE FROM bank_cards WHERE id = $1`

	tag, err := pg.pool.Exec(ctx, sql, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRowAffected
	}

	return nil
}

func (pg *PGStorage) ActivateCard(ctx context.Context, uuid string, id int64) error {
	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE bank_cards SET active = false WHERE user_uuid = $1 AND active = true`, uuid)
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `UPDATE bank_cards SET active = NOT active WHERE id = $1 and user_uuid = $2`, id, uuid)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrRowAffected // защита от активации чужой карты
	}

	return tx.Commit(ctx)
}

func (pg *PGStorage) GetPassList(ctx context.Context, uuid string) ([]string, error) {
	rows, err := pg.pool.Query(ctx, `SELECT old_pass FROM pass_list WHERE user_uuid = $1`, uuid)

	if err != nil {
		return nil, err
	}

	var list []string
	if rows.Next() {
		var pass string
		err := rows.Scan(&pass)

		if err != nil {
			return nil, err
		}

		list = append(list, pass)
	}

	return list, nil
}
