package storage

import (
	"context"
	"errors"
	"regexp"
	"sso/internal/domain"
	"sso/internal/lib/logger/handlers/slogdiscard"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_Register(t *testing.T) {
	db, mock, err := sqlmock.New()

	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	t.Cleanup(func() {
		assert.NoError(t, mock.ExpectationsWereMet(), "not all sqlmock expectations were met")
	})

	s := NewStorage(sqlxDB, slogdiscard.NewDiscardLogger())

	ctx := context.Background()
	testEmail := "test@example.com"
	testPassHash := "hashed_password"
	testRole := "user"
	expectedUserId := int64(1)
	rows := sqlmock.NewRows([]string{"id", "email", "password", "role"}).
		AddRow(expectedUserId, testEmail, testPassHash, testRole)

	t.Run("success - new user registration", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(queryInsertUser)).
			WithArgs(testEmail, testPassHash, "user").
			WillReturnRows(rows)

		userToSave := domain.User{
			Email:        testEmail,
			PasswordHash: testPassHash,
			Role:         testRole,
		}

		savedUser, err := s.SaveUser(ctx, userToSave)
		require.NoError(t, err)
		assert.Equal(t, expectedUserId, savedUser.ID)
		assert.Equal(t, testEmail, savedUser.Email)
		assert.Equal(t, testPassHash, savedUser.PasswordHash)
		assert.Equal(t, testRole, savedUser.Role)

	})

	t.Run("failed - user email already exists", func(t *testing.T) {
		pqErr := &pq.Error{
			Code:    "23505",
			Message: "duplicate key value violates unique constraint \"users_email_key\"",
		}
		mock.ExpectQuery(regexp.QuoteMeta(queryInsertUser)).
			WithArgs(testEmail, testPassHash, testRole).
			WillReturnError(pqErr)

		userToSave := domain.User{
			Email:        testEmail,
			PasswordHash: testPassHash,
			Role:         testRole,
		}

		savedUser, err := s.SaveUser(ctx, userToSave)

		require.Error(t, err)
		assert.Empty(t, savedUser)
		assert.ErrorContains(t, err, "Email already exists")
	})

	t.Run("failed - database connection lost", func(t *testing.T) {

		dbErr := errors.New("database connection lost")

		mock.ExpectQuery(regexp.QuoteMeta(queryInsertUser)).
			WithArgs(testEmail, testPassHash, testRole).
			WillReturnError(dbErr)

		userToSave := domain.User{
			Email:        testEmail,
			PasswordHash: testPassHash,
			Role:         testRole,
		}

		savedUser, err := s.SaveUser(ctx, userToSave)
		require.Error(t, err)
		assert.Empty(t, savedUser)
		assert.ErrorContains(t, err, "database connection lost")
	})
}
