package repository

import (
	"context"

	"github.com/Albert-Ti/go-keeper/internal/models"
)

type Repository interface {
	AddUser(ctx context.Context, email, code, password string) error
	GetUser(ctx context.Context, email string) (models.User, error)
	UpdateUser(ctx context.Context, p models.UpdateUserParams) error
}

func NewRepository(connString string) (Repository, error) {
	return NewPGStorage(connString)
}
