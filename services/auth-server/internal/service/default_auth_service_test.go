package service

import (
	"context"
	"errors"
	"sso/internal/domain"
	"sso/internal/dto"
	"sso/internal/lib/logger/handlers/slogdiscard"
	"sso/internal/service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type authServiceTestSuite struct {
	ctx          context.Context
	mockSaver    *mocks.UserSaver
	mockFinder   *mocks.UserFinder
	mockEncoder  *mocks.PasswordEncoder
	mockTokenGen *mocks.TokenGenerator
	service      *defaultAuthService
}

func setup(t *testing.T) *authServiceTestSuite {
	t.Helper()

	mockSaver := new(mocks.UserSaver)
	mockFinder := new(mocks.UserFinder)
	mockEncoder := new(mocks.PasswordEncoder)
	mockTokenGen := new(mocks.TokenGenerator)

	logger := slogdiscard.NewDiscardLogger()

	service := NewDefaultAuthService(
		logger,
		mockSaver,
		mockFinder,
		mockEncoder,
		mockTokenGen,
	)

	return &authServiceTestSuite{
		ctx:          context.Background(),
		mockSaver:    mockSaver,
		mockFinder:   mockFinder,
		mockEncoder:  mockEncoder,
		mockTokenGen: mockTokenGen,
		service:      service,
	}
}

func TestRegister_Success(t *testing.T) {
	s := setup(t)

	email := "test@mail.com"
	password := "password"
	hashed := []byte("hashed_password")
	registerRequest := dto.NewRegisterUserRequest(email, password)

	s.mockFinder.
		On("ExistsByEmail", s.ctx, email).
		Return(false, nil)

	s.mockEncoder.
		On("EncodePassword", password).
		Return(hashed, nil)

	s.mockSaver.
		On("SaveUser", s.ctx, mock.AnythingOfType("domain.User")).
		Return(domain.User{
			ID:           1,
			Email:        email,
			PasswordHash: string(hashed),
			Role:         domain.RoleUser,
		}, nil)

	registerResponse, err := s.service.Register(s.ctx, registerRequest)

	require.NoError(t, err)
	assert.Equal(t, int64(1), registerResponse.ID)

	s.mockFinder.AssertExpectations(t)
	s.mockEncoder.AssertExpectations(t)
	s.mockSaver.AssertExpectations(t)
}

func TestRegister_Failed_EmailAlreadyExists(t *testing.T) {
	s := setup(t)

	email := "test@mail.com"
	password := "password"
	registerRequest := dto.NewRegisterUserRequest(email, password)

	s.mockFinder.
		On("ExistsByEmail", s.ctx, email).
		Return(true, nil)

	savedUser, err := s.service.Register(s.ctx, registerRequest)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrEmailAlreadyExists)
	assert.ErrorContains(t, err, "email already exists")
	assert.Empty(t, savedUser)
	s.mockEncoder.AssertNotCalled(t, "EncodePassword", mock.Anything)
	s.mockSaver.AssertNotCalled(t, "SaveUser", mock.Anything)
}

func TestRegister_Failed_DbErrorOnSave(t *testing.T) {
	s := setup(t)

	dbErr := errors.New("insert failed")
	email := "test@mail.com"
	password := "password"
	hashed := []byte(password)
	registerRequest := dto.NewRegisterUserRequest(email, password)

	userToSave := domain.User{
		Email:        email,
		PasswordHash: string(hashed),
		Role:         "user",
	}

	s.mockFinder.
		On("ExistsByEmail", s.ctx, email).
		Return(false, nil)

	s.mockEncoder.
		On("EncodePassword", password).
		Return(hashed, nil)

	s.mockSaver.
		On("SaveUser", s.ctx, userToSave).
		Return(domain.User{}, dbErr)

	savedUser, err := s.service.Register(s.ctx, registerRequest)
	require.Error(t, err)
	require.ErrorIs(t, err, dbErr)
	assert.Empty(t, savedUser)
}
