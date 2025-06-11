# backend/Makefile

MAIN_GO := cmd/nas/main.go

GO_VERSION := $(shell go version | awk '{print $$3}')
NODE_VERSION := $(shell node -v)

.PHONY: build

build:
	@echo "Building backend..."
	@go build -o kuranas $(MAIN_GO)
	@echo "Backend build complete."

test:
	@echo "Running tests..."
	@go test ./tests/files_test/...
	@echo "Tests completed."

run:
	@echo "Running backend..."
	@go run -ldflags="-X 'nas-go/api/api.Version=develop' \
    -X 'nas-go/api/api.CommitHash=develop' \
    -X 'nas-go/api/api.GoVersion=$(GO_VERSION)' \
    -X 'nas-go/api/api.NodeVersion=$(NODE_VERSION)'" \
    -tags=dev cmd/nas/main.go