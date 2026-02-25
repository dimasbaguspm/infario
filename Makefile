
COMPOSE=docker compose
SERVICE_NAME=api
MAIN_PATH=cmd/app/main.go

.PHONY: up down restart logs swagger shell

## up: Start the full stack (DB, Redis, API) in detached mode
up: swagger
	@echo "Launching Infario..."
	$(COMPOSE) up

## down: Stop all containers and remove networks
down:
	@echo "Shutting down..."
	$(COMPOSE) down

## restart: Quick restart of the API service only (useful after code changes)
restart: swagger
	@echo "Restarting API..."
	$(COMPOSE) restart $(SERVICE_NAME)

## swagger: Generate docs locally (so they are available for the container build)
swagger:
	@echo "Generating Swagger docs..."
	swag init -g $(MAIN_PATH) --parseDependency --parseInternal

## logs: Tail logs from the API container
logs:
	$(COMPOSE) logs -f $(SERVICE_NAME)

## shell: Drop into the API container shell
shell:
	$(COMPOSE) exec $(SERVICE_NAME) sh