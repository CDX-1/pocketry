package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/CDX-1/pocketry/internal/db"
)

type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	if queries == nil {
		panic("admin: database queries are nil")
	}

	return &Service{
		queries: queries,
	}
}

func (s *Service) ListUsers(ctx context.Context, limit int64) ([]db.ListUsersRow, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	users, err := s.queries.ListUsers(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (s *Service) GetUser(ctx context.Context, userID int64) (db.User, error) {
	if userID <= 0 {
		return db.User{}, errors.New("user ID must be positive")
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return db.User{}, fmt.Errorf("user %d does not exist", userID)
	}
	if err != nil {
		return db.User{}, fmt.Errorf("get user %d: %w", userID, err)
	}

	return db.User{
		ID:                  user.ID,
		Username:            user.Username,
		UsernameNormalized:  user.UsernameNormalized,
		CryptoPolicyVersion: user.CryptoPolicyVersion,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	}, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("user ID must be positive")
	}

	err := s.queries.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("delete user %d: %w", userID, err)
	}

	return nil
}