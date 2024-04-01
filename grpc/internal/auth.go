package internal

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GetUserID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	userID := md.Get("userID")
	if len(userID) == 0 {
		return ""
	}
	return userID[0]
}

func AuthInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	userID := GetUserID(ctx)
	if userID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "missing user ID")
	}

	return handler(ctx, req)
}
