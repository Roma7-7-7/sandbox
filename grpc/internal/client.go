package internal

import (
	"context"
	"fmt"

	"github.com/Roma7-7-7/sandbox/grpc/proto"
)

type (
	Client struct {
		proto.UserServiceClient
	}
)

func NewClient(client proto.UserServiceClient) *Client {
	return &Client{client}
}

func (c *Client) CreateUser(ctx context.Context, name string) (*User, error) {
	req := &proto.CreateUserRequest{
		User: &proto.User{
			Name: name,
		},
	}
	res, err := c.UserServiceClient.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &User{
		ID:        res.User.GetId(),
		Name:      res.User.GetName(),
		Surname:   res.User.GetSurname(),
		Age:       int(res.User.GetAge()),
		CreatedAt: res.User.CreatedAt.AsTime(),
		UpdatedAt: res.User.UpdatedAt.AsTime(),
		Disabled:  res.User.Disabled,
	}, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	req := &proto.GetUserRequest{
		Id: id,
	}
	res, err := c.UserServiceClient.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &User{
		ID:        res.User.GetId(),
		Name:      res.User.GetName(),
		Surname:   res.User.GetSurname(),
		Age:       int(res.User.GetAge()),
		CreatedAt: res.User.CreatedAt.AsTime(),
		UpdatedAt: res.User.UpdatedAt.AsTime(),
		Disabled:  res.User.Disabled,
	}, nil
}

func (c *Client) DeleteUser(ctx context.Context, id string) error {
	req := &proto.DeleteUserRequest{
		Id: id,
	}
	_, err := c.UserServiceClient.DeleteUser(ctx, req)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
