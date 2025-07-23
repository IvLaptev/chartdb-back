package handler

import (
	"context"
	"fmt"
	"log/slog"

	chartdbapi "github.com/IvLaptev/chartdb-back/api/chartdb/v1"
	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/service/user"
	xerrors "github.com/IvLaptev/chartdb-back/pkg/errors"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
	"github.com/IvLaptev/chartdb-back/pkg/utils/ptr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const minPasswordLength = 8

type UserHandler struct {
	Logger      *slog.Logger
	UserService user.Service

	chartdbapi.UnimplementedUserServiceServer
}

func (h *UserHandler) Create(ctx context.Context, req *chartdbapi.CreateUserRequest) (*chartdbapi.User, error) {
	var passwordHash *string
	if req.Password != "" {
		if len(req.Password) < minPasswordLength {
			return nil, xerrors.WrapInvalidArgument(fmt.Errorf("password must be at least %d characters long", minPasswordLength))
		}
		passwordHash = ptr.To(utils.SHA1(req.Password))
	}

	userModel, err := h.UserService.CreateUser(ctx, &user.CreateUserParams{
		Login:        req.Login,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return userToPB(userModel), nil
}

func (h *UserHandler) Login(ctx context.Context, req *chartdbapi.LoginUserRequest) (*chartdbapi.LoginUserResponse, error) {
	token, err := h.UserService.LoginUser(ctx, &user.LoginUserParams{
		Login:        req.Login,
		PasswordHash: utils.SHA1(req.Password),
	})
	if err != nil {
		return nil, fmt.Errorf("login user: %w", err)
	}

	return &chartdbapi.LoginUserResponse{Token: token}, nil
}

func (h *UserHandler) Confirm(ctx context.Context, req *chartdbapi.ConfirmUserRequest) (*chartdbapi.User, error) {
	userModel, err := h.UserService.ConfirmUser(ctx, &user.ConfirmUserParams{
		UserConfirmationID: model.UserConfirmationID(req.Cid),
	})
	if err != nil {
		return nil, fmt.Errorf("confirm user: %w", err)
	}

	return userToPB(userModel), nil
}

func userToPB(user *model.User) *chartdbapi.User {
	var userType chartdbapi.UserType
	switch user.Type {
	case model.UserTypeAdmin:
		userType = chartdbapi.UserType_USER_TYPE_ADMIN
	case model.UserTypeTeacher:
		userType = chartdbapi.UserType_USER_TYPE_TEACHER
	case model.UserTypeStudent:
		userType = chartdbapi.UserType_USER_TYPE_STUDENT
	case model.UserTypeGuest:
		userType = chartdbapi.UserType_USER_TYPE_GUEST
	}

	return &chartdbapi.User{
		Id:        user.ID.String(),
		Login:     user.Login,
		Type:      userType,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
