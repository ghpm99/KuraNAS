# Makefile

FRONTEND_DIR := frontend
BACKEND_DIR := backend
DIST_DIR := $(FRONTEND_DIR)/dist
OUTPUT_DIR := build

BACKEND_COVERAGE_MIN := 80
BACKEND_GO_CACHE_DIR := $(abspath .cache/go-build)
BACKEND_GO_MOD_CACHE_DIR := $(abspath .cache/go-mod)
BACKEND_GO_ENV := GOCACHE=$(BACKEND_GO_CACHE_DIR) GOMODCACHE=$(BACKEND_GO_MOD_CACHE_DIR)

.PHONY: all frontend backend move clean deploy ci ci-frontend ci-backend lint-frontend test-frontend lint-backend test-backend prepare-backend-go-cache release-main-ff

all: frontend backend move deploy

frontend:
	@$(MAKE) -C frontend

backend:
	@$(MAKE) -C backend

move:
	@echo "Moving files..."
	@mkdir -p $(OUTPUT_DIR)/dist
	@mv $(DIST_DIR) $(OUTPUT_DIR)
	@mv $(BACKEND_DIR)/kuranas $(OUTPUT_DIR)
	@cp -r $(BACKEND_DIR)/icons $(OUTPUT_DIR)
	@cp -r $(BACKEND_DIR)/translations $(OUTPUT_DIR)
	@echo "Files moved."

clean:
	@echo "Cleaning..."
	@rm -rf $(OUTPUT_DIR)
	@echo "Clean complete."

ci: ci-frontend ci-backend
	@echo ""
	@echo "========================================"
	@echo "  All quality gates passed"
	@echo "========================================"

ci-frontend: lint-frontend test-frontend

ci-backend: lint-backend test-backend

prepare-backend-go-cache:
	@mkdir -p $(BACKEND_GO_CACHE_DIR) $(BACKEND_GO_MOD_CACHE_DIR)

lint-frontend:
	@echo ""
	@echo "======== Frontend Lint ========"
	@cd $(FRONTEND_DIR) && yarn lint
	@echo "Frontend lint passed."

test-frontend:
	@echo ""
	@echo "======== Frontend Tests ========"
	@cd $(FRONTEND_DIR) && yarn test --coverage --watchAll=false
	@echo ""
	@echo "======== Frontend Build ========"
	@cd $(FRONTEND_DIR) && yarn build
	@echo "Frontend tests and build passed."
	@cd $(FRONTEND_DIR) && yarn typecheck:test

lint-backend:
	@$(MAKE) prepare-backend-go-cache
	@echo ""
	@echo "======== Backend Lint (gofmt) ========"
	@cd $(BACKEND_DIR) && BADFILES=$$(rg --files -g '*.go' | while IFS= read -r f; do \
		TMP_FMT=$$(mktemp) && TMP_SRC=$$(mktemp) && \
		gofmt "$$f" > "$$TMP_FMT" && \
		tr -d '\r' < "$$f" > "$$TMP_SRC" && \
		if ! cmp -s "$$TMP_FMT" "$$TMP_SRC"; then echo "$$f"; fi; \
		rm -f "$$TMP_FMT" "$$TMP_SRC"; \
	done) && \
	if [ -n "$$BADFILES" ]; then \
		echo "gofmt check failed on:" && echo "$$BADFILES" && exit 1; \
	fi
	@echo "gofmt passed."
	@echo ""
	@echo "======== Backend Lint (go vet) ========"
	@cd $(BACKEND_DIR) && $(BACKEND_GO_ENV) go vet ./...
	@echo "go vet passed."

test-backend:
	@$(MAKE) prepare-backend-go-cache
	@echo ""
	@echo "======== Backend Tests + Coverage ========"
	@cd $(BACKEND_DIR) && $(BACKEND_GO_ENV) go test ./... -coverprofile=coverage.out
	@echo ""
	@echo "======== Backend Coverage Threshold ========"
	@cd $(BACKEND_DIR) && $(BACKEND_GO_ENV) \
		COVERAGE=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
		echo "Backend coverage: $${COVERAGE}% (minimum: $(BACKEND_COVERAGE_MIN)%)"; \
		awk -v c="$$COVERAGE" -v min="$(BACKEND_COVERAGE_MIN)" 'BEGIN { exit(c >= min ? 0 : 1) }' || \
		(echo "FAILED: Coverage below $(BACKEND_COVERAGE_MIN)% threshold" && exit 1)
	@echo "Backend tests and coverage passed."

-include Makefile.local

deploy:
	@echo "Calling local deploy..."
	@$(MAKE) -f Makefile.local deploy
	@echo "Local deploy complete."

release-main-ff:
	@echo ""
	@echo "======== Release Main (fast-forward) ========"
	@CURRENT_BRANCH=$$(git branch --show-current) && \
	if [ -z "$$CURRENT_BRANCH" ]; then \
		echo "FAILED: Could not determine current branch" && exit 1; \
	fi && \
	if ! git diff-index --quiet HEAD --; then \
		echo "FAILED: Working tree has uncommitted changes" && exit 1; \
	fi && \
	git fetch origin && \
	git checkout main && \
	git pull --ff-only origin main && \
	git merge --ff-only origin/develop && \
	git push origin main && \
	git checkout "$$CURRENT_BRANCH"
	@echo "Main was fast-forwarded to origin/develop."
