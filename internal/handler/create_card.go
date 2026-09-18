package handler

import (
	"context"
	"time"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) CreateCard(ctx context.Context, in *pb.CreateCardRequest) (*pb.CreateCardResponse, error) {
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

	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	if err := g.Svc.CreateCard(ctx, uuid, in.GetCardNumber(), expiry); err != nil {
		return nil, status.Error(codes.Internal, "failed to create card")
	}

	return &pb.CreateCardResponse{}, nil
}
