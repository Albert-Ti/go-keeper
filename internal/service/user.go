package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

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

type UserService struct {
	repo   repository.UserRepo
	sender *email.Sender
}

func NewUserService(repo repository.UserRepo, sender *email.Sender) *UserService {
	return &UserService{repo, sender}
}

func (s *UserService) Register(ctx context.Context, email, pass string) error {
	salt, err := utils.RandomHash(8)
	if err != nil {
		return err
	}

	hashPass := utils.HashString(salt, pass)
	code, err := utils.GenerateCodeEmail()
	if err != nil {
		return err
	}
	var pgErr *pgconn.PgError
	if err := s.repo.AddUser(ctx, email, string(code), hashPass); err != nil {
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return ErrAlreadyExists
		}
		return err
	}

	if err := s.sendEmailCode(ctx, email, string(code)); err != nil {
		return err
	}
	return nil
}

func (s *UserService) Login(ctx context.Context, email string, pass string) (models.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrNoRows
		}
		return models.User{}, err
	}

	if !user.IsConfirmEmail {
		code, err := utils.GenerateCodeEmail()
		if err != nil {
			return models.User{}, err
		}
		params := models.UpdateUserParams{Email: email, EmailCode: ptr.String(string(code))}
		if err := s.repo.UpdateUser(ctx, params); err != nil {
			return models.User{}, err
		}
		if err := s.sendEmailCode(ctx, email, string(code)); err != nil {
			return models.User{}, err
		}
		return models.User{}, ErrEmailNotConfirmed
	}

	salt := strings.Split(user.Pass, ".")[0]
	hashPass := utils.HashString(salt, pass)

	if hashPass != user.Pass {
		return models.User{}, ErrInvalidPassword
	}

	return user, nil
}

func (s *UserService) ConfirmEmail(ctx context.Context, email, code string) error {
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

func (s *UserService) GetProfile(ctx context.Context, uuid string) (models.Profile, error) {
	user, err := s.repo.GetProfile(ctx, uuid)
	if err != nil {
		return models.Profile{}, err
	}

	return user, nil
}

func (s *UserService) ChangePass(ctx context.Context, uuid, passCurr, passNew string) error {
	user, err := s.repo.GetProfile(ctx, uuid)
	if err != nil {
		return err
	}

	// 1. действующий пароль введён верно
	if !utils.CheckPass(user.Pass, passCurr) {
		return ErrInvalidPassword
	}

	passList, err := s.repo.GetOldPassList(ctx, uuid)
	if err != nil {
		return err
	}

	for _, pass := range append([]string{user.Pass}, passList...) {
		// 2. новый пароль не совпадает ни с текущим, ни с прошлыми
		if utils.CheckPass(pass, passNew) {
			return ErrPasswordReused
		}
	}

	salt := strings.Split(user.Pass, ".")[0]
	hashPassNew := utils.HashString(salt, passNew)

	return s.repo.ChangePass(ctx, uuid, user.Pass, hashPassNew)
}

func (s *UserService) sendEmailCode(ctx context.Context, email, code string) error {
	if s.sender != nil {
		if err := s.sender.SendConfirmationCode(email, code); err != nil {
			return err
		}
		return nil
	}

	err := grpc.SetHeader(ctx, metadata.Pairs("email_code", code))
	if err != nil {
		return err
	}

	return nil
}
