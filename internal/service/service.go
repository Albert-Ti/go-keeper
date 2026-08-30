package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository"
	"github.com/Albert-Ti/go-keeper/internal/utils"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo}
}

func (s *Service) Register(ctx context.Context, email, password string) (int, error) {
	salt, err := utils.RandomHash(8)
	if err != nil {
		return 0, err
	}
	hash := utils.HashPassword(salt, password)
	return s.repo.AddUser(ctx, email, hash)
}

func (s *Service) Login(ctx context.Context, email string, password string) (models.User, error) {
	user, err := s.repo.GetUser(ctx, email)
	if err != nil {
		return models.User{}, err
	}

	salt := strings.Split(user.Password, ".")[0]
	fmt.Println(salt)

	return models.User{}, nil
}
