package handler

import (
	"context"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *GrpcServer) GetCards(ctx context.Context, in *pb.CardsRequest) (*pb.CardsResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	cards, err := g.Svc.GetCards(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get cards")
	}

	var list []*pb.CardData
	for _, v := range cards {
		list = append(list, pb.CardData_builder{
			Id:         v.ID,
			Active:     v.Active,
			CardNumber: string(v.CardNumber[12:]),
			ExpiryDate: timestamppb.New(v.ExpiryDate),
		}.Build())
	}

	response := pb.CardsResponse_builder{
		Cards: list,
	}.Build()

	return response, nil
}
