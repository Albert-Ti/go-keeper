package repository

import (
	"context"

	"github.com/Albert-Ti/go-keeper/internal/models"
)

type Repository interface {
	AddUser(ctx context.Context, email, password string) (int, error)
	GetUser(ctx context.Context, email string) (models.User, error)
}

func NewRepository(connString string) (Repository, error) {
	return NewPGStorage(connString)
}
