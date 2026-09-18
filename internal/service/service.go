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
		return nil
	}

	salt := strings.Split(user.Pass, ".")[0]
	hashPassOld := utils.HashPass(salt, passOld)

	if hashPassOld != user.Pass {
		return ErrInvalidPassword
	}

	hashPassNew := utils.HashPass(salt, passNew)

	params := models.UpdateUserParams{UUID: uuid, Pass: &hashPassNew}
	return s.repo.UpdateUser(ctx, params)
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
