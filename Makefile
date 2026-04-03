include .env

.PHONY: run dev migrate-up migrate-down lint test

run:
	go run ./cmd/server

dev:
	docker-compose up --build

migrate-new:
	migrate create -ext sql -dir ./migrations -seq $(name)

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

lint:
	golangci-lint run ./...

test:
	go test -race -cover ./...