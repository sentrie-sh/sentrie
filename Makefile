OUTPUT_DIR?=~/.local/bin
VERSION?=$(shell result=$$(git tag --sort=-version:refname | head -n 1); echo $${result:-0.0.1-dev})
COMMIT:=$(shell git rev-parse HEAD)
# Commit date in UTC, the same value GoReleaser stamps for releases.
DATE:=$(shell TZ=UTC git log -1 --format=%cd --date=format-local:%Y-%m-%dT%H:%M:%SZ)
TREE_STATE:=$(shell if [ -z "$$(git status --porcelain)" ]; then echo clean; else echo dirty; fi)

# A go.work above the checkout (for example in $$HOME) must not join the module
# graph: inside worktrees/<branch> it resolves packages to the primary checkout.
export GOWORK := off

build:
	go build -o $(OUTPUT_DIR)/sentrie -ldflags "-X main.builtBy=make -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.treeState=$(TREE_STATE) -X main.date=$(DATE) -s -w" .

clean-git:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then \
		echo "Error: Not on the main branch"; \
		exit 1; \
	fi

	git fetch -p
	git branch -vv
	git branch -vv | grep ': gone]' | awk '{print $$1}' | xargs -r git branch -D

# Conformance check for the knowledge graph under docs/codebase-graph.
# The schema it asserts is documented in docs/codebase-graph/overview.md.
docs-graph:
	go test -count=1 ./internal/docsgraph/

# Each recipe line runs in a fresh shell; env must be set on the same line as go generate.
gen:
	GIT_USER_NAME="$(shell git config user.name)" GIT_USER_EMAIL="$(shell git config user.email)" go generate ./...