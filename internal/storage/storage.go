package storage

import (
	"context"
	"errors"
	"log/slog"
	"sso/internal/domain"

	"github.com/jmoiron/sqlx"
)

var (
	UserNotFoundErr = errors.New("User not found")
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
		return domain.User{}, UserNotFoundErr
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
		return domain.User{}, UserNotFoundErr
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
		return domain.User{}, err
	}
	return savedUser, nil
}
