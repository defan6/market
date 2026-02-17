package auth

import (
	"context"
	"sso/internal/dto"

	ssov1 "github.com/defan6/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Login(ctx context.Context, loginRequest *dto.LoginUserRequest) (loginResponse *dto.LoginUserResponse, err error)
	Register(ctx context.Context, registerRequest *dto.RegisterUserRequest) (registerResponse *dto.RegisterUserResponse, err error)
	IsAdmin(ctx context.Context, isAdminRequest *dto.IsAdminRequest) (isAdminResponse *dto.IsAdminResponse, err error)
}

type UserService interface {
	ListUsers(context.Context, *dto.ListUserRequest) (*dto.ListUserResponse, error)
}
type serverAPI struct {
	ssov1.UnimplementedAuthServer
	authService AuthService
	userService UserService
}

func Register(gRPC *grpc.Server, authService AuthService, userService UserService) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{authService: authService, userService: userService})
}

func (s *serverAPI) ListUsers(
	ctx context.Context,
	req *ssov1.ListUserRequest,
) (*ssov1.ListUserResponse, error) {

	listUserRequest := dto.NewListUserRequest(req.GetFilters())
	listUsersRes, err := s.userService.ListUsers(ctx, listUserRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}
	response := mapToGRPCListUserResponse(listUsersRes)
	return response, nil

}

func mapToGRPCListUserResponse(res *dto.ListUserResponse) *ssov1.ListUserResponse {
	var list []*ssov1.User
	for _, user := range res.Users {
		userRes := &ssov1.User{
			Id:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		}
		list = append(list, userRes)
	}

	return &ssov1.ListUserResponse{Users: list}
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}
	loginReq := dto.NewLoginUserRequest(req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	loginResponse, err := s.authService.Login(ctx, loginReq)
	if err != nil {
		// TODO ...
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &ssov1.LoginResponse{
		Token: loginResponse.Token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}
	registerRequest := dto.NewRegisterUserRequest(req.GetEmail(), req.GetPassword())
	registerResponse, err := s.authService.Register(ctx, registerRequest)
	if err != nil {
		// TODO ...
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &ssov1.RegisterResponse{UserId: registerResponse.ID}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}
	isAdminRequest := dto.NewIsAdminRequest(req.UserId)
	isAdminResponse, err := s.authService.IsAdmin(ctx, isAdminRequest)
	if err != nil {
		// TODO ...
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &ssov1.IsAdminResponse{IsAdmin: isAdminResponse.IsAdmin}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if req.GetAppId() == 0 {
		return status.Error(codes.InvalidArgument, "appId is required")
	}
	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() <= 0 {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	return nil
}
