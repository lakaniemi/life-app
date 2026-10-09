package db

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/lakaniemi/life-app/api/internal/dbtest"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()
	q := New(dbtest.New(t))

	user, err := q.CreateUser(t.Context(), CreateUserParams{GoogleSub: "sub-1", Name: "Alice"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.GoogleSub != "sub-1" || user.Name != "Alice" {
		t.Errorf("user = %+v, want GoogleSub sub-1 and Name Alice", user)
	}
	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero, want it set by the database")
	}

	_, err = q.CreateUser(t.Context(), CreateUserParams{GoogleSub: "sub-1", Name: "Someone else"})
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.UniqueViolation {
		t.Errorf("duplicate google_sub: err = %v, want unique violation", err)
	}
}

func TestGetUserByGoogleSub(t *testing.T) {
	t.Parallel()
	q := New(dbtest.New(t))

	created, err := q.CreateUser(t.Context(), CreateUserParams{GoogleSub: "sub-1", Name: "Alice"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	tests := []struct {
		name      string
		googleSub string
		wantErr   error
	}{
		{name: "existing user", googleSub: "sub-1"},
		{name: "unknown sub", googleSub: "sub-2", wantErr: pgx.ErrNoRows},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := q.GetUserByGoogleSub(t.Context(), tt.googleSub)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got.ID != created.ID {
				t.Errorf("ID = %v, want %v", got.ID, created.ID)
			}
		})
	}
}
