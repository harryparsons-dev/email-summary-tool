# Email Summary Tool

An email-summary application under active development. The local development stack includes:

- a Go API with live reload
- a Vue/Vite frontend with hot module replacement
- PostgreSQL 17 with persistent local storage

## Prerequisites

Install Git and a current version of Docker Desktop, or Docker Engine with Docker Compose v2. Go, Node.js, and PostgreSQL do not need to be installed on the host machine.

## First-time setup

Clone the repository and enter it:

```sh
git clone git@github.com:harryparsons-dev/email-summary-tool.git
cd email-summary-tool
```

Create the local environment file:

```sh
cp .env.example .env
```

Open `.env` and replace the example PostgreSQL password. Then build and start the development stack:

```sh
docker compose up --build
```

The services are available at:

- Frontend: <http://localhost:5173>
- API health endpoint: <http://localhost:8080>
- PostgreSQL: `localhost:5434`

The `.env` file is ignored by Git. Commit changes to `.env.example` whenever a new required environment variable is introduced, but never add real credentials to it.

## Common commands

Start the stack in the background:

```sh
docker compose up --build --detach
```

Follow service logs:

```sh
docker compose logs --follow
```

Stop the stack while preserving database data:

```sh
docker compose down
```

Rebuild after changing a Dockerfile or dependency manifest:

```sh
docker compose up --build
```

Run the backend tests:

```sh
docker compose exec api go test ./...
```

Create an empty up/down migration pair:

```sh
docker compose exec api go run -tags pgx5 github.com/golang-migrate/migrate/v4/cmd/migrate create -ext sql -dir database/migrations/sql -seq -digits 6 migration_name
```

Migration files live in `database/migrations/sql`. The API does not run them
automatically; applying or rolling back migrations must be done explicitly.

Run the frontend type checker:

```sh
docker compose exec frontend npm run type-check
```

## Resetting local data

PostgreSQL data and frontend dependencies are stored in Docker volumes and survive normal container restarts. To stop the stack and delete those volumes:

```sh
docker compose down --volumes
```

This permanently deletes the local development database. The next `docker compose up --build` creates a clean one.

## Environment variables

Docker Compose reads the following values from `.env`:

| Variable | Purpose |
| --- | --- |
| `POSTGRES_DB` | Name of the local application database |
| `POSTGRES_USER` | PostgreSQL application user |
| `POSTGRES_PASSWORD` | Password for the PostgreSQL user |

Changing these values does not update an existing PostgreSQL volume. Reset the local data volume if the database has already been initialized and the credentials need to change.
