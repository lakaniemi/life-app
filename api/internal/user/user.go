// Package user reads and changes user profiles. Users are created on first
// sign-in by internal/auth.
package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// Service reads and changes user profiles.
type Service struct {
	queries *db.Queries
}

// NewService returns a Service that stores users through queries.
func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}

// Get returns the user.
func (s *Service) Get(ctx context.Context, userID uuid.UUID) (db.User, error) {
	found, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return db.User{}, fmt.Errorf("get user: %w", err)
	}
	return found, nil
}

// Rename sets the user's name and returns the updated user.
func (s *Service) Rename(ctx context.Context, userID uuid.UUID, name string) (db.User, error) {
	renamed, err := s.queries.UpdateUserName(ctx, db.UpdateUserNameParams{ID: userID, Name: name})
	if err != nil {
		return db.User{}, fmt.Errorf("rename user: %w", err)
	}
	return renamed, nil
}
