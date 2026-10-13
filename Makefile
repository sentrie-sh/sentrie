OUTPUT_DIR?=~/.local/bin
VERSION?=$(shell result=$$(git tag --sort=-version:refname | head -n 1); echo $${result:-0.0.1-dev})
COMMIT:=$(shell git rev-parse HEAD)
# Commit date in UTC, the same value GoReleaser stamps for releases.
DATE:=$(shell TZ=UTC git log -1 --format=%cd --date=format-local:%Y-%m-%dT%H:%M:%SZ)
TREE_STATE:=$(shell if [ -z "$$(git status --porcelain)" ]; then echo clean; else echo dirty; fi)
GOLANGCI_LINT?=golangci-lint

# A go.work above the checkout (for example in $$HOME) must not join the module
# graph: inside worktrees/<branch> it resolves packages to the primary checkout.
export GOWORK := off

.PHONY: build test lint docs-graph gen clean-git help

build: ## Build the sentrie binary into $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)/sentrie -ldflags "-X main.builtBy=make -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.treeState=$(TREE_STATE) -X main.date=$(DATE) -s -w" .

test: ## Run every test, including lang_test fixtures and the graph check
	go test ./...

lint: ## Run go vet and golangci-lint (CI pins v2.4.0)
	go vet ./...
	$(GOLANGCI_LINT) run --timeout=5m

# Conformance check for the knowledge graph under docs/codebase-graph.
# The schema it asserts is documented in docs/codebase-graph/overview.md.
docs-graph: ## Validate the codebase knowledge graph
	go test -count=1 ./internal/docsgraph/

# Each recipe line runs in a fresh shell; env must be set on the same line as go generate.
gen: ## Run go generate (needs a git identity for license headers)
	GIT_USER_NAME="$(shell git config user.name)" GIT_USER_EMAIL="$(shell git config user.email)" go generate ./...

clean-git: ## On main: fetch -p and delete local branches whose upstream is gone
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then \
		echo "Error: Not on the main branch"; \
		exit 1; \
	fi

	git fetch -p
	git branch -vv
	git branch -vv | grep ': gone]' | awk '{print $$1}' | xargs -r git branch -D

help: ## List targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "%-12s %s\n", $$1, $$2}'
