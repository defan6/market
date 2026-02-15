package service

import (
	"context"
	"sso/internal/domain"
	"sso/internal/lib/logger/handlers/slogdiscard"
	"sso/internal/service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegister_Success(t *testing.T) {
	ctx := context.Background()

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

	email := "test@mail.com"
	password := "password"
	hashed := []byte("hashed_password")

	mockFinder.
		On("ExistsByEmail", ctx, email).
		Return(false, nil)

	mockEncoder.
		On("EncodePassword", password).
		Return(hashed, nil)

	mockSaver.
		On("SaveUser", ctx, mock.AnythingOfType("domain.User")).
		Return(domain.User{
			ID:           1,
			Email:        email,
			PasswordHash: string(hashed),
			Role:         domain.RoleUser,
		}, nil)

	userID, err := service.Register(ctx, email, password)

	require.NoError(t, err)
	assert.Equal(t, int64(1), userID)

	mockFinder.AssertExpectations(t)
	mockEncoder.AssertExpectations(t)
	mockSaver.AssertExpectations(t)
}
