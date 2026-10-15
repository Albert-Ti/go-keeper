DB_URL = postgres://postgres:postgres@localhost:5432/db?sslmode=disable
RUN_PATH = cmd/server/main.go
MIGRATIONS_PATH = ./migrations
PROTO_DIR := pkg/proto
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

.PHONY: run ping test migrate-up migrate-down migrate-create protoc

# Использование: make run-pg RACE=1 (Запуск сервера или теста с флагом -race)
RACE_FLAG :=
ifdef RACE
RACE_FLAG := -race
endif

# ---------------------- RUN
# Запуск сервера с настройками по умолчанию
run:
	go run $(RACE_FLAG) $(RUN_PATH) -d="$(DB_URL)"

run-smtp:
	go run $(RACE_FLAG) $(RUN_PATH) -d="$(DB_URL)" -e="true"

run-client:
	go run ./cmd/client/...

# ---------------------- MIGRATIONS
# Создание новой миграции: make migrate-create name=my_migration
migrate-create:
	@test -n "$(name)" || (echo "Error: name is required. Use: make migrate-create name=my_migration" && exit 1)
	migrate create -ext sql -dir "$(MIGRATIONS_PATH)" -seq "$(name)"

# Применить все миграции
migrate-up:
	migrate -database "$(DB_URL)" -path "./migrations" up

# Откатить все миграции
migrate-down:
	migrate -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" down

# Показать текущую версию миграции
migrate-v:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_PATH) version

# Показать статус миграций
migrate-status:
	migrate -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" status

# Принудительно выставить версию миграции без её применения: make migrate-force version=1
migrate-force:
	@test -n "$(version)" || (echo "Error: version is required. Use: make migrate-force version=1" && exit 1)
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_PATH) force $(version)

# Удалить все таблицы из базы (требует подтверждения)
migrate-drop:
	@echo "This will DROP EVERYTHING! Continue? [y/N]" && read ans && [ $${ans:-N} = y ]
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_PATH) drop -f

# Полный сброс базы: drop + повторное применение миграций
migrate-reset: migrate-drop migrate-up
	@echo "Database reset and migrations reapplied"


# ---------------------- DOCKER
# Поднять контейнеры (с пересозданием)
docker-up:
	docker compose up -d --force-recreate

# Остановить контейнеры
docker-down:
	docker compose down

# Зайти в контейнер postgres по bash
docker-exec:
	docker compose exec -t postgres bash

# Удалить том с данными s3 storage
docker-volume-rm:
	docker volume rm go-keeper_seaweedfs-data

# Команда для удаления контейнеров, образов, томов и сетей за один раз которые не используются
docker-prune:
	docker system prune -a --volumes

# ---------------------- PROTOBUF
protoc:
	protoc \
		--go_out=$(PROTO_DIR) \
		--go_opt=paths=source_relative \
		--go_opt=default_api_level=API_OPAQUE \
		--go-grpc_out=$(PROTO_DIR) \
		--go-grpc_opt=paths=source_relative \
		-I $(PROTO_DIR) \
		$(PROTO_FILES)

# ---------------------- MOCKGEN
# Сгенерировать моки репозитория через mockgen
mockgen:
	mockgen -source=internal/repository/repository.go -destination=internal/repository/mocks/mock_repository.go -package=mocks 
