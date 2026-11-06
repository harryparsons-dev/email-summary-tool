# Repository Guidelines

## Project structure

- `main/` contains the API entry point.
- `server/`, `routes/`, `services/`, `validator/`, and `config/` contain backend application code.
- `database/models/` contains Bun models; `database/migrations/sql/` contains embedded PostgreSQL migrations.
- `tests/api/` contains authenticated API integration tests and their shared test server.
- `frontend/` is a Vue 3, TypeScript, Vite, Nuxt UI, and Tailwind application.
- `pkg/` contains shared Go test helpers. Preserve existing package and file naming unless the task includes cleanup.

## Development commands

- Copy `.env.example` to `.env` and set local secrets before starting the application.
- Start the development stack with `docker compose up --build`.
- Start the test database with `docker compose -f compose.test.yaml up --detach --wait test-postgres`.
- Run all Go tests with `go test -count=1 ./...`.
- Run API integration tests with `make test-api`; the test database container must already be running.
- Run frontend type checking with `docker compose exec frontend npm run type-check`.
- Run the complete frontend build with `docker compose exec frontend npm run build` when frontend behavior or build configuration changes.
- Create migrations with `make migration NAME=descriptive_name` rather than naming migration files manually.

## Implementation conventions

- Format changed Go files with `gofmt` and follow the existing package boundaries and Echo handler patterns.
- Keep request validation in request/validator types, persistence concerns in models or services, and HTTP response handling in handlers.
- In the frontend, use Vue single-file components with TypeScript and reuse the existing service and model layers for API access and types.
- Keep changes focused. Do not reformat generated frontend declaration files or unrelated code.
- Never commit `.env`, credentials, access tokens, or real user data. Update `.env.example` when a required environment variable changes.

## Database migrations

- Every schema change requires matching `.up.sql` and `.down.sql` files.
- Treat migrations as immutable after they are merged or used on a shared database. Make corrections in a new migration.
- Ensure new migration timestamps are later than all migrations intended for the same release.
- Keep Bun models and API behavior consistent with the resulting schema.

## Testing expectations

- Add or update focused tests for behavior changes and regressions.
- Prefer table-driven Go tests where several cases share setup.
- API integration tests should use the shared `TestServer` in `tests/api/main_test.go`; do not create a second server or database lifecycle.
- Expected integration-test JSON is subset-matched, so assert stable contract fields rather than generated values unless those values are the behavior under test.
- Every API integration test must define its expected `Body` inline in its `ExpectedResult`. `BodyDoesNotExist` is supplemental and does not replace `Body`; define it inline too. Never extract response-body expectations into variables or helpers, even when this duplicates them.
- Do not make tests start or stop the shared test database container.
- For backend changes, run the relevant package tests and normally `go test -count=1 ./...`.
- For frontend changes, run `npm run type-check` through the frontend container; run `npm run build` when the change can affect bundling or production output.

## Working-tree safety

- Assume pre-existing modifications belong to the user. Do not discard, overwrite, or include unrelated changes.
- Avoid destructive Git or Docker operations unless explicitly requested. In particular, `docker compose down --volumes` deletes local database data.
- Report commands that could not be run and the reason in the final handoff.
