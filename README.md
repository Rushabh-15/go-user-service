# Ainyx User API

A small RESTful API in Go for managing users with a `name` and `dob` (date of
birth). The API stores the date of birth and returns each user's **age,
calculated dynamically** at request time using Go's `time` package.

## Features

- CRUD for users: create, read, update, delete, and list
- Age computed on the fly from `dob` (never stored, so it never goes stale)
- Type-safe database access generated with **SQLC**
- Input validation with **go-playground/validator**
- Structured logging with **Uber Zap**
- Clean HTTP status codes and a uniform error envelope

Bonus features (all included):

- **Docker** support — multi-stage build plus a Compose stack that runs
  migrations automatically
- **Pagination** on `GET /users` via `limit` / `offset`
- **Unit test** for the age-calculation function (table-driven, covers
  leap-year boundaries)
- **Middleware** — injects an `X-Request-Id` response header and logs request
  duration

## Tech stack

- [GoFiber](https://gofiber.io/) — HTTP framework
- PostgreSQL + [SQLC](https://sqlc.dev/) — type-safe queries
- [pgx](https://github.com/jackc/pgx) — PostgreSQL driver (pure Go)
- [Uber Zap](https://github.com/uber-go/zap) — logging
- [go-playground/validator](https://github.com/go-playground/validator) — validation
- [golang-migrate](https://github.com/golang-migrate/migrate) — schema migrations

## Prerequisites

- **Docker** and **Docker Compose** for the quick start, **or**
- **Go 1.26+** and a local **PostgreSQL** for local development

## Quick start (Docker)

The simplest way to run everything — database, migrations, and API — with one
command:

```bash
git clone https://github.com/Rushabh-15/ainyx-user-api
cd ainyx-user-api
docker compose up --build
```

Compose starts PostgreSQL, waits for it to be healthy, runs the migrations, then
starts the API on http://localhost:8080. Verify it:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Tear everything down (including the database volume):

```bash
docker compose down -v
```

## Local development (without Docker)

1. Start a local PostgreSQL and create a database named `usersdb`.
2. Copy the example environment file:

   ```bash
   cp .env.example .env
   ```

3. Apply migrations (requires the `migrate` CLI):

   ```bash
   migrate -path db/migrations \
     -database "postgres://postgres:postgres@localhost:5432/usersdb?sslmode=disable" up
   ```

4. Run the server:

   ```bash
   go run ./cmd/server
   ```

## Environment variables

| Variable       | Description                                       | Default       |
| -------------- | ------------------------------------------------- | ------------- |
| `DATABASE_URL` | PostgreSQL connection string                      | _(required)_  |
| `PORT`         | Port the HTTP server listens on                   | `8080`        |
| `APP_ENV`      | `production` (JSON logs) or other (console logs)  | `development` |

## Database migrations

Migrations live in `db/migrations/` as versioned `.up.sql` / `.down.sql` pairs.
Under Docker they run automatically via a one-shot `migrate` service. Locally,
use the `migrate` CLI. To roll back the most recent migration:

```bash
migrate -path db/migrations -database "$DATABASE_URL" down 1
```

## Regenerating database code

After changing `db/query.sql` or the schema, regenerate the SQLC layer:

```bash
sqlc generate
```

## API reference

Base URL: `http://localhost:8080`. Every response carries an `X-Request-Id`
header for tracing.

### Create user — `POST /users`

Request:

```json
{ "name": "Alice", "dob": "1990-05-10" }
```

Response `201 Created`:

```json
{ "id": 1, "name": "Alice", "dob": "1990-05-10" }
```

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","dob":"1990-05-10"}'
```

### Get user — `GET /users/:id`

Response `200 OK` (includes the computed `age`):

```json
{ "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 }
```

```bash
curl http://localhost:8080/users/1
```

### Update user — `PUT /users/:id`

Request:

```json
{ "name": "Alice Updated", "dob": "1991-03-15" }
```

Response `200 OK`:

```json
{ "id": 1, "name": "Alice Updated", "dob": "1991-03-15" }
```

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","dob":"1991-03-15"}'
```

### Delete user — `DELETE /users/:id`

Response `204 No Content` (empty body).

```bash
curl -X DELETE http://localhost:8080/users/1
```

### List users — `GET /users`

Paginated via `limit` (default 10, max 100) and `offset` (default 0).
Response `200 OK`:

```json
[ { "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 } ]
```

```bash
curl "http://localhost:8080/users?limit=5&offset=0"
```

### Error format

All errors share one envelope:

```json
{ "error": "validation failed", "details": { "name": "this field is required" } }
```

| Status | Meaning                               |
| ------ | ------------------------------------- |
| `400`  | Invalid body, params, or validation   |
| `404`  | User not found                        |
| `500`  | Unexpected server error               |

## Running tests

```bash
go test ./...
```

The age-calculation function is covered by a table-driven test, including
leap-year boundary cases.

## Project structure

```
cmd/server/          application entrypoint
config/              environment configuration
db/migrations/       versioned SQL migrations
db/sqlc/             SQLC-generated database code
internal/handler/    HTTP handlers
internal/service/    business logic + age calculation
internal/repository/ data-access wrapper over SQLC
internal/routes/     route registration
internal/middleware/ request id + request logging
internal/models/     request/response types
internal/logger/     Zap logger setup
```

## Design notes

- **Age is never stored** — only `dob` is persisted; age is derived per request.
- **Layering**: handler (HTTP) → service (business rules) → repository (SQLC).
  The service declares the small repository interface it needs, keeping the
  layers decoupled and the service unit-testable.
- **Errors** are sentinel values returned from the service and mapped to HTTP
  status codes in the handler — the service knows nothing about HTTP, and the
  handler knows nothing about the database driver.
