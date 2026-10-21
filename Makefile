# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ LAYOUT ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DEFAULT			= \033[0m
GREEN			= \033[1;32m
CYAN			= \033[1;36m
RED				= \033[1;31m

define ANIMATION
	@for frame in ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏; do \
		printf "$$frame $(1)...\r"; \
		sleep 0.02; \
	done; \
	printf "\n"
endef


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ VARIABLES ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

SERVER_PKG		= ./cmd/server
CLIENT_PKG		= ./cmd/client
GUI_PKG			= ./cmd/gui
BIN_DIR			= bin

GO				= go


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ RULES ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

install:
	$(call ANIMATION,Installing activity dependencies)
	@$(GO) mod download
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN_DIR)/tap-server $(SERVER_PKG)
	@$(GO) build -o $(BIN_DIR)/tap-client $(CLIENT_PKG)
	@$(GO) build -o $(BIN_DIR)/tap-gui $(GUI_PKG)
	@printf "$(GREEN)✓$(DEFAULT) Installed (binaries in $(BIN_DIR)/)\n"

run-server:
	@$(GO) run $(SERVER_PKG) $(ARGS)

run-client:
	@$(GO) run $(CLIENT_PKG) $(ARGS)

run-client-gui:
	@$(GO) run $(GUI_PKG) $(ARGS)

clean:
	$(call ANIMATION,Removing binaries and caches)
	@rm -rf $(BIN_DIR) coverage.out
	@$(GO) clean -testcache
	@printf "$(GREEN)✓$(DEFAULT) Cleaned\n"

lint:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		printf "$(RED)✗$(DEFAULT) Not gofmt-formatted:\n$$unformatted\n"; \
		exit 1; \
	fi
	@$(GO) vet ./...

lint-strict: lint
	@command -v golangci-lint >/dev/null 2>&1 || { \
		printf "$(RED)✗$(DEFAULT) golangci-lint is not installed\n"; \
		exit 1; \
	}
	@golangci-lint run ./...

test:
	$(call ANIMATION,Running tests)
	@$(GO) test -race -v ./...


help:
	@printf "Usage: make [target]\n"
	@printf "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	@printf "$(CYAN)install$(DEFAULT)		- Download dependencies and build server, CLI and GUI into $(BIN_DIR)/\n"
	@printf "$(CYAN)run-server$(DEFAULT)		- Run the server (use ARGS=\"<args>\" to pass arguments)\n"
	@printf "$(CYAN)run-client$(DEFAULT)		- Run the CLI client (use ARGS=\"<args>\" to pass arguments)\n"
	@printf "$(CYAN)run-client-gui$(DEFAULT)	- Run the GUI client (use ARGS=\"<args>\" to pass arguments)\n"
	@printf "$(CYAN)clean$(DEFAULT)		- Remove built binaries and test cache\n"
	@printf "$(CYAN)lint$(DEFAULT)		- Check gofmt formatting and run go vet\n"
	@printf "$(CYAN)lint-strict$(DEFAULT)	- Run lint, then golangci-lint\n"
	@printf "$(CYAN)test$(DEFAULT)		- Run tests with the race detector\n"
	@printf "$(CYAN)help$(DEFAULT)		- Show this help message\n"

.PHONY:			install run-server run-client run-client-gui clean lint lint-strict test help
