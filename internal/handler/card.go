package handler

import (
	"context"
	"errors"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *GrpcServer) CreateCard(ctx context.Context, in *pb.CreateCardRequest) (*pb.CreateCardResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	if !utils.IsLuhnValid(string(in.GetCardNumber())) {
		return nil, status.Error(codes.InvalidArgument, "invalid card number")
	}

	if len(in.GetCardNumber()) != 16 {
		return nil, status.Error(codes.InvalidArgument, "invalid card number, must be 16 digits")
	}

	if len(in.GetExpiryDate()) != 5 {
		return nil, status.Error(codes.InvalidArgument, "invalid expiry date format, expected MM/YY")
	}

	t, err := time.Parse("01/06", in.GetExpiryDate())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid expiry date format, expected MM/YY")
	}

	expiry := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC)

	if expiry.Before(time.Now().UTC()) {
		return nil, status.Error(codes.InvalidArgument, "card expiry date has already passed")
	}

	if err := g.CardSvc.CreateCard(ctx, uuid, in.GetCardNumber(), expiry); err != nil {
		return nil, status.Error(codes.Internal, "failed to create card")
	}

	return &pb.CreateCardResponse{}, nil
}

func (g *GrpcServer) GetCards(ctx context.Context, in *pb.CardsRequest) (*pb.CardsResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	cards, err := g.CardSvc.GetCards(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get cards")
	}

	var list []*pb.CardData
	for _, v := range cards {
		list = append(list, pb.CardData_builder{
			Id:         v.ID,
			Active:     v.Active,
			CardNumber: v.CardNumber[12:],
			ExpiryDate: timestamppb.New(v.ExpiryDate),
		}.Build())
	}

	response := pb.CardsResponse_builder{
		Cards: list,
	}.Build()

	return response, nil
}

func (g *GrpcServer) ActivateCard(ctx context.Context, in *pb.ActiveCardRequest) (*pb.ActiveCardResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	err = g.CardSvc.ActivateCard(ctx, uuid, in.GetId())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "card not found")
		}
		return nil, status.Error(codes.Internal, "failed to activate card")
	}
	return &pb.ActiveCardResponse{}, nil
}

func (g *GrpcServer) DeleteCard(ctx context.Context, in *pb.DeleteCardRequest) (*pb.DeleteCardResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	err = g.CardSvc.DeleteCard(ctx, uuid, in.GetId())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "card not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete card")
	}
	return &pb.DeleteCardResponse{}, nil
}
