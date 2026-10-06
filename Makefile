.PHONY: dev build run migrate

build:
	CGO_ENABLED=1 go build -o bin/chatix-server ./cmd/server

run: build
	./bin/chatix-server

migrate:
	# Требует установленный goose: go install github.com/pressly/goose/v3/cmd/goose@latest
	goose -dir migrations postgres "host=/var/run/postgresql dbname=chatix sslmode=disable" up

dev:
	# Запуск с автоматическим рестартом при изменении кода (требует air: go install github.com/air-verse/air@latest)
	air