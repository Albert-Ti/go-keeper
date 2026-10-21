package service

import (
	"context"
	"encoding/base64"
	"errors"
	"io"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/aws/smithy-go/ptr"
)

type DataService struct {
	repo    repository.DataRepo
	storage repository.ObjStorage
}

func NewDataService(repo repository.DataRepo, storage repository.ObjStorage) *DataService {
	return &DataService{repo, storage}
}

func (s *DataService) SaveArbitraryData(
	ctx context.Context, uuid string, filename, osPath []byte, filetype string, clientSize int64, reader *utils.GoKeeperStream[*pb.CreateArbitraryDataRequest]) (int64, error) {

	objectKey := uuid + "/" + base64.StdEncoding.EncodeToString(filename) + "." + filetype

	id, err := s.repo.CreateArbitraryData(
		ctx, uuid, filename, osPath, filetype, objectKey, uint(models.DataStatusUploading), clientSize)
	if err != nil {
		return -1, err
	}

	if err := s.storage.Put(ctx, objectKey, reader); err != nil {
		if updErr := s.repo.UpdateArbitraryData(ctx, uuid, id, models.UpdateArbitraryDataParams{
			Status:    ptr.Uint(uint(models.DataStatusFailed)),
			TotalSize: &reader.Total,
		}); updErr != nil {
			return -1, updErr
		}
		return -1, err
	}

	if err := s.repo.UpdateArbitraryData(ctx, uuid, id, models.UpdateArbitraryDataParams{
		Status:    ptr.Uint(uint(models.DataStatusSuccess)),
		TotalSize: &reader.Total,
	}); err != nil {
		return -1, err
	}

	return id, nil
}

func (s *DataService) GetArbitraryData(ctx context.Context, uuid string) ([]models.ArbitraryData, error) {
	return s.repo.GetArbitraryData(ctx, uuid)
}

func (s *DataService) DeleteArbitraryData(ctx context.Context, uuid string, dataID int64) error {
	data, err := s.repo.GetArbitraryDataByID(ctx, uuid, dataID)
	if err != nil {
		return err
	}
	if err := s.storage.Delete(ctx, data.ObjectKey); err != nil {
		return err
	}

	if err := s.repo.DeleteArbitraryData(ctx, uuid, dataID); err != nil {
		if errors.Is(err, repository.ErrRowAffected) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *DataService) ReloadArbitraryData(ctx context.Context, uuid string, dataID int64, reader *utils.GoKeeperStream[*pb.ReloadArbitraryDataRequest]) error {
	data, err := s.repo.GetArbitraryDataByID(ctx, uuid, dataID)
	if err != nil {
		return err
	}

	if err := s.storage.Put(ctx, data.ObjectKey, reader); err != nil {
		return err
	}

	if err := s.repo.UpdateArbitraryData(ctx, uuid, dataID, models.UpdateArbitraryDataParams{
		Status:    ptr.Uint(uint(models.DataStatusSuccess)),
		TotalSize: &reader.Total,
	}); err != nil {
		return err
	}

	return nil
}

func (s *DataService) DownloadArbitraryData(ctx context.Context, uuid string, dataID int64) (io.ReadCloser, int64, error) {
	data, err := s.repo.GetArbitraryDataByID(ctx, uuid, dataID)
	if err != nil {
		return nil, 0, err
	}

	obj, err := s.storage.Get(ctx, data.ObjectKey)
	if err != nil {
		return nil, 0, err
	}

	return obj, data.TotalSize, nil
}
