package cached

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository"
)

type DataCachedRepo struct {
	next  repository.DataRepo
	cache repository.Cache
}

func NewDataCachedRepo(next repository.DataRepo, cache repository.Cache) *DataCachedRepo {
	return &DataCachedRepo{
		next:  next,
		cache: cache,
	}
}

func (c *DataCachedRepo) CreateArbitraryData(
	ctx context.Context, uuid string, filename, osPath []byte, filetype, objectKey string, status uint, clientSize int64) (int64, error) {
	return c.next.CreateArbitraryData(ctx, uuid, filename, osPath, filetype, objectKey, status, clientSize)
}

func (c *DataCachedRepo) GetArbitraryData(ctx context.Context, uuid string) ([]models.ArbitraryData, error) {
	key := fmt.Sprintf("user:%s:%s", uuid, "arbitraryData")

	cashed, err := c.cache.Get(ctx, key)
	if err == nil {
		var arbitraryData []models.ArbitraryData
		if err := json.Unmarshal([]byte(cashed), &arbitraryData); err == nil {
			return arbitraryData, nil
		}
	}

	arbitraryData, err := c.next.GetArbitraryData(ctx, uuid)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(&arbitraryData)
	if err == nil {
		err = c.cache.Set(ctx, key, string(data))
		if err != nil {
			slog.Debug("cached fail", "method", "CreateArbitraryData", "error_value", err, "user", uuid)
		}
	}
	slog.Debug("cached", "method", "CreateArbitraryData", "user", uuid)

	return arbitraryData, nil
}

func (c *DataCachedRepo) GetArbitraryDataByID(ctx context.Context, uuid string, dataID int64) (models.ArbitraryData, error) {
	return c.next.GetArbitraryDataByID(ctx, uuid, dataID)
}

func (c *DataCachedRepo) UpdateArbitraryData(ctx context.Context, uuid string, dataID int64, params models.UpdateArbitraryDataParams) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "arbitraryData")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.UpdateArbitraryData(ctx, uuid, dataID, params)
	}
	return nil
}

func (c *DataCachedRepo) DeleteArbitraryData(ctx context.Context, uuid string, dataID int64) error {
	key := fmt.Sprintf("user:%s:%s", uuid, "arbitraryData")

	err := c.cache.Delete(ctx, key)
	if err == nil {
		return c.next.DeleteArbitraryData(ctx, uuid, dataID)
	}
	return nil
}
