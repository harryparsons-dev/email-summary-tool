.PHONY: help migration test-api

TEST_DATABASE_URL ?= postgres://postgres:postgres@127.0.0.1:5435/email_summary_test?sslmode=disable

help:
	@echo "Available targets:"
	@echo "  make migration NAME=add_users  Create a timestamped up/down migration pair"
	@echo "  make test-api                  Run API integration tests against the test database"

migration:
	@./scripts/new-migration.sh "$(NAME)"

test-api:
	@TEST_DATABASE_URL='$(TEST_DATABASE_URL)' go test -count=1 ./tests/...
