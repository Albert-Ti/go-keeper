package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/email"
	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	"github.com/aws/smithy-go/ptr"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	ErrAlreadyExists     = errors.New("user already exist")
	ErrInvalidPassword   = errors.New("incorrect password")
	ErrInvalidCodeEmail  = errors.New("invalid confirmation code")
	ErrEmailNotConfirmed = errors.New("email has not been confirmed")
	ErrNoRows            = errors.New("no rows")
	ErrCardNotFound      = errors.New("card not found or not active")
	ErrPasswordReused    = errors.New("password was used before")
)

type Service struct {
	repo   repository.Repository
	opts   *config.Options
	sender *email.Sender
}

func NewService(repo repository.Repository, opts *config.Options, sender *email.Sender) *Service {
	return &Service{repo, opts, sender}
}

func (s *Service) Register(ctx context.Context, email, pass string) error {
	salt, err := utils.RandomHash(8)

	var pgErr *pgconn.PgError
	if err != nil {
		return err
	}

	hash := utils.HashPass(salt, pass)
	code := utils.GenerateCodeEmail()

	if err := s.repo.AddUser(ctx, email, code, hash); err != nil {
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return ErrAlreadyExists
		}
		return err
	}

	if err := s.sendEmailCode(ctx, email, code); err != nil {
		return err
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email string, pass string) (models.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrNoRows
		}
		return models.User{}, err
	}

	if !user.IsConfirmEmail {
		code := utils.GenerateCodeEmail()
		if err := s.sendEmailCode(ctx, email, code); err != nil {
			return models.User{}, err
		}
		return models.User{}, ErrEmailNotConfirmed
	}

	salt := strings.Split(user.Pass, ".")[0]
	hashPass := utils.HashPass(salt, pass)

	if hashPass != user.Pass {
		return models.User{}, ErrInvalidPassword
	}

	return user, nil
}

func (s *Service) ConfirmEmail(ctx context.Context, email, code string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.EmailCode == code {
		params := models.UpdateUserParams{Email: user.Email, IsConfirmEmail: ptr.Bool(true)}
		if err := s.repo.UpdateUser(ctx, params); err != nil {
			return err
		}
	} else {
		return ErrInvalidCodeEmail
	}

	return nil
}

func (s *Service) GetProfile(ctx context.Context, uuid string) (models.Profile, error) {
	user, err := s.repo.GetUserByID(ctx, uuid)
	if err != nil {
		return models.Profile{}, err
	}

	return user, nil
}

func (s *Service) ChangePass(ctx context.Context, uuid, passOld, passNew string) error {
	user, err := s.repo.GetUserByID(ctx, uuid)
	if err != nil {
		return err
	}

	// 1. действующий пароль введён верно
	if !utils.CheckPass(user.Pass, passOld) {
		return ErrInvalidPassword
	}

	// 2. новый пароль не совпадает ни с текущим, ни с прошлыми
	listHistory, err := s.repo.GetPassList(ctx, uuid)
	if err != nil {
		return err
	}
	for _, h := range append([]string{user.Pass}, listHistory...) {
		if utils.CheckPass(h, passNew) {
			return ErrPasswordReused
		}
	}

	salt := strings.Split(user.Pass, ".")[0]
	hashPassNew := utils.HashPass(salt, passNew)

	return s.repo.ChangePass(ctx, uuid, user.Pass, hashPassNew)
}

func (s *Service) GetCards(ctx context.Context, uuid string) ([]models.Card, error) {
	return s.repo.GetCards(ctx, uuid)
}

func (s *Service) CreateCard(ctx context.Context, uuid string, number string, expiry time.Time) error {
	return s.repo.CreateCard(ctx, uuid, number, expiry)
}

func (s *Service) DeleteCard(ctx context.Context, id int64) error {
	err := s.repo.DeleteCard(ctx, id)

	if err != nil {
		return ErrCardNotFound
	}
	return nil
}

func (s *Service) ActivateCard(ctx context.Context, uuid string, id int64) error {
	return s.repo.ActivateCard(ctx, uuid, id)
}

func (s *Service) sendEmailCode(ctx context.Context, email, code string) error {
	if s.sender != nil {
		if err := s.sender.SendConfirmationCode(email, code); err != nil {
			return err
		}
		return nil
	}

	params := models.UpdateUserParams{Email: email, EmailCode: &code}
	err := grpc.SetHeader(ctx, metadata.Pairs("email_code", code))
	if err != nil {
		return err
	}

	if err := s.repo.UpdateUser(ctx, params); err != nil {
		return err
	}

	return nil
}
