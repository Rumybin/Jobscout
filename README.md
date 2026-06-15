# JobScout

Remote-first tech job discovery and saved-job tracking API built with Go.

## Prerequisites

- Go 1.22+
- Docker and Docker Compose

## Quick Start

```bash
# Copy environment configuration
cp env.example .env

# Start PostgreSQL
docker compose up -d

# Run the API server
go run ./cmd/api

# Health check
curl http://localhost:8080/health
```

## Useful Commands

```bash
go run ./cmd/api      # Start the server
go test ./...          # Run all tests
docker compose up -d   # Start PostgreSQL
docker compose down    # Stop PostgreSQL
```

## Environment Variables

See `env.example` for the full list of required environment variables.
