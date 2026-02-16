GOOSE ?= goose
MIGRATIONS_DIR ?= db/migrations
DB_HOST ?= 127.0.0.1

ifneq (,$(wildcard .env))
include .env
export DB_PASS
endif

DB_PASS_CLEAN := $(patsubst "%",%,$(patsubst '%',%,$(DB_PASS)))

ifeq ($(strip $(DB_PASS_CLEAN)),)
DB_URL ?= postgres://moviestack@$(DB_HOST):5432/moviestack_dev?sslmode=disable
else
DB_URL ?= postgres://moviestack:$(DB_PASS_CLEAN)@$(DB_HOST):5432/moviestack_dev?sslmode=disable
endif

.PHONY: migrate-status migrate-up migrate-down migrate-reset migrate-create

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-reset:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" reset

migrate-create:
ifndef name
	$(error usage: make migrate-create name=add_feature)
endif
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql
