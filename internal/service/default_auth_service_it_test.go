package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sso/internal/dto"
	"sso/internal/grpc/auth"
	"sso/internal/lib/logger/handlers/slogdiscard"
	"sso/internal/lib/security/encoder"
	"sso/internal/service/mocks"
	"sso/internal/storage"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	suite.Suite
	db      *sqlx.DB
	service auth.AuthService
}

func TestIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)

	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		require.NoError(s.T(), pgContainer.Terminate(ctx))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(s.T(), err)

	s.db, err = sqlx.Connect("postgres", connStr)
	require.NoError(s.T(), err)

	migrationsPath, err := filepath.Abs("../../cmd/db/migrations")
	if err != nil {
		wd, _ := os.Getwd()
		fmt.Println("Current working directory:", wd)
	}
	require.NoError(s.T(), err)
	migrationUrl := "file://" + filepath.ToSlash(migrationsPath)

	m, err := migrate.New(
		migrationUrl,
		connStr,
	)

	require.NoError(s.T(), err)
	err = m.Up()
	require.NoError(s.T(), err)

	logger := slogdiscard.NewDiscardLogger()
	storage := storage.NewStorage(s.db, logger)
	passwordEncoder := encoder.NewPasswordEncoder()
	mockTokenGen := new(mocks.TokenGenerator)

	s.service = NewDefaultAuthService(
		logger,
		storage,
		storage,
		passwordEncoder,
		mockTokenGen,
	)
}

func (s *IntegrationTestSuite) SetupTest() {
	_, err := s.db.Exec("TRUNCATE TABLE users RESTART IDENTITY ")
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) TestRegister_Success() {
	t := s.T()
	ctx := context.Background()
	email := "test@mail.com"
	password := "password"

	registerRequest := dto.NewRegisterUserRequest(email, password)
	registerResponse, err := s.service.Register(ctx, registerRequest)

	require.NoError(t, err)
	assert.Greater(t, registerResponse.ID, int64(0))

	var dbEmail string
	var dbPassword string
	err = s.db.QueryRow("SELECT email, password FROM users WHERE id = $1", registerResponse.ID).
		Scan(&dbEmail, &dbPassword)
	require.NoError(t, err)
	assert.Equal(t, email, dbEmail)
	assert.NotEmpty(t, dbPassword)

	passwordEncoder := encoder.NewPasswordEncoder()
	match, err := passwordEncoder.ComparePassword(password, dbPassword)
	assert.True(t, match, "password should match the hash")
}

func (s *IntegrationTestSuite) TestRegister_Failed_EmailAlreadyExists() {
	t := s.T()
	ctx := context.Background()
	firstEmail := "test@mail.com"
	firstPassword := "password"

	secondEmail := "test@mail.com"
	secondPassword := "password"

	firstRegisterRequest := dto.NewRegisterUserRequest(firstEmail, firstPassword)
	secondRegisterRequest := dto.NewRegisterUserRequest(secondEmail, secondPassword)

	_, err := s.service.Register(ctx, firstRegisterRequest)
	require.NoError(t, err)

	_, err = s.service.Register(ctx, secondRegisterRequest)
	require.Error(t, err, ErrEmailAlreadyExists)

}

func (s *IntegrationTestSuite) TestRegister_Failed_DatabaseConnectionLost() {
	t := s.T()

	ctx := context.Background()
	email := "test@mail.com"
	password := "password"
	registerRequest := dto.NewRegisterUserRequest(email, password)
	s.db.Close()

	_, err := s.service.Register(ctx, registerRequest)
	require.Error(t, err)
}
