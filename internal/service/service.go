package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrAlreadyExists = errors.New("User already exist")
	ErrUnauthorized  = errors.New("Invalid password")
	ErrConfirmEmail  = errors.New("")
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo}
}

func (s *Service) Register(ctx context.Context, email, password string) (string, error) {
	salt, err := utils.RandomHash(8)

	var pgErr *pgconn.PgError
	if err != nil {
		return "", err
	}

	hash := utils.HashPassword(salt, password)

	token, err := s.repo.AddUser(ctx, email, hash)
	if err != nil {
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return "", ErrAlreadyExists
		}
		return "", err
	}

	return token, nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (models.User, error) {
	user, err := s.repo.GetUser(ctx, email)
	if err != nil {
		return models.User{}, err
	}

	salt := strings.Split(user.Password, ".")[0]
	hashPass := utils.HashPassword(salt, password)

	if hashPass != user.Password {
		return models.User{}, ErrUnauthorized
	}

	return user, nil
}
