package main

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	igrpc "github.com/Roma7-7-7/sandbox/grpc/internal"
	pb "github.com/Roma7-7-7/sandbox/grpc/proto"
)

func main() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic:", r)
				err = status.Errorf(codes.Internal, "panic: %v", r)
			}
		}()
		return handler(ctx, req)
	}, igrpc.AuthInterceptor))
	pb.RegisterUserServiceServer(s, igrpc.NewUserGRPCService(igrpc.NewUserService()))
	if err = s.Serve(lis); err != nil {
		panic(err)
	}
}
