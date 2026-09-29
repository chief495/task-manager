.PHONY: mocks test build run run-debug clean

# Перегенерация моков
mocks:
	cd tests/mocks && go generate

# Запуск тестов
test:
	cd tests && go test -v

# Перегенерация моков + тесты
test-full: mocks test

# Сборка проекта
build:
	go build -o bin/main ./cmd/api

# Запуск (продакшен — подключение к postgres-1)
run:
	go run ./cmd/api

# Запуск (локальная разработка — подключение к localhost)
run-debug:
	TM_DEBUG=true go run ./cmd/api

# Очистка
clean:
	rm -rf bin/

# Обновление зависимостей
deps:
	go mod tidy
	go mod download

# Проверка go.mod на чистоту
deps-check:
	go mod tidy
	git diff --exit-code go.mod go.sum
