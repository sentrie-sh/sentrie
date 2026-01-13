OUTPUT_DIR?=~/.local/bin
VERSION?=$(shell result=$$(git tag --sort=-version:refname | head -n 1); echo $${result:-0.0.1-dev})

build:
	go build -o $(OUTPUT_DIR)/sentrie -ldflags "-X main.builtBy=make -X main.version=$(VERSION) -X main.commit=$(shell git rev-parse HEAD) -X main.treeState=dirty -X main.date=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ') -s -w" .

clean-git:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then \
		echo "Error: Not on the main branch"; \
		exit 1; \
	fi

	git fetch -p
	git branch -vv
	git branch -vv | grep ': gone]' | awk '{print $1}' | xargs git branch -D

