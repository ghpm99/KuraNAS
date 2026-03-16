# Makefile

FRONTEND_DIR := frontend
BACKEND_DIR := backend
DIST_DIR := $(FRONTEND_DIR)/dist
OUTPUT_DIR := build


BACKEND_COVERAGE_MIN := 80

.PHONY: all frontend backend move clean deploy ci ci-frontend ci-backend lint-frontend test-frontend lint-backend test-backend

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

lint-backend:
	@echo ""
	@echo "======== Backend Lint (gofmt) ========"
	@cd $(BACKEND_DIR) && BADFILES=$$(gofmt -l . | while IFS= read -r f; do \
		diff <(gofmt "$$f") <(tr -d '\r' < "$$f") >/dev/null 2>&1 || echo "$$f"; \
	done) && \
	if [ -n "$$BADFILES" ]; then \
		echo "gofmt check failed on:" && echo "$$BADFILES" && exit 1; \
	fi
	@echo "gofmt passed."
	@echo ""
	@echo "======== Backend Lint (go vet) ========"
	@cd $(BACKEND_DIR) && go vet ./...
	@echo "go vet passed."

test-backend:
	@echo ""
	@echo "======== Backend Tests + Coverage ========"
	@cd $(BACKEND_DIR) && go test ./... -coverprofile=coverage.out
	@echo ""
	@echo "======== Backend Coverage Threshold ========"
	@cd $(BACKEND_DIR) && \
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
