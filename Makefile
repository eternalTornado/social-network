-include .env

MIGRATIONS_PATH = ./cmd/migrate/migrations
BINARY          = ./bin/main

# Khớp với docker-compose.yml
DB_CONTAINER = postgres-db
DB_USER      = admin
DB_NAME      = socialnetwork

.DEFAULT_GOAL := help

.PHONY: help dev run build test fmt vet tidy clean \
	db-up db-down db-reset db-logs psql \
	migration migrate-up migrate-down migrate-version migrate-force

## help: In danh sách lệnh
help:
	@grep -E '^## ' Makefile | sed 's/^## //' | awk '{i = index($$0, ": "); printf "  \033[36m%-18s\033[0m %s\n", substr($$0, 1, i - 1), substr($$0, i + 2)}'

# ---------- Go ----------

## dev: Chạy API với air (tự build lại khi sửa code)
dev:
	air

## run: Chạy API một lần, không live reload
run:
	go run ./cmd/api

## build: Build binary ra ./bin/main
build:
	go build -o $(BINARY) ./cmd/api

## test: Chạy toàn bộ test
test:
	go test -v ./...

## fmt: Format code
fmt:
	go fmt ./...

## vet: Soát lỗi thường gặp bằng go vet
vet:
	go vet ./...

## tidy: Dọn go.mod / go.sum
tidy:
	go mod tidy

## clean: Xoá thư mục bin/
clean:
	rm -rf ./bin

# ---------- Database (Docker) ----------

## db-up: Bật Postgres và đợi tới khi nhận kết nối
db-up:
	docker compose up -d
	@for i in $$(seq 1 30); do \
		docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER) -d $(DB_NAME) >/dev/null 2>&1 && echo "Postgres đã sẵn sàng" && exit 0; \
		sleep 1; \
	done; \
	echo "Postgres chưa sẵn sàng sau 30s, xem log: make db-logs"; exit 1

## db-down: Tắt Postgres (giữ dữ liệu)
db-down:
	docker compose down

## db-reset: XOÁ HẾT dữ liệu, bật lại DB trắng rồi chạy migration
db-reset:
	docker compose down -v
	$(MAKE) db-up
	$(MAKE) migrate-up

## db-logs: Xem log Postgres (Ctrl+C để thoát)
db-logs:
	docker compose logs -f db

## psql: Mở psql bên trong container
psql:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

# ---------- Migration (golang-migrate) ----------

## migration: Tạo cặp file migration, ví dụ: make migration create_users
migration:
	migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

## migrate-up: Chạy mọi migration chưa áp dụng
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_ADDR)" up

## migrate-down: Undo N migration gần nhất, ví dụ: make migrate-down 1
migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_ADDR)" down $(filter-out $@,$(MAKECMDGOALS))

## migrate-version: Xem version hiện tại của DB (và cờ dirty)
migrate-version:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_ADDR)" version

## migrate-force: Gỡ dirty state, ví dụ: make migrate-force 3
migrate-force:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_ADDR)" force $(filter-out $@,$(MAKECMDGOALS))

# Cho phép truyền tham số kiểu "make migration add_foo": mọi tên lạ khớp rule này và không làm gì
%:
	@:
