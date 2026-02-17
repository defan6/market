package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrUserNotFound       = errors.New("User not found")
	ErrEmailAlreadyExists = errors.New("Email already exists")
)

var (
	queryExistsByEmail = `SELECT EXISTS 
(SELECT 1 FROM users WHERE email = $1)
`
	queryFindByEmail = `SELECT * 
FROM users WHERE email = $1
`
	queryFindById = `SELECT * 
FROM users WHERE id = $1
`

	queryInsertUser = `INSERT INTO users
(email, password, role) VALUES ($1, $2, $3) RETURNING *
`
)

type Storage struct {
	log *slog.Logger
	db  *sqlx.DB
}

func NewStorage(db *sqlx.DB, log *slog.Logger) *Storage {
	return &Storage{
		db:  db,
		log: log,
	}
}

func (s *Storage) FindUserByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	user := domain.User{}
	err := s.db.GetContext(ctx, &user, queryFindByEmail, email)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *Storage) ExistsByEmail(
	ctx context.Context,
	email string,
) (bool, error) {
	var exists bool
	if err := s.db.QueryRowContext(ctx, queryExistsByEmail,
		email).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Storage) FindUserByID(
	ctx context.Context,
	userID int64,
) (domain.User, error) {
	user := domain.User{}
	err := s.db.GetContext(ctx, &user, queryFindById, userID)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *Storage) SaveUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	savedUser := domain.User{}
	err := s.db.QueryRowxContext(ctx,
		queryInsertUser,
		user.Email,
		user.PasswordHash,
		user.Role).
		StructScan(&savedUser)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.User{}, ErrEmailAlreadyExists
		}
		return domain.User{}, err
	}
	return savedUser, nil
}

func (s *Storage) GetListUsers(ctx context.Context, filters map[string]string) ([]domain.User, error) {
	var users []domain.User

	query := "SELECT id, email, role FROM users"
	args := []interface{}{}
	where := []string{}
	i := 1
	for field, value := range filters {
		where = append(where, fmt.Sprintf("%s = $%d", field, i))
		args = append(args, value)
		i++
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	err := s.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, err
	}
	return users, nil
}
