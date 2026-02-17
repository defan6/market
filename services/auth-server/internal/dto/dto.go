package dto

type TokenGenerateRequest struct {
	UserID int64
	Email  string
	AppID  int64
}

func NewTokenGenerateRequest(userID int64, email string, appID int64) *TokenGenerateRequest {
	return &TokenGenerateRequest{
		UserID: userID,
		Email:  email,
		AppID:  appID,
	}
}

type TokenGenerateResponse struct {
	Token string
}

func NewTokenGenerateResponse(token string) *TokenGenerateResponse {
	return &TokenGenerateResponse{
		Token: token,
	}
}

type ListUserRequest struct {
	Filters map[string]string
}

func NewListUserRequest(filters map[string]string) *ListUserRequest {
	return &ListUserRequest{
		Filters: filters,
	}
}

type ListUserResponse struct {
	Users []*UserResponse
}

func NewListUserResponse(users []*UserResponse) *ListUserResponse {
	return &ListUserResponse{
		Users: users,
	}
}

type RegisterUserRequest struct {
	Email    string
	Password string
}

func NewRegisterUserRequest(email, password string) *RegisterUserRequest {
	return &RegisterUserRequest{
		Email:    email,
		Password: password,
	}
}

type UserResponse struct {
	ID    int64
	Email string
	Role  string
}

func NewUserResponse(id int64, email string, role string) *UserResponse {
	return &UserResponse{
		ID:    id,
		Email: email,
		Role:  role,
	}
}

type LoginUserResponse struct {
	Token string
}

func NewLoginUserResponse(token string) *LoginUserResponse {
	return &LoginUserResponse{
		Token: token,
	}
}

type LoginUserRequest struct {
	Email    string
	Password string
	AppID    int
}

func NewLoginUserRequest(email, password string, appID int) *LoginUserRequest {
	return &LoginUserRequest{
		Email:    email,
		Password: password,
		AppID:    appID,
	}
}

type IsAdminRequest struct {
	ID int64
}

func NewIsAdminRequest(id int64) *IsAdminRequest {
	return &IsAdminRequest{
		ID: id,
	}
}

type IsAdminResponse struct {
	IsAdmin bool
}

func NewIsAdminResponse(isAdmin bool) *IsAdminResponse {
	return &IsAdminResponse{
		IsAdmin: isAdmin,
	}
}

type RegisterUserResponse struct {
	ID int64
}

func NewRegisterUserResponse(id int64) *RegisterUserResponse {
	return &RegisterUserResponse{
		ID: id,
	}
}
