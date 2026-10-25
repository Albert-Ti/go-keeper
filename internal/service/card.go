package service

import (
	"context"
	"errors"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository"
)

type CardService struct {
	repo repository.CardRepo
}

func NewCardService(repo repository.CardRepo) *CardService {
	return &CardService{repo}
}

func (s *CardService) GetCards(ctx context.Context, uuid string) ([]models.Card, error) {
	return s.repo.GetCards(ctx, uuid)
}

func (s *CardService) CreateCard(ctx context.Context, uuid string, number []byte, expiry time.Time) error {
	return s.repo.CreateCard(ctx, uuid, number, expiry)
}

func (s *CardService) DeleteCard(ctx context.Context, uuid string, cardID int64) error {
	err := s.repo.DeleteCard(ctx, uuid, cardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowAffected) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *CardService) ActivateCard(ctx context.Context, uuid string, cardID int64) error {
	err := s.repo.ActivateCard(ctx, uuid, cardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowAffected) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
