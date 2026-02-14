GOOSE ?= goose
MIGRATIONS_DIR ?= db/migrations
DB_URL ?= postgres://localhost:5432/moviestack?sslmode=disable

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
