.PHONY: help migration

help:
	@echo "Available targets:"
	@echo "  make migration NAME=add_users  Create a timestamped up/down migration pair"

migration:
	@./scripts/new-migration.sh "$(NAME)"
