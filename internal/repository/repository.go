package repository

import (
	"context"

	"github.com/Albert-Ti/go-keeper/internal/models"
)

type Repository interface {
	AddUser(ctx context.Context, email, code, password string) error
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByID(ctx context.Context, uuid string) (models.Profile, error)
	UpdateUser(ctx context.Context, p models.UpdateUserParams) error
	GetCards(ctx context.Context, uuid string) ([]models.Card, error)
}

func NewRepository(connString string) (Repository, error) {
	return NewPGStorage(connString)
}
