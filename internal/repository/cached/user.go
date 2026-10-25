package cached

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository"
)

type UserCachedRepo struct {
	next  repository.UserRepo
	cache repository.Cache
}

func NewUserCachedRepo(next repository.UserRepo, cache repository.Cache) *UserCachedRepo {
	return &UserCachedRepo{
		next:  next,
		cache: cache,
	}
}

func (c *UserCachedRepo) AddUser(ctx context.Context, email, code, pass string) error {
	return c.next.AddUser(ctx, email, code, pass)
}

func (c *UserCachedRepo) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	return c.next.GetUserByEmail(ctx, email)
}

func (c *UserCachedRepo) GetProfile(ctx context.Context, uuid string) (models.Profile, error) {
	key := fmt.Sprintf("user:%s", uuid)
	cashed, err := c.cache.Get(ctx, key)

	if err == nil {
		var profile models.Profile
		if err := json.Unmarshal([]byte(cashed), &profile); err == nil {
			return profile, nil
		}
	}

	profile, err := c.next.GetProfile(ctx, uuid)
	if err != nil {
		return models.Profile{}, err
	}

	data, err := json.Marshal(profile)
	if err == nil {
		err = c.cache.Set(ctx, key, string(data))
		if err != nil {
			slog.Debug("cached fail", "method", "GetProfile", "error_value", err, "user", uuid)
		}
	}
	slog.Debug("cached", "method", "GetProfile", "user", uuid)

	return profile, nil
}

func (c *UserCachedRepo) UpdateUser(ctx context.Context, p models.UpdateUserParams) error {
	return c.next.UpdateUser(ctx, p)
}

func (c *UserCachedRepo) ChangePass(ctx context.Context, uuid, passCurr, passNew string) error {
	return c.next.ChangePass(ctx, uuid, passCurr, passNew)
}

func (c *UserCachedRepo) GetOldPassList(ctx context.Context, uuid string) ([]string, error) {
	return c.next.GetOldPassList(ctx, uuid)
}
