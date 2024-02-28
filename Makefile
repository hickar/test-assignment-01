env ?= .env

.PHONY: up-services
up-services: ## Deploy all services with Docker Compose
	docker compose --env-file $(env) up -d

.PHONY: down-services down-services-force
down-services: ## Bring down all services
	docker compose down
down-services-force: ## Bring down all services, delete all volumes
	docker compose down -v

.PHONY: lint lint-format lint-check
lint: lint-format lint-check ## Launches formating and golangci-lint
lint-format: ## Launches gofumpt
	gofumpt -l -w .
lint-check: ## Launches golangci-lint check
	golangci-lint run ./account_service/...
	golangci-lint run ./order_service/...
	# golangci-lint run ./shared/...

.PHONY: proto-gen
proto-gen: ## Generate service definition from .proto file
	protoc \ 
	--go_out="." --go_opt="paths=source_relative" \
	--go-grpc_out="." --go-grpc_opt="paths=source_relative" \
	./order_service/proto/order.proto

.PHONY: help
help: ## Display help information
	@grep -E '^[a-zA-Z_-]+:.*## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
