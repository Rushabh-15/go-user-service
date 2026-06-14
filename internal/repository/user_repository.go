package repository

import (
	"context"
	"time"

	"ainyx-user-api/db/sqlc"
)

// UserRepository is a thin, typed wrapper over the sqlc-generated Queries.
// It hides sqlc's params structs from the service layer.
type UserRepository struct {
	q *sqlc.Queries
}

func New(q *sqlc.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) CreateUser(ctx context.Context, name string, dob time.Time) (sqlc.User, error) {
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Name: name,
		Dob:  dob,
	})
}

func (r *UserRepository) GetUser(ctx context.Context, id int32) (sqlc.User, error) {
	return r.q.GetUser(ctx, id)
}
