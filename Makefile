DB_URL = postgres://postgres:postgres@localhost:5432/db?sslmode=disable
RUN_PATH = cmd/server/main.go
MIGRATIONS_PATH = ./migrations

.PHONY: run ping test migrate-up migrate-down migrate-create

# Использование: make run-pg RACE=1 (Запуск сервера или теста с флагом -race)
RACE_FLAG :=
ifdef RACE
RACE_FLAG := -race
endif

# ---------------------- RUN
# Запуск сервера с настройками по умолчанию (in-memory хранилище)
run:
	go run $(RACE_FLAG) $(RUN_PATH) -d="$(DB_URL)"

run-smtp:
	go run $(RACE_FLAG) $(RUN_PATH) -d="$(DB_URL)" -e="true"

# ---------------------- MIGRATIONS
# Создание новой миграции: make migrate-create name=my_migration
migrate-create:
	@test -n "$(name)" || (echo "Error: name is required. Use: make migrate-create name=my_migration" && exit 1)
	migrate create -ext sql -dir "$(MIGRATIONS_PATH)" -seq "$(name)"

# Применить все миграции
migrate-up:
	migrate -database "$(DB_URL)" -path "$(MIGRATIONS_PATH)" up

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

# Удалить том с данными Postgres
docker-volume-rm:
	docker volume rm shorten_url_data || true 


# ---------------------- PROTOBUF
protoc:
	protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --go_opt=default_api_level=API_OPAQUE \
	-I . \
  pkg/proto/base.proto


# ---------------------- MOCKGEN
# Сгенерировать моки репозитория через mockgen
mockgen:
	mockgen -source=internal/repository/repository.go -destination=internal/repository/mocks/mock_repository.go -package=mocks 
