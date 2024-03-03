.PHONY: run
run: ## Запуск сервиса в локальной среде 
	go run ./cmd/main.go -config="config.local.yaml"

.PHONY: proto-gen
proto-gen: ## Генерация кода GRPC сервисов из .proto файлов 
	protoc --go_out="." --go_opt="paths=source_relative" \
		--go-grpc_out="." --go-grpc_opt="paths=source_relative" \
		./order_service/proto/order.proto

.PHONY: test
test: ## Запуск тестов 
	go test -v ./...

.PHONY: proto-gen
proto-gen: ## Генерация кода GRPC сервисов из .proto файлов 
	protoc --go_out="." --go_opt="paths=source_relative" \
		--go-grpc_out="." --go-grpc_opt="paths=source_relative" \
		./order_service/proto/order.proto

.PHONY: help
help: ## Вывод списка доступных комманд 
	@grep -E '^[a-zA-Z_-]+:.*## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
