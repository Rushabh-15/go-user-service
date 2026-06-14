package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"ainyx-user-api/db/sqlc"
)

// Sentinel errors let the handler map failures to HTTP status codes without
// knowing anything about the database or the business rules.
var (
	ErrNotFound  = errors.New("user not found")
	ErrFutureDOB = errors.New("dob cannot be in the future")
)

// UserWithAge is the service's view of a user, carrying the age computed
// dynamically from the date of birth. Handlers map it to the JSON response.
type UserWithAge struct {
	User sqlc.User
	Age  int
}

// Repository is the small data-access surface the service depends on.
// Declaring it here (the consumer) is idiomatic Go and lets tests pass a mock.
// The concrete implementation lives in the repository package.
type Repository interface {
	CreateUser(ctx context.Context, name string, dob time.Time) (sqlc.User, error)
	GetUser(ctx context.Context, id int32) (sqlc.User, error)
	UpdateUser(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error)
	DeleteUser(ctx context.Context, id int32) (int64, error)
	ListUsers(ctx context.Context, limit, offset int32) ([]sqlc.User, error)
}

type Service struct {
	repo Repository
	log  *zap.Logger
}

func New(repo Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// CreateUser applies business rules, then persists the user.
func (s *Service) CreateUser(ctx context.Context, name string, dob time.Time) (sqlc.User, error) {
	if dob.After(time.Now()) {
		return sqlc.User{}, ErrFutureDOB
	}

	user, err := s.repo.CreateUser(ctx, name, dob)
	if err != nil {
		return sqlc.User{}, err
	}

	s.log.Info("user created", zap.Int32("id", user.ID), zap.String("name", user.Name))
	return user, nil
}

// GetUser fetches a user together with the dynamically computed age.
func (s *Service) GetUser(ctx context.Context, id int32) (UserWithAge, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserWithAge{}, ErrNotFound
		}
		return UserWithAge{}, err
	}
	return UserWithAge{User: user, Age: CalculateAge(user.Dob, time.Now())}, nil
}

// UpdateUser applies business rules and replaces the user's fields.
// The UPDATE uses RETURNING, so a missing row surfaces as pgx.ErrNoRows.
func (s *Service) UpdateUser(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error) {
	if dob.After(time.Now()) {
		return sqlc.User{}, ErrFutureDOB
	}

	user, err := s.repo.UpdateUser(ctx, id, name, dob)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, ErrNotFound
		}
		return sqlc.User{}, err
	}

	s.log.Info("user updated", zap.Int32("id", user.ID))
	return user, nil
}

// DeleteUser removes a user, returning ErrNotFound if nothing was deleted.
func (s *Service) DeleteUser(ctx context.Context, id int32) error {
	rows, err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	s.log.Info("user deleted", zap.Int32("id", id))
	return nil
}

// ListUsers returns a page of users, each with its computed age.
func (s *Service) ListUsers(ctx context.Context, limit, offset int32) ([]UserWithAge, error) {
	users, err := s.repo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Compute "now" once so every row on the page uses the same reference.
	now := time.Now()
	result := make([]UserWithAge, 0, len(users))
	for _, u := range users {
		result = append(result, UserWithAge{User: u, Age: CalculateAge(u.Dob, now)})
	}
	return result, nil
}

// CalculateAge returns the number of complete years between dob and now.
// It compares month and day (not day-of-year) so leap years are handled
// correctly: e.g. someone born 2000-03-01 turns 1 on 2001-03-01 even though
// 2000 had a Feb 29 that shifts the day-of-year.
func CalculateAge(dob, now time.Time) int {
	years := now.Year() - dob.Year()
	if now.Month() < dob.Month() ||
		(now.Month() == dob.Month() && now.Day() < dob.Day()) {
		years--
	}
	return years
}
