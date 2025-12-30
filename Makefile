OUTPUT_DIR?=~/.local/bin

build:

	go build -o $(OUTPUT_DIR)/sentrie -ldflags "-X main.builtWithMakefile=true" .

clean-git:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then \
		echo "Error: Not on the main branch"; \
		exit 1; \
	fi

	git fetch -p
	git branch -vv
	git branch -vv | grep ': gone]' | awk '{print $1}' | xargs git branch -D

