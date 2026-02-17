package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain"
	"sso/internal/dto"
	"sso/internal/storage"
	"strconv"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type defaultAuthService struct {
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
) *defaultAuthService {
	return &defaultAuthService{
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

func (a *defaultAuthService) Register(
	ctx context.Context,
	registerRequest *dto.RegisterUserRequest,
) (*dto.RegisterUserResponse, error) {
	exists, err := a.userFinder.ExistsByEmail(ctx, registerRequest.Email)
	if err != nil {
		return &dto.RegisterUserResponse{}, fmt.Errorf("Error checking if user exists: %w", err)
	}
	if exists {
		return &dto.RegisterUserResponse{}, ErrEmailAlreadyExists
	}

	passwordHash, err := a.passwordEncoder.EncodePassword(registerRequest.Password)
	if err != nil {
		return &dto.RegisterUserResponse{}, fmt.Errorf("Error encoding password: %w", err)
	}
	user := domain.User{
		Email:        registerRequest.Email,
		PasswordHash: string(passwordHash),
		Role:         domain.RoleUser,
	}

	savedUser, err := a.userSaver.SaveUser(ctx, user)
	if err != nil {
		return &dto.RegisterUserResponse{}, fmt.Errorf("Error saving user: %w", err)
	}
	registerResponse := dto.NewRegisterUserResponse(savedUser.ID)
	return registerResponse, nil
}

func (a *defaultAuthService) IsAdmin(
	ctx context.Context,
	isAdminRequest *dto.IsAdminRequest,
) (*dto.IsAdminResponse, error) {
	res, err := a.userFinder.FindUserByID(ctx, isAdminRequest.ID)
	if err != nil {
		return &dto.IsAdminResponse{}, fmt.Errorf("Error checking if user exists: %w", err)
	}
	isAdmin := res.Role == domain.RoleAdmin
	isAdminResponse := dto.NewIsAdminResponse(isAdmin)
	return isAdminResponse, nil
}

func (a *defaultAuthService) Login(
	ctx context.Context,
	loginRequest *dto.LoginUserRequest,
) (*dto.LoginUserResponse, error) {
	findUserRes, err := a.userFinder.FindUserByEmail(ctx, loginRequest.Email)
	if err != nil && errors.Is(err, storage.ErrUserNotFound) {
		return &dto.LoginUserResponse{}, ErrInvalidCredentials
	}
	if err != nil {
		return &dto.LoginUserResponse{}, fmt.Errorf("error finding user by email: %w", err)
	}

	isValidPassword, err := a.passwordEncoder.ComparePassword(loginRequest.Password, findUserRes.PasswordHash)
	if err != nil || !isValidPassword {
		return &dto.LoginUserResponse{}, ErrInvalidCredentials
	}
	details := domain.NewUserDetails(findUserRes.ID, findUserRes.Email, findUserRes.Role)
	genTokenRes, err := a.tokenGenerator.GenerateToken(ctx, details, strconv.Itoa(loginRequest.AppID))
	loginResponse := dto.NewLoginUserResponse(genTokenRes.Token)
	if err != nil {
		return &dto.LoginUserResponse{}, fmt.Errorf("error generating token: %w", err)
	}
	return loginResponse, nil
}
