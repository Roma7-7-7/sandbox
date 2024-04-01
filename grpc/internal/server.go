package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	tpb "google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/Roma7-7-7/sandbox/grpc/proto"
)

type (
	UserGRPCServer struct {
		userService *UserService
		pb.UnimplementedUserServiceServer
	}
)

func NewUserGRPCService(userService *UserService) *UserGRPCServer {
	return &UserGRPCServer{userService: userService}
}

func (s *UserGRPCServer) CreateUser(_ context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	now := time.Now()

	user := User{
		Name:      req.User.Name,
		Surname:   req.User.Surname,
		Age:       int(req.User.Age),
		CreatedAt: now,
		UpdatedAt: now,
	}

	res, err := s.userService.Create(user)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return nil, status.Errorf(
				codes.AlreadyExists,
				fmt.Sprintf("user already exists: %s", user.Name),
			)
		}

		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("create user: %v", err),
		)
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:        &res.ID,
			Name:      res.Name,
			Surname:   res.Surname,
			Age:       int32(user.Age),
			CreatedAt: &tpb.Timestamp{Seconds: res.CreatedAt.Unix()},
			UpdatedAt: &tpb.Timestamp{Seconds: res.UpdatedAt.Unix()},
			Disabled:  res.Disabled,
		}}, nil
}

func (s *UserGRPCServer) GetUser(_ context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.userService.Get(req.Id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, status.Errorf(
				codes.NotFound,
				fmt.Sprintf("user not found: %s", req.Id),
			)
		}
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("get user: %v", err),
		)
	}

	return &pb.GetUserResponse{
		User: &pb.User{
			Id:        &user.ID,
			Name:      user.Name,
			Surname:   user.Surname,
			Age:       int32(user.Age),
			CreatedAt: &tpb.Timestamp{Seconds: user.CreatedAt.Unix()},
			UpdatedAt: &tpb.Timestamp{Seconds: user.UpdatedAt.Unix()},
			Disabled:  user.Disabled,
		},
	}, nil
}

func (s *UserGRPCServer) DeleteUser(_ context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	err := s.userService.Delete(req.Id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, status.Errorf(
				codes.NotFound,
				fmt.Sprintf("user not found: %s", req.Id),
			)

		}
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("delete user: %v", err),
		)
	}

	return &emptypb.Empty{}, nil
}
