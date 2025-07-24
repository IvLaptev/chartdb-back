package user

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/IvLaptev/chartdb-back/internal/auth"
	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/storage"
	"github.com/IvLaptev/chartdb-back/pkg/ctxlog"
	"github.com/IvLaptev/chartdb-back/pkg/emailsender"
	xerrors "github.com/IvLaptev/chartdb-back/pkg/errors"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
	"github.com/IvLaptev/chartdb-back/pkg/utils/ptr"
	"github.com/go-playground/validator/v10"
)

const (
	userIDLength             int64 = 20
	userConfirmationIDLength int64 = 40

	tokenExpirationTime = time.Hour * 24
)

var (
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidLogin      = errors.New("invalid login")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrConfirmationCodeExpired  = errors.New("confirmation code expired")
	ErrConfirmationCodeNotFound = errors.New("confirmation code not found")

	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type Service interface {
	CreateUser(ctx context.Context, params *CreateUserParams) (*model.User, error)

	LoginUser(ctx context.Context, params *LoginUserParams) (*model.UserToken, error)
	ConfirmUser(ctx context.Context, params *ConfirmUserParams) (*model.User, error)
	Authenticate(ctx context.Context, token string) (context.Context, error)
}

type ServiceImpl struct {
	Logger               *slog.Logger
	Storage              storage.Storage
	UserConfirmationTime time.Duration
	EmailSender          emailsender.EmailSender

	tokenSecret []byte
}

type CreateUserParams struct {
	Login        string `validate:"email,endswith=mirea.ru"`
	PasswordHash *string
}

func (s *ServiceImpl) CreateUser(ctx context.Context, params *CreateUserParams) (*model.User, error) {
	ctxlog.Info(ctx, s.Logger, "create user", slog.Any("params", params))

	var userType model.UserType
	if params.PasswordHash == nil {
		userType = model.UserTypeGuest
	} else {
		userType = model.UserTypeStudent
	}

	userList, err := s.Storage.User().GetAllUsers(ctx, []*model.FilterTerm{
		{
			Key:       model.TermKeyLogin,
			Value:     params.Login,
			Operation: model.FilterOperationExact,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}

	// Don't register guest user on every login
	if len(userList) == 1 && userType == model.UserTypeGuest {
		return userList[0], nil
	}
	// Registered user can't be registered twice
	if len(userList) > 0 {
		return nil, xerrors.WrapInvalidArgument(ErrUserAlreadyExists)
	}

	var confirmedAt *time.Time
	switch userType {
	case model.UserTypeGuest:
		confirmedAt = ptr.To(time.Now())
	case model.UserTypeStudent:
		err := validator.New().StructPartial(params, "Login")
		if err != nil {
			return nil, xerrors.WrapInvalidArgument(ErrInvalidLogin)
		}
	}

	userID, err := utils.GenerateID(userIDLength)
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	userConfirmationID, err := utils.GenerateID(userConfirmationIDLength)
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	var userModel *model.User
	err = s.Storage.DoInTransaction(ctx, func(ctx context.Context) error {
		userModel, err = s.Storage.User().CreateUser(ctx, &storage.CreateUserParams{
			ID:           model.UserID(userID),
			Login:        params.Login,
			PasswordHash: params.PasswordHash,
			Type:         userType,
			ConfirmedAt:  confirmedAt,
		})
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		userConfirmationModel, err := s.Storage.UserConfirmation().CreateUserConfirmation(ctx, &storage.CreateUserConfirmationParams{
			ID:       model.UserConfirmationID(userConfirmationID),
			UserID:   model.UserID(userID),
			Duration: s.UserConfirmationTime,
		})
		if err != nil {
			return fmt.Errorf("create user confirmation: %w", err)
		}

		err = s.EmailSender.SendCreateUserEmail(userModel.Login, userConfirmationModel.ID.String())
		if err != nil {
			return fmt.Errorf("send create user email: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't create user: %w", err)
	}

	return userModel, nil
}

type LoginUserParams struct {
	Login        string
	PasswordHash string
}

func (s *ServiceImpl) LoginUser(ctx context.Context, params *LoginUserParams) (*model.UserToken, error) {
	ctxlog.Info(ctx, s.Logger, "login user", slog.Any("params", params))

	userList, err := s.Storage.User().GetAllUsers(ctx, []*model.FilterTerm{
		{
			Key:       model.TermKeyLogin,
			Value:     params.Login,
			Operation: model.FilterOperationExact,
		},
		{
			Key:       model.TermKeyPasswordHash,
			Value:     params.PasswordHash,
			Operation: model.FilterOperationExact,
		},
		{
			Key:       model.TermKeyConfirmedAt,
			Value:     nil,
			Operation: model.FilterOperationNotEqual,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}

	if len(userList) == 0 {
		return nil, xerrors.WrapNotFound(ErrUserNotFound)
	}

	userToken, err := createToken(userList[0], s.tokenSecret)
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	return &model.UserToken{
		Value:  userToken,
		UserID: userList[0].ID,
	}, nil
}

type ConfirmUserParams struct {
	UserConfirmationID model.UserConfirmationID
}

func (s *ServiceImpl) ConfirmUser(ctx context.Context, params *ConfirmUserParams) (*model.User, error) {
	ctxlog.Info(ctx, s.Logger, "confirm user", slog.Any("params", params))

	now := time.Now()
	userConfirmation, err := s.Storage.UserConfirmation().GetUserConfirmationByID(ctx, params.UserConfirmationID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, xerrors.WrapNotFound(ErrConfirmationCodeNotFound)
		}

		return nil, fmt.Errorf("get user confirmation by id: %w", err)
	}

	if userConfirmation.ExpiresAt.Before(now) {
		return nil, xerrors.WrapInvalidArgument(ErrConfirmationCodeExpired)
	}

	userModel, err := s.Storage.User().PatchUser(ctx, &storage.PatchUserParams{
		ID:          userConfirmation.UserID,
		ConfirmedAt: utils.NewOptional(&now),
	})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, xerrors.WrapNotFound(ErrUserNotFound)
		}
		return nil, fmt.Errorf("patch user: %w", err)
	}

	return userModel, nil
}

func (s *ServiceImpl) Authenticate(ctx context.Context, token string) (context.Context, error) {
	ctxlog.Info(ctx, s.Logger, "authenticate user")

	var userModel *model.User
	tokenParts := strings.Split(token, " ")
	switch len(tokenParts) {
	case 1:
		userID, err := base64.StdEncoding.DecodeString(tokenParts[0])
		if err != nil {
			return nil, xerrors.WrapUnauthenticated(ErrInvalidUserID)
		}

		userList, err := s.Storage.User().GetAllUsers(ctx, []*model.FilterTerm{
			{
				Key:       model.TermKeyID,
				Value:     string(userID),
				Operation: model.FilterOperationExact,
			},
			{
				Key:       model.TermKeyType,
				Value:     model.UserTypeGuest.String(),
				Operation: model.FilterOperationExact,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("get all users: %w", err)
		}

		if len(userList) == 0 {
			return nil, xerrors.WrapUnauthenticated(ErrInvalidToken)
		}

		userModel = userList[0]
	case 2:
		if tokenParts[0] != "Bearer" {
			return nil, xerrors.WrapUnauthenticated(ErrInvalidToken)
		}

		userToken, err := parseToken(tokenParts[1], s.tokenSecret)
		if err != nil {
			return nil, xerrors.WrapUnauthenticated(ErrInvalidToken)
		}

		if userToken.ExpiresAt.Before(time.Now()) {
			return nil, xerrors.WrapUnauthenticated(ErrTokenExpired)
		}

		userModel, err = s.Storage.User().GetUserByID(ctx, userToken.User.ID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, xerrors.WrapUnauthenticated(ErrInvalidToken)
			}
			return nil, fmt.Errorf("get user by id: %w", err)
		}
	default:
		return nil, xerrors.WrapUnauthenticated(ErrInvalidToken)
	}

	ctx = auth.SetSubject(ctx, &auth.Subject{
		UserID:   userModel.ID,
		UserType: userModel.Type,
	})

	return ctx, nil
}

func NewService(
	logger *slog.Logger,
	storage storage.Storage,
	emailSender emailsender.EmailSender,
	userConfirmationTime time.Duration,
	tokenSecret []byte,
) Service {
	return &ServiceImpl{
		Logger:               logger,
		Storage:              storage,
		EmailSender:          emailSender,
		UserConfirmationTime: userConfirmationTime,
		tokenSecret:          tokenSecret,
	}
}

type token struct {
	User      *model.User `json:"user_id"`
	ExpiresAt time.Time   `json:"expires_at"`
}

func createToken(user *model.User, tokenSecret []byte) (string, error) {
	token := &token{
		User:      user,
		ExpiresAt: time.Now().Add(tokenExpirationTime),
	}

	jsonToken, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	tokenString, err := utils.AES128Encrypt(string(jsonToken), tokenSecret)
	if err != nil {
		return "", fmt.Errorf("encrypt: %w", err)
	}

	return tokenString, nil
}

func parseToken(tokenString string, tokenSecret []byte) (*token, error) {
	jsonToken, err := utils.AES128Decrypt(tokenString, tokenSecret)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	var token token
	err = json.Unmarshal([]byte(jsonToken), &token)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &token, nil
}
