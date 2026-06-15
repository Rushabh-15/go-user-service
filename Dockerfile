# ---- Build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies first so this layer is cached and only re-runs
# when go.mod / go.sum change (not on every source edit).
COPY go.mod go.sum ./
RUN go mod download

# Build a fully static binary. CGO is off because pgx is pure Go, so the
# result needs no C libraries. -ldflags "-s -w" strips the symbol table and
# DWARF debug info to shrink the binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.20

# Run as an unprivileged user instead of root.
RUN adduser -D appuser

WORKDIR /app
COPY --from=builder /app/server /app/server

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/server"]
