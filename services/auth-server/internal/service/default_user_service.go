package service

import (
	"context"
	"errors"
	"log/slog"
	"sso/internal/domain"
	"sso/internal/dto"
)

var ErrInvalidFilters = errors.New("invalid filters")

type defaultUserService struct {
	log    *slog.Logger
	storer UserStorer
}

type UserStorer interface {
	GetListUsers(ctx context.Context, filters map[string]string) ([]domain.User, error)
}

func NewDefaultUserService(log *slog.Logger, userStorer UserStorer) *defaultUserService {
	return &defaultUserService{
		log:    log,
		storer: userStorer,
	}
}

func (s *defaultUserService) ListUsers(ctx context.Context, request *dto.ListUserRequest) (*dto.ListUserResponse, error) {
	filters, err := getFilters(request)
	if err != nil {
		return nil, err
	}
	listUsers, err := s.storer.GetListUsers(ctx, filters)
	if err != nil {
		return nil, err
	}
	listResponse := mapToListUserResponse(listUsers)
	return listResponse, nil
}

func mapToListUserResponse(users []domain.User) *dto.ListUserResponse {
	list := make([]*dto.UserResponse, 0, len(users))
	for _, u := range users {
		list = append(list, mapToUserResponse(u))
	}

	return dto.NewListUserResponse(list)
}

func mapToUserResponse(user domain.User) *dto.UserResponse {
	return dto.NewUserResponse(user.ID, user.Email, user.Role)
}

func getFilters(request *dto.ListUserRequest) (map[string]string, error) {
	filters := make(map[string]string)

	for k, v := range request.Filters {
		if !isValidFilter(k, v) {
			return nil, ErrInvalidFilters
		}
		filters[k] = v
	}
	return filters, nil
}

func isValidFilter(key, value string) bool {
	return true
}
