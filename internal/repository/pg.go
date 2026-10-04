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

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(connString string) (*Postgres, error) {
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

	return &Postgres{
		pool: pool,
	}, nil
}

func (pg *Postgres) AddUser(ctx context.Context, email, code, pass string) error {
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

func (pg *Postgres) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
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

func (pg *Postgres) GetProfile(ctx context.Context, uuid string) (models.Profile, error) {
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

func (pg *Postgres) UpdateUser(ctx context.Context, p models.UpdateUserParams) error {
	setParts := make([]string, 0, 2)
	args := make([]any, 0, 3)
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

func (pg *Postgres) ChangePass(ctx context.Context, uuid, passOld, passNew string) error {
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

func (pg *Postgres) GetPassList(ctx context.Context, uuid string) ([]string, error) {
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

func (pg *Postgres) GetCards(ctx context.Context, uuid string) ([]models.Card, error) {
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

func (pg *Postgres) CreateCard(ctx context.Context, uuid string, number string, expiry time.Time) error {
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

func (pg *Postgres) ActivateCard(ctx context.Context, uuid string, cardID int64) error {
	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE bank_cards SET active = false WHERE user_uuid = $1 AND active = true`, uuid)
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `UPDATE bank_cards SET active = NOT active WHERE id = $1 and user_uuid = $2`, cardID, uuid)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrRowAffected // защита от активации чужой карты
	}

	return tx.Commit(ctx)
}

func (pg *Postgres) DeleteCard(ctx context.Context, uuid string, cardID int64) error {
	sql := `DELETE FROM bank_cards WHERE user_uuid = $1 AND id = $2 AND active = false`

	tag, err := pg.pool.Exec(ctx, sql, uuid, cardID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRowAffected
	}

	return nil
}

func (pg *Postgres) CreateArbitraryData(
	ctx context.Context, uuid, filename, filetype, osPath, objectKey string, status uint, clientSize int64) (int64, error) {
	row := pg.pool.QueryRow(ctx,
		`INSERT INTO 
		arbitrary_data (user_uuid, name, type, status, os_path, object_key, client_size)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		uuid, filename, filetype, status, osPath, objectKey, clientSize)

	var returningID int64
	err := row.Scan(&returningID)
	if err != nil {
		return -1, err
	}
	return returningID, err
}

func (pg *Postgres) GetArbitraryData(ctx context.Context, uuid string) ([]models.ArbitraryData, error) {
	sql := `
	SELECT id, name, type, status, object_key, os_path, client_size, total_size, created_at
	FROM arbitrary_data 
	WHERE user_uuid = $1
	`
	rows, err := pg.pool.Query(ctx, sql, uuid)
	if err != nil {
		return nil, err
	}

	var list []models.ArbitraryData
	for rows.Next() {
		var data models.ArbitraryData
		err := rows.Scan(
			&data.ID,
			&data.Name,
			&data.Type,
			&data.Status,
			&data.ObjectKey,
			&data.OSPath,
			&data.ClientSize,
			&data.TotalSize,
			&data.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, data)
	}
	return list, nil
}

func (pg *Postgres) GetArbitraryDataByID(ctx context.Context, uuid string, dataID int64) (models.ArbitraryData, error) {
	sql := `
	SELECT status, object_key
	FROM arbitrary_data 
	WHERE user_uuid = $1 AND id = $2
	`
	row := pg.pool.QueryRow(ctx, sql, uuid, dataID)

	var data models.ArbitraryData
	err := row.Scan(
		&data.Status,
		&data.ObjectKey,
	)
	if err != nil {
		return models.ArbitraryData{}, err
	}

	return data, nil
}

func (pg *Postgres) UpdateArbitraryData(
	ctx context.Context, uuid string, dataID int64, params models.UpdateArbitraryDataParams) error {
	setParts := make([]string, 0, 2)
	args := make([]any, 0, 4)
	argIdx := 1

	if params.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *params.Status)
		argIdx++
	}

	if params.TotalSize != nil {
		setParts = append(setParts, fmt.Sprintf("total_size = $%d", argIdx))
		args = append(args, *params.TotalSize)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}
	args = append(args, uuid, dataID)

	query := fmt.Sprintf(
		"UPDATE arbitrary_data SET %s WHERE user_uuid = $%d AND id = $%d",
		strings.Join(setParts, ", "),
		argIdx,
		argIdx+1,
	)

	_, err := pg.pool.Exec(ctx, query, args...)
	return err
}

func (pg *Postgres) DeleteArbitraryData(ctx context.Context, uuid string, dataID int64) error {
	sql := `DELETE FROM arbitrary_data WHERE user_uuid = $1 AND id = $2`

	tag, err := pg.pool.Exec(ctx, sql, uuid, dataID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRowAffected
	}

	return nil
}
