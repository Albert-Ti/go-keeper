package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/models"
)

// pattern Decorator
type CachedRepository struct {
	next  Repository
	cache Cache
}

func NewCachedRepository(next Repository, cache Cache) Repository {
	return &CachedRepository{
		next:  next,
		cache: cache,
	}
}

func (c *CachedRepository) AddUser(ctx context.Context, email, code, pass string) error {
	return c.next.AddUser(ctx, email, code, pass)
}

func (c *CachedRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	return c.next.GetUserByEmail(ctx, email)
}

func (c *CachedRepository) GetProfile(ctx context.Context, uuid string) (models.Profile, error) {
	key := fmt.Sprintf("user:%s", uuid)
	cashed, err := c.cache.Get(ctx, key)

	if err == nil {
		var profile models.Profile
		if err := json.Unmarshal([]byte(cashed), &profile); err == nil {
			return profile, nil
		}
	}

	// pg method
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

func (c *CachedRepository) UpdateUser(ctx context.Context, p models.UpdateUserParams) error {
	return c.next.UpdateUser(ctx, p)
}

func (c *CachedRepository) ChangePass(ctx context.Context, uuid, passOld, passNew string) error {
	return c.next.ChangePass(ctx, uuid, passOld, passNew)
}

func (c *CachedRepository) GetCards(ctx context.Context, uuid string) ([]models.Card, error) {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	cashed, err := c.cache.Get(ctx, key)

	if err == nil {
		var cards []models.Card
		if err := json.Unmarshal([]byte(cashed), &cards); err == nil {
			return cards, nil
		}
	}

	// pg method
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

func (c *CachedRepository) CreateCard(ctx context.Context, uuid string, number string, expiry time.Time) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.CreateCard(ctx, uuid, number, expiry)
	}

	return nil
}

func (c *CachedRepository) DeleteCard(ctx context.Context, uuid string, cardID int64) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.DeleteCard(ctx, uuid, cardID)

	}
	return nil
}

func (c *CachedRepository) ActivateCard(ctx context.Context, uuid string, cardID int64) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.ActivateCard(ctx, uuid, cardID)

	}
	return nil
}

func (c *CachedRepository) GetPassList(ctx context.Context, uuid string) ([]string, error) {
	return c.next.GetPassList(ctx, uuid)
}
