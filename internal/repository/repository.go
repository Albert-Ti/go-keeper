package repository

import (
	"context"
	"io"
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

	GetArbitraryData(ctx context.Context, uuid string) ([]models.ArbitraryData, error)
	CreateArbitraryData(ctx context.Context, uuid, filename, filetype, objectKey string, status uint) error
	DeleteArbitraryData(ctx context.Context, uuid string, dataID int64) error
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

type ObjStorage interface {
	Put(ctx context.Context, key string, r io.Reader) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

func NewObjStorage(endpoint, accessKey, secretKey string) (ObjStorage, error) {
	return NewS3Client(context.Background(), endpoint, accessKey, secretKey)
}
