SHELL := /bin/bash
.PHONY: help migrate-up migrate-down migrate-status migrate-create
.DELETE_ON_ERROR:
.SUFFIXES:

ENV ?= dev
ENVS := dev staging prod

ifeq ($(filter $(ENV), $(ENVS)), )
$(error Invalid ENV=$(ENV). ENV must be one of $(ENVS))
endif

.DEFAULT_GOAL := help
help:
	@echo "Available targets:"
	@echo "  migrate-up - Migrate up to the latest version"
	@echo "  migrate-down - Migrate down to a specific version or steps"
	@echo "  migrate-status - Show migration status"
	@echo "  migrate-create - Create a new migration"
	@echo ""
	@echo "Example Usage"
	@echo "  make migrate-up ENV=dev"
	@echo "  make migrate-down ENV=dev version=20240101120000"
	@echo "  make migrate-status ENV=dev"
	@echo "  make migrate-create ENV=dev name=add_users_table"

define prod-confirm
	@if [ "$(ENV)" = "prod" ]; then \
		echo "Are you sure you want to run this command in production?"; \
		read -p "Type 'y' to continue: " confirm; \
		if [ "$$confirm" != "y" ]; then \
			echo "Aborting..."; \
			exit 1; \
		fi; \
	fi
endef

define build-url
$(shell infisical run --env=$(ENV) -- bash -c 'echo "$$DATABASE_URL"')
endef

MIGRATE := migrate -path sql

migrate-status:
	@echo "Checking migration status..."
	$(MIGRATE) -database "$(call build-url)" version

migrate-up:
	$(call prod-confirm)
	@echo "Migrating up..."
	$(MIGRATE) -database "$(call build-url)" up

migrate-down:
	$(call prod-confirm)
	@if [ -n "$(version)" ]; then \
		echo "Migrating to version $(version)..."; \
		$(MIGRATE) -database "$(call build-url)" goto "$(version)"; \
	elif [ -n "$(steps)" ]; then \
		echo "Migrating down $(steps) steps..."; \
		$(MIGRATE) -database "$(call build-url)" down "$(steps)"; \
	else \
		echo "Missing version or steps parameter. Usage: make migrate-down version=<version> or make migrate-down steps=<steps>"; \
		exit 1; \
	fi

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Missing name parameter. Usage: make migrate-create name=<name>"; \
		exit 1; \
	fi
	@echo "Creating migration..."
	migrate create -ext sql -dir sql "$(name)"
