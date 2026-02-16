package user

import (
	"context"
	"encoding/base64"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/IvLaptev/chartdb-back/internal/auth"
	"github.com/IvLaptev/chartdb-back/internal/model"
	"github.com/IvLaptev/chartdb-back/internal/storage"
	"github.com/IvLaptev/chartdb-back/internal/storage/postgres"
	"github.com/IvLaptev/chartdb-back/internal/tests"
	"github.com/IvLaptev/chartdb-back/pkg/emailsender"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
	"github.com/IvLaptev/chartdb-back/pkg/utils/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	gomock "go.uber.org/mock/gomock"
)

type UserServiceSuite struct {
	suite.Suite

	UserService *ServiceImpl
	storage     storage.Storage
	logger      *slog.Logger
	emailsender *emailsender.MockEmailSender
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

func (s *UserServiceSuite) SetupSuite() {
	var err error
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	s.storage, err = postgres.NewStorage(*tests.NewPostgresTestConfig(), s.logger)
	assert.NoError(s.T(), err)
	s.emailsender = emailsender.NewMockEmailSender(gomock.NewController(s.T()))
}

func (s *UserServiceSuite) SetupTest() {
	s.storage.Erase(context.Background())
	s.UserService = NewService(s.logger, s.storage, s.emailsender, 30*time.Minute, 1*time.Hour, []byte("secret"))
}

func (s *UserServiceSuite) TestCreateUser_GuestOk() {
	ctx := context.Background()

	userLogin := "00И0000"
	userModel, err := s.UserService.CreateUser(ctx, &CreateUserParams{
		Login:        userLogin,
		PasswordHash: utils.NewSecret[*string](nil),
	})
	s.Require().NoError(err)

	userModel, err = s.storage.User().GetUserByID(ctx, userModel.ID)
	s.Require().NoError(err)
	s.Require().Equal("00И0000", userModel.Login)
	s.Require().Equal(model.UserTypeGuest, userModel.Type)
	s.Require().Nil(userModel.PasswordHash.Value)

	confirmationList, err := s.storage.UserConfirmation().GetAllUserConfirmation(ctx, nil)
	s.Require().NoError(err)
	s.Require().Empty(confirmationList)
}

func (s *UserServiceSuite) TestCreateUser_StudentWrongEmail() {
	ctx := context.Background()

	userLogin := "00И0000"
	_, err := s.UserService.CreateUser(ctx, &CreateUserParams{
		Login:        userLogin,
		PasswordHash: utils.NewSecret(ptr.To("password")),
	})
	s.Require().Error(err)
}

func (s *UserServiceSuite) TestCreateUser_StudentOk() {
	ctx := context.Background()
	userLogin := "test@edu.mirea.ru"

	s.emailsender.EXPECT().SendCreateUserEmail(userLogin, gomock.Any()).Return(nil)

	userModel, err := s.UserService.CreateUser(ctx, &CreateUserParams{
		Login:        userLogin,
		PasswordHash: utils.NewSecret(ptr.To("password")),
	})
	s.Require().NoError(err)
	s.Require().Equal(model.UserTypeStudent, userModel.Type)
	s.Require().NotNil(userModel.PasswordHash.Value)
	s.Require().Equal(userLogin, userModel.Login)
	s.Require().Nil(userModel.ConfirmedAt)

	confirmationList, err := s.storage.UserConfirmation().GetAllUserConfirmation(ctx, nil)
	s.Require().NoError(err)
	s.Require().Equal(1, len(confirmationList))
}

func (s *UserServiceSuite) TestAuthenticate_GuestOk() {
	ctx := context.Background()

	s.TestCreateUser_GuestOk()

	userLogin := "00И0000"
	token := base64.StdEncoding.EncodeToString([]byte(userLogin))
	ctx, err := s.UserService.Authenticate(ctx, token)
	s.Require().NoError(err)

	subject, err := auth.GetSubject(ctx)
	s.Require().NoError(err)
	s.Require().NotEqual(userLogin, subject.UserID)
}
