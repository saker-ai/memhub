.PHONY: all build run web embed server clean dev

APP_NAME    := memoh-server
CMD_PATH    := ./cmd/agent
OUTPUT      := ./$(APP_NAME)
WEB_DIR     := apps/web
EMBED_DIR   := internal/embedded/web
DIST_DIR    := $(WEB_DIR)/dist

VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS     := -s -w \
	-X github.com/memohai/memoh/internal/version.Version=$(VERSION) \
	-X github.com/memohai/memoh/internal/version.CommitHash=$(COMMIT_HASH) \
	-X github.com/memohai/memoh/internal/version.BuildTime=$(BUILD_TIME)

all: build

# Install frontend dependencies and build
web: node_modules
	pnpm --filter @memohai/web build

node_modules: pnpm-lock.yaml
	pnpm install
	@touch node_modules

# Gzip frontend dist and copy to embed directory
embed: web
	@rm -rf $(EMBED_DIR)/assets $(EMBED_DIR)/*.gz
	@mkdir -p $(EMBED_DIR)/assets
	@if [ -f $(DIST_DIR)/index.html ]; then \
		gzip -9 -k $(DIST_DIR)/index.html && mv $(DIST_DIR)/index.html.gz $(EMBED_DIR)/index.html.gz; \
	fi
	@if [ -f $(DIST_DIR)/logo.png ]; then \
		gzip -9 -k $(DIST_DIR)/logo.png && mv $(DIST_DIR)/logo.png.gz $(EMBED_DIR)/logo.png.gz; \
	fi
	@find $(DIST_DIR)/assets -type f 2>/dev/null | while read f; do \
		rel=$${f#$(DIST_DIR)/assets/}; \
		gzip -9 -k "$$f" && mv "$$f.gz" "$(EMBED_DIR)/assets/$$rel.gz"; \
	done
	@echo "Embedded web assets prepared."

# Build Go server binary
server: embed
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(OUTPUT) $(CMD_PATH)

# Build everything (frontend + backend)
build: server

# Build and run the server
run: build
	$(OUTPUT)

# Dev mode: run backend only (without embedding frontend, use vite dev server separately)
dev:
	go run $(CMD_PATH)

# Clean build artifacts
clean:
	rm -f $(OUTPUT)
	rm -rf $(EMBED_DIR)/assets $(EMBED_DIR)/*.gz
	rm -rf $(DIST_DIR)
