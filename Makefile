.PHONY: all clean test build

export DOCKER_CLI_HINTS=false

DOWNLOADER := $(shell \
    if command -v curl >/dev/null 2>&1; then \
        echo "curl -fsSL -o"; \
    elif command -v wget >/dev/null 2>&1; then \
        echo "wget -qO"; \
    fi \
)

include make.d/*.mk

all: clean help build check comparison tags test schema-check update