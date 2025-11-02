OUTPUT_DIR?=~/.local/bin

build:
	$(eval TIMESTAMP := $(shell date +%Y%m%d%H%M%S))
	$(eval GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD | sed 's/[\/_]/-/g' | sed 's/[^a-zA-Z0-9.-]//g'))

	go build -o $(OUTPUT_DIR)/sentrie -ldflags "-X main.version=0.0.0-dev-$(GIT_BRANCH).$(TIMESTAMP)" .

clean-git:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then \
		echo "Error: Not on the main branch"; \
		exit 1; \
	fi

	git fetch -p
	git branch -vv
	git branch -vv | grep ': gone]' | awk '{print $1}' | xargs git branch -D

