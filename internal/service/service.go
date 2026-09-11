package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

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
	ErrUnauthorized      = errors.New("incorrect email or password")
	ErrInvalidCodeEmail  = errors.New("invalid confirmation code")
	ErrEmailNotConfirmed = errors.New("email has not been confirmed")
	ErrNoRows            = errors.New("no rows")
)

type Service struct {
	repo   repository.Repository
	opts   *config.Options
	sender *email.Sender
}

func NewService(repo repository.Repository, opts *config.Options, sender *email.Sender) *Service {
	return &Service{repo, opts, sender}
}

func (s *Service) Register(ctx context.Context, email, password string) error {
	salt, err := utils.RandomHash(8)

	var pgErr *pgconn.PgError
	if err != nil {
		return err
	}

	hash := utils.HashPassword(salt, password)
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

func (s *Service) Login(ctx context.Context, email string, password string) (models.User, error) {
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

	salt := strings.Split(user.Password, ".")[0]
	hashPass := utils.HashPassword(salt, password)

	if hashPass != user.Password {
		return models.User{}, ErrUnauthorized
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

func (s *Service) UpdatePassword(ctx context.Context, uuid string) error {
	return nil
}

func (s *Service) GetCards(ctx context.Context, uuid string) ([]models.Card, error) {
	return s.repo.GetCards(ctx, uuid)
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
