package cached

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository"
)

type CardCachedRepo struct {
	next  repository.CardRepo
	cache repository.Cache
}

func NewCardCachedRepo(next repository.CardRepo, cache repository.Cache) *CardCachedRepo {
	return &CardCachedRepo{
		next:  next,
		cache: cache,
	}
}

func (c *CardCachedRepo) GetCards(ctx context.Context, uuid string) ([]models.Card, error) {
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
		err = c.cache.Set(ctx, key, string(data))
		if err != nil {
			slog.Debug("cached fail", "method", "GetCards", "error_value", err, "user", uuid)
		}
	}
	slog.Debug("cached", "method", "GetCards", "user", uuid, "user", uuid)

	return cards, nil
}

func (c *CardCachedRepo) CreateCard(ctx context.Context, uuid string, number []byte, expiry time.Time) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.CreateCard(ctx, uuid, number, expiry)
	}

	return nil
}

func (c *CardCachedRepo) ActivateCard(ctx context.Context, uuid string, cardID int64) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.ActivateCard(ctx, uuid, cardID)
	}
	return nil
}

func (c *CardCachedRepo) DeleteCard(ctx context.Context, uuid string, cardID int64) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "cards")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.DeleteCard(ctx, uuid, cardID)

	}
	return nil
}
