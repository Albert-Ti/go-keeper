package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/models"
)

// pattern Decorator
type CachedDatabase struct {
	next  Database
	cache Cache
}

func NewCachedDatabase(next Database, cache Cache) Database {
	return &CachedDatabase{
		next:  next,
		cache: cache,
	}
}

func (c *CachedDatabase) AddUser(ctx context.Context, email, code, pass string) error {
	return c.next.AddUser(ctx, email, code, pass)
}

func (c *CachedDatabase) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	return c.next.GetUserByEmail(ctx, email)
}

func (c *CachedDatabase) GetProfile(ctx context.Context, uuid string) (models.Profile, error) {
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
		_ = c.cache.Set(
			ctx,
			key,
			string(data),
			5*time.Minute,
		)
	}
	return profile, nil
}

func (c *CachedDatabase) UpdateUser(ctx context.Context, p models.UpdateUserParams) error {
	return c.next.UpdateUser(ctx, p)
}

func (c *CachedDatabase) ChangePass(ctx context.Context, uuid, passOld, passNew string) error {
	return c.next.ChangePass(ctx, uuid, passOld, passNew)
}

func (c *CachedDatabase) GetCards(ctx context.Context, uuid string) ([]models.Card, error) {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	cashed, err := c.cache.Get(ctx, key)

	if err == nil {
		var cards []models.Card
		if err := json.Unmarshal([]byte(cashed), &cards); err == nil {
			return cards, nil
		}
	}

	cards, err := c.next.GetCards(ctx, uuid)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(cards)
	if err == nil {
		_ = c.cache.Set(
			ctx,
			key,
			string(data),
			5*time.Minute,
		)
	}
	return cards, nil
}

func (c *CachedDatabase) CreateCard(ctx context.Context, uuid string, number string, expiry time.Time) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.CreateCard(ctx, uuid, number, expiry)
	}

	return nil
}

func (c *CachedDatabase) DeleteCard(ctx context.Context, uuid string, cardID int64) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.DeleteCard(ctx, uuid, cardID)

	}
	return nil
}

func (c *CachedDatabase) ActivateCard(ctx context.Context, uuid string, cardID int64) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.ActivateCard(ctx, uuid, cardID)

	}
	return nil
}

func (c *CachedDatabase) GetPassList(ctx context.Context, uuid string) ([]string, error) {
	return c.next.GetPassList(ctx, uuid)
}

func (c *CachedDatabase) CreateData(ctx context.Context)
func (c *CachedDatabase) GetData(ctx context.Context)
func (c *CachedDatabase) GetUserData(ctx context.Context)
func (c *CachedDatabase) UpdateData(ctx context.Context)
func (c *CachedDatabase) DeleteData(ctx context.Context)
