env ?= .env

.PHONY: up-services
up-services: ## Запуск всех сервисов через Docker Compose
	docker compose -f="docker-compose.yaml" --env-file="$(env)" up -d --wait

.PHONY: down-services down-services-force
down-services: ## Остановка 
	docker compose down
down-services-force: ## Остановка всех сервисов, включая удаление всех volume'ов 
	docker compose down -v

.PHONY: lint lint-format lint-check
lint: lint-format lint-check ## Запуск форматтера и линтеров 
lint-format: ## Запуск gofumpt 
	gofumpt -l -w .
lint-check: ## Запуск линтеров 
	golangci-lint run ./...

.PHONY: test
test: test-unit test-e2e ## Запуск всех тестов
test-unit: ## Запуск unit-тестов
	go test -v -tags="unit_test" ./...
test-e2e: ## Запуск e2e-тестов
	go test -v -tags="e2e_test" ./...

.PHONY: proto-gen
proto-gen: ## Генерация кода GRPC сервисов из .proto файлов 
	protoc --go_out="." --go_opt="paths=source_relative" \
		--go-grpc_out="." --go-grpc_opt="paths=source_relative" \
		./order_service/proto/order.proto

.PHONY: configure
configure: ## Настройка окружения 
	cp .env.example .env

.PHONY: help
help: ## Вывод списка доступных комманд 
	@grep -E '^[a-zA-Z_-]+:.*## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
