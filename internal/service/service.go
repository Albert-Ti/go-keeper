package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/email"
	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	"github.com/aws/smithy-go/ptr"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

const (
	UPLOADING uint = iota
	SUCCESS
	FAILED
)

var (
	ErrAlreadyExists     = errors.New("user already exist")
	ErrInvalidPassword   = errors.New("incorrect password")
	ErrInvalidCodeEmail  = errors.New("invalid confirmation code")
	ErrEmailNotConfirmed = errors.New("email has not been confirmed")
	ErrNoRows            = errors.New("no rows")
	ErrNotFound          = errors.New("NotFound")
	ErrPasswordReused    = errors.New("password was used before")
)

type Service struct {
	db         repository.Database
	opts       *config.Options
	sender     *email.Sender
	objStorage repository.ObjStorage
}

func NewService(
	db repository.Database,
	opts *config.Options,
	sender *email.Sender,
	objStorage repository.ObjStorage,
) *Service {
	return &Service{db, opts, sender, objStorage}
}

func (s *Service) Register(ctx context.Context, email, pass string) error {
	salt, err := utils.RandomHash(8)

	var pgErr *pgconn.PgError
	if err != nil {
		return err
	}

	hash := utils.HashString(salt, pass)
	code := utils.GenerateCodeEmail()

	if err := s.db.AddUser(ctx, email, code, hash); err != nil {
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return ErrAlreadyExists
		}
		return err
	}

	if err := s.sendEmailCode(ctx, email, code); err != nil {
		return err
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email string, pass string) (models.User, error) {
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrNoRows
		}
		return models.User{}, err
	}

	if !user.IsConfirmEmail {
		code := utils.GenerateCodeEmail()
		if err := s.sendEmailCode(ctx, email, code); err != nil {
			return models.User{}, err
		}
		return models.User{}, ErrEmailNotConfirmed
	}

	salt := strings.Split(user.Pass, ".")[0]
	hashPass := utils.HashString(salt, pass)

	if hashPass != user.Pass {
		return models.User{}, ErrInvalidPassword
	}

	return user, nil
}

func (s *Service) ConfirmEmail(ctx context.Context, email, code string) error {
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.EmailCode == code {
		params := models.UpdateUserParams{Email: user.Email, IsConfirmEmail: ptr.Bool(true)}
		if err := s.db.UpdateUser(ctx, params); err != nil {
			return err
		}
	} else {
		return ErrInvalidCodeEmail
	}

	return nil
}

func (s *Service) GetProfile(ctx context.Context, uuid string) (models.Profile, error) {
	user, err := s.db.GetProfile(ctx, uuid)
	if err != nil {
		return models.Profile{}, err
	}

	return user, nil
}

func (s *Service) ChangePass(ctx context.Context, uuid, passOld, passNew string) error {
	user, err := s.db.GetProfile(ctx, uuid)
	if err != nil {
		return err
	}

	// 1. действующий пароль введён верно
	if !utils.CheckPass(user.Pass, passOld) {
		return ErrInvalidPassword
	}

	// 2. новый пароль не совпадает ни с текущим, ни с прошлыми
	listHistory, err := s.db.GetPassList(ctx, uuid)
	if err != nil {
		return err
	}
	for _, h := range append([]string{user.Pass}, listHistory...) {
		if utils.CheckPass(h, passNew) {
			return ErrPasswordReused
		}
	}

	salt := strings.Split(user.Pass, ".")[0]
	hashPassNew := utils.HashString(salt, passNew)

	return s.db.ChangePass(ctx, uuid, user.Pass, hashPassNew)
}

func (s *Service) GetCards(ctx context.Context, uuid string) ([]models.Card, error) {
	return s.db.GetCards(ctx, uuid)
}

func (s *Service) CreateCard(ctx context.Context, uuid string, number string, expiry time.Time) error {
	return s.db.CreateCard(ctx, uuid, number, expiry)
}

func (s *Service) DeleteCard(ctx context.Context, uuid string, cardID int64) error {
	err := s.db.DeleteCard(ctx, uuid, cardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowAffected) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Service) ActivateCard(ctx context.Context, uuid string, cardID int64) error {
	return s.db.ActivateCard(ctx, uuid, cardID)
}

func (s *Service) sendEmailCode(ctx context.Context, email, code string) error {
	if s.sender != nil {
		if err := s.sender.SendConfirmationCode(email, code); err != nil {
			return err
		}
		return nil
	}

	params := models.UpdateUserParams{Email: email, EmailCode: &code}
	err := grpc.SetHeader(ctx, metadata.Pairs("email_code", code))
	if err != nil {
		return err
	}

	if err := s.db.UpdateUser(ctx, params); err != nil {
		return err
	}

	return nil
}

func (s *Service) SaveArbitraryData(
	ctx context.Context, uuid, filename, filetype, osPath string, clientSize int64, reader *utils.GoKeeperStream[*pb.CreateArbitraryDataRequest]) (int64, error) {

	key := uuid + "/" + filename

	id, err := s.db.CreateArbitraryData(
		ctx, uuid, filename, filetype, osPath, key, UPLOADING, clientSize)
	if err != nil {
		return -1, err
	}

	if err := s.objStorage.Put(ctx, key, reader); err != nil {
		if updErr := s.db.UpdateArbitraryData(ctx, uuid, id, models.UpdateArbitraryDataParams{
			Status:    ptr.Uint(FAILED),
			TotalSize: &reader.Total,
		}); updErr != nil {
			return -1, updErr
		}
		return -1, err
	}

	if err := s.db.UpdateArbitraryData(ctx, uuid, id, models.UpdateArbitraryDataParams{
		Status:    ptr.Uint(SUCCESS),
		TotalSize: &reader.Total,
	}); err != nil {
		return -1, err
	}

	return id, nil
}

func (s *Service) GetArbitraryData(ctx context.Context, uuid string) ([]models.ArbitraryData, error) {
	return s.db.GetArbitraryData(ctx, uuid)
}

func (s *Service) DeleteArbitraryData(ctx context.Context, uuid string, dataID int64) error {
	data, err := s.db.GetArbitraryDataByID(ctx, uuid, dataID)
	if err != nil {
		return err
	}
	if err := s.objStorage.Delete(ctx, data.ObjectKey); err != nil {
		return err
	}

	if err := s.db.DeleteArbitraryData(ctx, uuid, dataID); err != nil {
		if errors.Is(err, repository.ErrRowAffected) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *Service) ReloadArbitraryData(ctx context.Context, uuid string, dataID int64, reader *utils.GoKeeperStream[*pb.ReloadArbitraryDataRequest]) error {
	data, err := s.db.GetArbitraryDataByID(ctx, uuid, dataID)
	if err != nil {
		return err
	}

	if err := s.objStorage.Put(ctx, data.ObjectKey, reader); err != nil {
		return err
	}

	if err := s.db.UpdateArbitraryData(ctx, uuid, dataID, models.UpdateArbitraryDataParams{
		Status:    ptr.Uint(SUCCESS),
		TotalSize: &reader.Total,
	}); err != nil {
		return err
	}

	return nil
}
