.PHONY: all clean test build

COMPOSE_RUN = docker compose run --rm --quiet-pull
export DOCKER_CLI_HINTS=false

include make.d/*.mk