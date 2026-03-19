# ShikaGari — Makefile

APP_NAME  = shikagari
BUILD_DIR = ./bin
CMD_PATH  = ./cmd/api

.PHONY: run build clean dev test migrate lint

## run: Build and run the application
run:
	go run $(CMD_PATH)/main.go

## dev: Run with live reload using Air
dev:
	air

## build: Compile to binary
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR) tmp uploads

## test: Run all tests
test:
	go test ./... -v -count=1

## test-coverage: Run tests with coverage report
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

## lint: Run Go linter
lint:
	golangci-lint run ./...

## tidy: Tidy and verify Go modules
tidy:
	go mod tidy
	go mod verify

## migrate: Apply raw SQL migrations manually
migrate:
	@echo "Applying migrations..."
	@for f in migrations/*.sql; do \
		echo "  → $$f"; \
		psql "$$DATABASE_URL" -f $$f; \
	done
	@echo "Done."

## seed: Seed the database with test data
seed:
	go run scripts/seed/main.go

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'