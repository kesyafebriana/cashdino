.PHONY: run build seed migrate-up migrate-down migrate-create test lint

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

seed:
	go run cmd/seed/main.go

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

test:
	go test ./...

lint:
	golangci-lint run
