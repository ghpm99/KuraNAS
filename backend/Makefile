# backend/Makefile

MAIN_GO := cmd/nas/main.go

.PHONY: build

build:
	@echo "Building backend..."
	@go build -o kuranas $(MAIN_GO)
	@echo "Backend build complete."

test:
	@echo "Running tests..."
	@go test ./tests/files_test/...
	@echo "Tests completed."