# backend/Makefile

MAIN_GO := cmd/nas/main.go

GO_VERSION := $(shell go version | awk '{print $$3}')
NODE_VERSION := $(shell node -v)

.PHONY: build

build:
	@echo "Building backend..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
	go build -ldflags "-X 'nas-go/api/api.Version=x.x.x' -X 'nas-go/api/api.CommitHash=abc123' -X 'nas-go/api/api.GoVersion=$(GO_VERSION)' -X 'nas-go/api/api.NodeVersion=$(NODE_VERSION)'" \
	-o kuranas.exe ./cmd/nas
	@echo "Backend build complete."

test:
	@echo "Running tests..."
	@go test -tags=dev -v ./tests/files_test/...
	@echo "Tests completed."

run:
	@echo "Running backend..."
	@go run -ldflags="-X 'nas-go/api/api.Version=develop' \
    -X 'nas-go/api/api.CommitHash=develop' \
    -X 'nas-go/api/api.GoVersion=$(GO_VERSION)' \
    -X 'nas-go/api/api.NodeVersion=$(NODE_VERSION)'" \
    -tags=dev cmd/nas/main.go