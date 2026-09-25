package repository

import (
	"context"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/models"
)

type Database interface {
	AddUser(ctx context.Context, email, code, pass string) error
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetProfile(ctx context.Context, uuid string) (models.Profile, error)
	UpdateUser(ctx context.Context, p models.UpdateUserParams) error
	ChangePass(ctx context.Context, uuid, passOld, passNew string) error
	GetPassList(ctx context.Context, uuid string) ([]string, error)

	GetCards(ctx context.Context, uuid string) ([]models.Card, error)
	CreateCard(ctx context.Context, uuid string, number string, expiry time.Time) error
	DeleteCard(ctx context.Context, uuid string, cardID int64) error
	ActivateCard(ctx context.Context, uuid string, cardID int64) error

	CreateData(ctx context.Context)
	GetData(ctx context.Context)
	GetUserData(ctx context.Context)
	UpdateData(ctx context.Context)
	DeleteData(ctx context.Context)
}

func NewDatabase(connString string) (Database, error) {
	return NewPostgres(connString)
}

type Cache interface {
	Ping(ctx context.Context)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
}

func NewCache(addr, pass string) Cache {
	return NewRedis(addr, pass)
}

type ObjectStorage interface {
	Put(ctx context.Context, key string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

func NewObjectStorage(endpoint string) (ObjectStorage, error) {
	return NewS3Client(context.Background(), endpoint)
}
