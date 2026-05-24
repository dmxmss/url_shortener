SHELL := /bin/sh

IMAGE_REGISTRY ?= ghcr.io/dmxmss
IMAGE_TAG ?= local

.PHONY: test build frontend-dev frontend-build run-api run-redirect compose-up compose-down compose-logs migrate docker-build docker-push helm-lint fmt

fmt:
	gofmt -w api redirect-service internal

test:
	go test ./...

build:
	go build -o bin/shortener-api ./api/cmd
	go build -o bin/redirect-service ./redirect-service/cmd

frontend-dev:
	npm --prefix frontend ci
	npm --prefix frontend run dev

frontend-build:
	npm --prefix frontend ci
	npm --prefix frontend run build

run-api:
	go run ./api/cmd

run-redirect:
	go run ./redirect-service/cmd

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f --tail=100

migrate:
	docker compose exec -T postgres psql -U shortener -d shortener < scripts/migrations/001_init.sql

docker-build:
	docker build -f api/Dockerfile -t $(IMAGE_REGISTRY)/shortener-api:$(IMAGE_TAG) .
	docker build -f redirect-service/Dockerfile -t $(IMAGE_REGISTRY)/redirect-service:$(IMAGE_TAG) .
	docker build -f frontend/Dockerfile -t $(IMAGE_REGISTRY)/frontend:$(IMAGE_TAG) .

docker-push:
	docker push $(IMAGE_REGISTRY)/shortener-api:$(IMAGE_TAG)
	docker push $(IMAGE_REGISTRY)/redirect-service:$(IMAGE_TAG)
	docker push $(IMAGE_REGISTRY)/frontend:$(IMAGE_TAG)

helm-lint:
	helm lint deploy/helm/url-shortener
