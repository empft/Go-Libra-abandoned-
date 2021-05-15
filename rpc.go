package main

import (
	"context"
	"sync"

	pb "github.com/stevealexrs/Go-Libra/connector/public"
)

type publicServer struct {
	pb.UnimplementedPublicServer
	mu sync.Mutex
}

func (s *publicServer) RequestInvitationCode(ctx context.Context, invitation *pb.InvitationRequest) (*pb.Empty, error) {
	

	return &pb.Empty{}, nil
}

func (s *publicServer) IsUsernameTaken(ctx context.Context, user *pb.User) (*pb.BoolValue, error) {
	return &pb.BoolValue{Value: true}, nil
}

func (s *publicServer) RegisterAccount(ctx context.Context, request *pb.RegisterRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *publicServer) LoginAccount(ctx context.Context, request *pb.LoginRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *publicServer) ForgetAccount(ctx context.Context, request *pb.ForgetRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}