# Makefile

FRONTEND_DIR := frontend
BACKEND_DIR := backend
PLUGIN_DIR := plugin
ANDROID_DIR := android
MOBILE_DIR := mobile
DIST_DIR := $(FRONTEND_DIR)/dist
OUTPUT_DIR := build
DOWNLOADS_DIR := downloads

BACKEND_COVERAGE_MIN := 80
BACKEND_GO_CACHE_DIR := $(abspath .cache/go-build)
BACKEND_GO_MOD_CACHE_DIR := $(abspath .cache/go-mod)
BACKEND_GO_ENV := GOCACHE=$(BACKEND_GO_CACHE_DIR) GOMODCACHE=$(BACKEND_GO_MOD_CACHE_DIR)

# Host-specific JDK / Android SDK for the gradle gates. Override via env or
# Makefile.local. JAVA_HOME falls back to the first JDK under ~/.local/jdks.
ANDROID_SDK_DIR ?= $(HOME)/Android/Sdk
JAVA_HOME ?= $(firstword $(wildcard $(HOME)/.local/jdks/jdk-*))

.PHONY: all frontend backend move clean deploy ci ci-frontend ci-backend ci-plugin ci-android ci-mobile gradle-ci lint-frontend test-frontend lint-backend test-backend prepare-backend-go-cache release-main-ff

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
	@if [ -d "$(DOWNLOADS_DIR)" ]; then \
		echo "Bundling client apps from $(DOWNLOADS_DIR)/..."; \
		cp -r $(DOWNLOADS_DIR) $(OUTPUT_DIR); \
	fi
	@echo "Files moved."

clean:
	@echo "Cleaning..."
	@rm -rf $(OUTPUT_DIR)
	@echo "Clean complete."

ci: ci-frontend ci-backend ci-plugin ci-android ci-mobile
	@echo ""
	@echo "========================================"
	@echo "  All quality gates passed"
	@echo "========================================"

ci-frontend: lint-frontend test-frontend

ci-backend: lint-backend test-backend

ci-plugin:
	@echo ""
	@echo "======== Plugin Lint + Tests ========"
	@cd $(PLUGIN_DIR) && npm ci && npm run lint && npm test
	@echo "Plugin quality gate passed."

ci-android:
	@$(MAKE) gradle-ci GRADLE_DIR=$(ANDROID_DIR) GRADLE_LABEL=Android

ci-mobile:
	@$(MAKE) gradle-ci GRADLE_DIR=$(MOBILE_DIR) GRADLE_LABEL=Mobile

gradle-ci:
	@echo ""
	@echo "======== $(GRADLE_LABEL) Tests + Build ========"
	@if [ -z "$(JAVA_HOME)" ]; then \
		echo "FAILED: JAVA_HOME is not set and no JDK found under ~/.local/jdks."; \
		echo "Set JAVA_HOME (env or Makefile.local) before running the gradle gates."; \
		exit 1; \
	fi
	@[ -f $(GRADLE_DIR)/local.properties ] || echo "sdk.dir=$(ANDROID_SDK_DIR)" > $(GRADLE_DIR)/local.properties
	@cd $(GRADLE_DIR) && JAVA_HOME="$(JAVA_HOME)" ./gradlew --no-daemon test assembleDebug
	@echo "$(GRADLE_LABEL) quality gate passed."

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
	@cd $(BACKEND_DIR) && $(BACKEND_GO_ENV) go vet -tags=dev ./...
	@echo "go vet passed."

test-backend:
	@$(MAKE) prepare-backend-go-cache
	@echo ""
	@echo "======== Backend Tests + Coverage ========"
	@cd $(BACKEND_DIR) && $(BACKEND_GO_ENV) go test -tags=dev ./... -coverprofile=coverage.out
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

release-main-ff-old:
	@echo ""
	@echo "======== Release Main (automated sync + fast-forward) ========"
	@CURRENT_BRANCH=$$(git branch --show-current); \
	if [ -z "$$CURRENT_BRANCH" ]; then \
		echo "FAILED: Could not determine current branch"; \
		exit 1; \
	fi; \
	if ! git diff-index --quiet HEAD --; then \
		echo "FAILED: Working tree has uncommitted changes"; \
		exit 1; \
	fi; \
	cleanup() { \
		if git rev-parse --verify -q MERGE_HEAD >/dev/null; then \
			git merge --abort >/dev/null 2>&1 || true; \
		fi; \
		ACTIVE_BRANCH=$$(git branch --show-current); \
		if [ -n "$$ACTIVE_BRANCH" ] && [ "$$ACTIVE_BRANCH" != "$$CURRENT_BRANCH" ]; then \
			git checkout "$$CURRENT_BRANCH" >/dev/null 2>&1 || true; \
		fi; \
	}; \
	trap cleanup EXIT; \
	git fetch origin; \
	git checkout develop; \
	git pull --ff-only origin develop; \
	if ! git merge --no-edit origin/main; then \
		echo "FAILED: Could not merge origin/main into develop automatically."; \
		echo "Resolve conflicts in develop and retry."; \
		exit 1; \
	fi; \
	git push origin develop; \
	git checkout main; \
	git pull --ff-only origin main; \
	git merge --ff-only origin/develop; \
	git push origin main
	@echo "Develop and main were updated successfully."

release-main-ff:
	@echo ""
	@echo "======== Sync main with develop (fast-forward) ========"
	@if ! git diff-index --quiet HEAD -- || ! git diff --cached --quiet; then \
		echo "FAILED: Working tree has uncommitted changes" && exit 1; \
	fi
	git fetch origin
	git checkout main
	git pull --ff-only origin main
	git merge --ff-only origin/develop
	git push origin main
	git checkout develop
	@echo ""
	@echo "Done: main and develop point to the same commit. Auto Tag cuts the tag on push."
