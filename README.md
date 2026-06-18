# JobScout

JobScout is a remote-first tech job discovery and saved-job tracking API built with Go. It helps users search external remote job listings, save interesting roles, track application status, add simple notes, and view basic saved-job analytics.

This repository is API-only for v1. A frontend may be added later, but the backend is kept modular enough to support one without a major rewrite.

## Features

- User registration, login, JWT authentication, and `GET /me`.
- Remote job search through the Remotive Public API.
- Normalized job search responses for external listings.
- Saved applications for external jobs and manual entries.
- User-scoped list, detail, partial update, delete, and status update operations.
- Simple application notes.
- Basic analytics summary for saved jobs.

## Tech Stack

- Go
- Chi router
- PostgreSQL
- sqlc
- golang-migrate compatible SQL migrations
- JWT authentication
- bcrypt password hashing
- Docker Compose for local PostgreSQL

## Requirements

- Go 1.26.4 or compatible
- Docker and Docker Compose
- `sqlc` when regenerating database query code
- A PostgreSQL migration runner, such as `migrate`, when applying migrations outside Docker/database bootstrapping

## Quick Start

Copy the example environment file:

```bash
cp env.example .env
```

Start PostgreSQL:

```bash
docker compose up -d
```

Apply database migrations with your migration runner. Example:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Run the API:

```bash
go run ./cmd/api
```

Check the server:

```bash
curl http://localhost:8080/health
```

## Environment Variables

`env.example` documents the required configuration:

- `APP_ENV`
- `APP_PORT`
- `DATABASE_URL`
- `JWT_SECRET`
- `REMOTIVE_BASE_URL`

Never commit real secrets. Use a strong `JWT_SECRET` outside local development.

## API Overview

Public endpoints:

- `POST /auth/register`
- `POST /auth/login`
- `GET /health`

Protected endpoints require `Authorization: Bearer <token>`:

- `GET /me`
- `GET /jobs/search`
- `POST /applications`
- `GET /applications`
- `GET /applications/{id}`
- `PATCH /applications/{id}`
- `DELETE /applications/{id}`
- `PATCH /applications/{id}/status`
- `POST /applications/{id}/notes`
- `GET /applications/{id}/notes`
- `DELETE /notes/{id}`
- `GET /analytics/summary`

## Analytics Summary

`GET /analytics/summary` returns saved-job analytics for the authenticated user:

```json
{
  "total_saved_jobs": 12,
  "by_status": {
    "Wishlist": 4,
    "Applied": 3
  },
  "by_source": {
    "remotive": 10,
    "manual": 2
  },
  "by_category": {
    "Software Development": 8,
    "Uncategorized": 1
  },
  "saved_per_month": [
    {
      "month": "2026-06",
      "count": 5
    }
  ]
}
```

## Development Commands

```bash
go run ./cmd/api
go test ./...
go vet ./...
go build ./...
go test -race -count=1 ./...
go test -bench=. ./internal/auth ./internal/httputil
sqlc generate
docker compose up -d
docker compose down
```

If `sqlc` is not installed locally, you can run:

```bash
go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
```

## Testing Strategy

The default test command is safe for local development:

```bash
go test ./...
```

Service-layer integration tests run only when `TEST_DATABASE_URL` is set. They create an isolated PostgreSQL schema, apply migrations, run the test, and drop the schema during cleanup.

To run the full local suite with database-backed service tests:

```bash
docker compose up -d
export TEST_DATABASE_URL="postgres://jobscout:jobscout@localhost:5432/jobscout?sslmode=disable"
go test -race -count=1 ./...
```

Run the lightweight benchmarks with:

```bash
go test -bench=. ./internal/auth ./internal/httputil
```

GitHub Actions provides PostgreSQL and sets `TEST_DATABASE_URL`, so CI runs the DB-backed tests automatically.

## Docker

Build the API image:

```bash
docker build -t jobscout-api .
```

Run the API container:

```bash
docker run --rm -p 8080:8080 \
  -e APP_ENV=production \
  -e APP_PORT=8080 \
  -e DATABASE_URL="postgres://user:password@host:5432/jobscout?sslmode=require" \
  -e JWT_SECRET="replace-with-a-strong-secret" \
  -e REMOTIVE_BASE_URL="https://remotive.com/api/remote-jobs" \
  jobscout-api
```

For local Docker-to-host PostgreSQL on Linux, add the host gateway mapping:

```bash
docker run --rm -p 8080:8080 \
  --add-host=host.docker.internal:host-gateway \
  -e DATABASE_URL="postgres://jobscout:jobscout@host.docker.internal:5432/jobscout?sslmode=disable" \
  -e JWT_SECRET="local-test-secret" \
  jobscout-api
```

Check the container:

```bash
curl http://localhost:8080/health
```

## Deployment Notes

JobScout is ready for Docker-first deployment on platforms such as Render, Railway, Fly.io, or a VPS. Configure all required environment variables in the hosting platform; do not bake secrets into the image.

Run database migrations separately from app startup:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

This keeps schema changes explicit and lets a deployment stop before serving traffic if migration fails.

## Project Structure

```text
cmd/api                 API entry point
internal/auth           authentication handlers, services, JWT helpers
internal/middleware     request middleware
internal/jobsource      external job source adapter
internal/jobs           job search API
internal/applications   saved jobs, statuses, and notes
internal/analytics      saved-job analytics
internal/database       database connection and sqlc generated code
internal/httputil       shared HTTP response and error helpers
migrations              PostgreSQL migrations
sql                     sqlc query definitions
```

## Notes

JobScout uses Remotive as the only external job source in v1. External job listings should be treated as active external listings, not guaranteed real-time openings. Users should verify and apply through the original external job URL.
