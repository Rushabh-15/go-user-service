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

func (r *UserRepository) UpdateUser(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error) {
	return r.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		Name: name,
		Dob:  dob,
		ID:   id,
	})
}

// DeleteUser returns the number of rows deleted (0 means no such user).
func (r *UserRepository) DeleteUser(ctx context.Context, id int32) (int64, error) {
	return r.q.DeleteUser(ctx, id)
}

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int32) ([]sqlc.User, error) {
	return r.q.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}
