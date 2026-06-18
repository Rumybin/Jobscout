<p align="center">
  <h1 align="center">JobScout</h1>
  <p align="center">
    Remote-first tech job discovery and saved-application tracking API.
  </p>
</p>

<p align="center">
  <a href="https://github.com/Rumybin/Jobscout/actions">
    <img alt="CI" src="https://img.shields.io/github/actions/workflow/status/Rumybin/Jobscout/ci.yml?branch=main&label=CI&style=flat-square">
  </a>
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat-square&logo=go&logoColor=white">
  <img alt="PostgreSQL" src="https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat-square&logo=postgresql&logoColor=white">
  <img alt="Docker" src="https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white">
  <img alt="API" src="https://img.shields.io/badge/API-only_v1-111827?style=flat-square">
</p>

## Overview

JobScout helps users discover remote tech jobs, save interesting roles, track application status, add simple notes, and view basic saved-job analytics. It is an API-only v1 backend built to support a future frontend without a major rewrite.

The first external job source is the Remotive Public API. JobScout stores saved jobs as snapshots so users keep their own application records even if an external listing changes later.

## Features

- JWT auth: register, login, and `GET /me`.
- Remote job search with normalized Remotive results.
- Save external jobs or create manual applications.
- User-scoped list, detail, update, delete, and status update.
- Simple notes per saved application.
- Analytics summary by status, source, category, and saved month.
- DB-backed service tests and GitHub Actions CI.
- Docker-ready production image.

## Tech Stack

| Area | Choice |
| --- | --- |
| Language | Go |
| Router | Chi |
| Database | PostgreSQL |
| SQL access | sqlc |
| Migrations | golang-migrate compatible SQL |
| Auth | JWT + bcrypt |
| External jobs | Remotive Public API |
| Local infra | Docker Compose |
| Tests | Go testing, httptest, DB-backed integration tests |

## Quick Start

Clone and enter the repo:

```bash
git clone https://github.com/Rumybin/Jobscout.git
cd Jobscout
```

Copy environment config:

```bash
cp env.example .env
export DATABASE_URL="postgres://jobscout:jobscout@localhost:5432/jobscout?sslmode=disable"
```

Start PostgreSQL:

```bash
docker compose up -d
```

Apply migrations:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

If `migrate` is not installed, apply the SQL files manually:

```bash
docker compose exec -T db psql -U jobscout -d jobscout -f - < migrations/000001_create_users.up.sql
docker compose exec -T db psql -U jobscout -d jobscout -f - < migrations/000002_applications_table.up.sql
docker compose exec -T db psql -U jobscout -d jobscout -f - < migrations/000003_application_notes_table.up.sql
```

Run the API:

```bash
go run ./cmd/api
```

Health check:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## Environment

`env.example` documents the required variables:

| Variable | Description |
| --- | --- |
| `APP_ENV` | Runtime environment, for example `development` or `production`. |
| `APP_PORT` | HTTP port. Defaults to `8080`. |
| `DATABASE_URL` | PostgreSQL connection string. |
| `JWT_SECRET` | Secret used to sign JWTs. Use a strong value in production. |
| `REMOTIVE_BASE_URL` | Remotive API endpoint. |

Never commit real secrets.

## API Overview

Public endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health` | Health check. |
| `POST` | `/auth/register` | Create a user and return a token. |
| `POST` | `/auth/login` | Authenticate and return a token. |

Protected endpoints require `Authorization: Bearer <token>`:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/me` | Current user profile. |
| `GET` | `/jobs/search` | Search remote jobs. |
| `POST` | `/applications` | Save external job or create manual application. |
| `GET` | `/applications` | List saved applications. |
| `GET` | `/applications/{id}` | Get one saved application. |
| `PATCH` | `/applications/{id}` | Update editable application fields. |
| `DELETE` | `/applications/{id}` | Delete an application. |
| `PATCH` | `/applications/{id}/status` | Update status only. |
| `POST` | `/applications/{id}/notes` | Add a note. |
| `GET` | `/applications/{id}/notes` | List notes for an application. |
| `DELETE` | `/notes/{id}` | Delete a note. |
| `GET` | `/analytics/summary` | Saved-job analytics summary. |

## Example Flow

Register:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'
```

Login and store token:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}' \
  | sed -E 's/.*"token":"([^"]+)".*/\1/')
```

Search jobs:

```bash
curl "http://localhost:8080/jobs/search?keyword=backend&limit=2" \
  -H "Authorization: Bearer $TOKEN"
```

Save a manual application:

```bash
curl -X POST http://localhost:8080/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Backend Developer","company_name":"Example Co"}'
```

Get analytics:

```bash
curl http://localhost:8080/analytics/summary \
  -H "Authorization: Bearer $TOKEN"
```

Example analytics response:

```json
{
  "total_saved_jobs": 1,
  "by_status": {
    "Wishlist": 1
  },
  "by_source": {
    "manual": 1
  },
  "by_category": {
    "Uncategorized": 1
  },
  "saved_per_month": [
    {
      "month": "2026-06",
      "count": 1
    }
  ]
}
```

## Testing

Default local test run:

```bash
go test ./...
```

Full DB-backed test run:

```bash
docker compose up -d
export TEST_DATABASE_URL="postgres://jobscout:jobscout@localhost:5432/jobscout?sslmode=disable"
go test -race -count=1 ./...
docker compose down
```

Benchmarks:

```bash
go test -bench=. ./internal/auth ./internal/httputil
```

GitHub Actions starts PostgreSQL and sets `TEST_DATABASE_URL`, so CI runs DB-backed tests automatically.

## Docker

Build the image:

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

For local Docker-to-host PostgreSQL on Linux:

```bash
docker run --rm -p 8080:8080 \
  --add-host=host.docker.internal:host-gateway \
  -e DATABASE_URL="postgres://jobscout:jobscout@host.docker.internal:5432/jobscout?sslmode=disable" \
  -e JWT_SECRET="local-test-secret" \
  jobscout-api
```

## Deployment Notes

JobScout is ready for Docker-first deployment on platforms such as Render, Railway, Fly.io, or a VPS.

Recommended production flow:

1. Provision PostgreSQL.
2. Set production environment variables in the hosting platform.
3. Run migrations separately:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

4. Deploy the Docker image.
5. Check `GET /health`.

Migrations intentionally do not run automatically during app startup. This keeps schema changes explicit and lets deployment stop before serving traffic if migration fails.

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

## Release

Current release tag:

```text
v1.0.1
```

## Notes

Remotive results should be treated as active external listings, not guaranteed real-time openings. Users should verify and apply through the original external job URL.
