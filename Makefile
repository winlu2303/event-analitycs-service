.PHONY: run test docker-up docker-down migrate migrate-all load-test

run:
	go run cmd/api/main.go

run-consumer:
	go run cmd/consumer/main.go

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

migrate-clickhouse:
	docker exec -i analytics-clickhouse clickhouse-client < migrations/clickhouse/001_init.sql

migrate-postgres:
	docker exec -i analytics-postgres psql -U admin -d analytics < migrations/postgres/001_init.sql

migrate-all: migrate-clickhouse migrate-postgres

test:
	go test ./tests/unit/...
	go test ./tests/integration/...

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

load-test:
	for i in {1..100}; do \
		curl -X POST http://localhost:8080/api/v1/events/track \
			-H "Content-Type: application/json" \
			-d '{"user_id":"user'$$i'","event_type":"page_view","page_url":"/home","metadata":{"test":true}}'; \
	done

benchmark:
	go test -bench=. ./...

lint:
	golangci-lint run

proto:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

swagger:
	swag init -g cmd/api/main.go