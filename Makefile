# Makefile

FRONTEND_DIR := frontend
BACKEND_DIR := backend
DIST_DIR := $(FRONTEND_DIR)/dist
OUTPUT_DIR := build


.PHONY: all frontend backend move clean deploy

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

-include Makefile.local

deploy:
	@echo "Calling local deploy..."
	@$(MAKE) -f Makefile.local deploy
	@echo "Local deploy complete."
