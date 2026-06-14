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

// Repository is the small data-access surface the service depends on.
// Declaring it here (the consumer) is idiomatic Go and lets tests pass a mock.
// The concrete implementation lives in the repository package.
type Repository interface {
	CreateUser(ctx context.Context, name string, dob time.Time) (sqlc.User, error)
	GetUser(ctx context.Context, id int32) (sqlc.User, error)
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

// GetUser fetches a user and returns it together with the age computed
// dynamically from the date of birth.
func (s *Service) GetUser(ctx context.Context, id int32) (sqlc.User, int, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, 0, ErrNotFound
		}
		return sqlc.User{}, 0, err
	}

	age := CalculateAge(user.Dob, time.Now())
	return user, age, nil
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
