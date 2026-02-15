package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain"
	"sso/internal/dto"
	"sso/internal/grpc/auth"
	"sso/internal/storage"
	"strconv"
)

var (
	InvalidCredentialsErr = errors.New("invalid credentials")
	UserAlreadyExistsErr  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type DefaultAuthService struct {
	log             *slog.Logger
	userSaver       UserSaver
	userFinder      UserFinder
	passwordEncoder PasswordEncoder
	tokenGenerator  TokenGenerator
}

func NewDefaultAuthService(log *slog.Logger,
	userSaver UserSaver,
	userFinder UserFinder,
	passwordEncoder PasswordEncoder,
	tokenGenerator TokenGenerator,
) auth.AuthService {
	return &DefaultAuthService{
		log:             log,
		userSaver:       userSaver,
		userFinder:      userFinder,
		passwordEncoder: passwordEncoder,
		tokenGenerator:  tokenGenerator,
	}
}

type UserSaver interface {
	SaveUser(ctx context.Context, user domain.User) (domain.User, error)
}

type UserFinder interface {
	FindUserByEmail(ctx context.Context, email string) (domain.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindUserByID(ctx context.Context, userID int64) (domain.User, error)
}

type TokenGenerator interface {
	GenerateToken(
		ctx context.Context,
		userDetails domain.UserDetails,
		appID string,
	) (*dto.TokenGenerateResponse, error)
}

type PasswordEncoder interface {
	EncodePassword(password string) ([]byte, error)
	ComparePassword(password, hash string) (bool, error)
}

func (a *DefaultAuthService) Register(
	ctx context.Context,
	email string,
	password string,
) (userID int64, err error) {
	exists, err := a.userFinder.ExistsByEmail(ctx, email)
	if err != nil {
		return 0, fmt.Errorf("Error checking if user exists: %w", err)
	}
	if exists {
		return 0, UserAlreadyExistsErr
	}

	passwordHash, err := a.passwordEncoder.EncodePassword(password)
	if err != nil {
		return 0, fmt.Errorf("Error encoding password: %w", err)
	}
	user := domain.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         domain.RoleUser,
	}

	savedUser, err := a.userSaver.SaveUser(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("Error saving user: %w", err)
	}
	return savedUser.ID, nil
}

func (a *DefaultAuthService) IsAdmin(
	ctx context.Context,
	userID int64,
) (bool, error) {
	res, err := a.userFinder.FindUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("Error checking if user exists: %w", err)
	}
	return res.Role == domain.RoleAdmin, nil
}

func (a *DefaultAuthService) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (token string, err error) {
	findUserRes, err := a.userFinder.FindUserByEmail(ctx, email)
	if err != nil && errors.Is(err, storage.ErrUserNotFound) {
		return "", InvalidCredentialsErr
	}
	if err != nil {
		return "", fmt.Errorf("error finding user by email: %w", err)
	}

	isValidPassword, err := a.passwordEncoder.ComparePassword(password, findUserRes.PasswordHash)
	if err != nil || !isValidPassword {
		return "", InvalidCredentialsErr
	}
	details := domain.NewUserDetails(findUserRes.ID, findUserRes.Email, findUserRes.Role)
	genTokenRes, err := a.tokenGenerator.GenerateToken(ctx, details, strconv.Itoa(appID))
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}
	return genTokenRes.Token, nil
}
