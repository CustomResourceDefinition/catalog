GREEN='\e[1;32m%-6s\e[m\n'
TOOL_VERSION = $(shell grep '^golang ' .tool-versions | sed 's/golang //')
MOD_VERSION = $(shell grep '^go ' go.mod | sed 's/go //')

test: build test-docker test-makefile test-editorcheck
	$(COMPOSE_RUN) \
	-v $$(pwd)/build/ephemeral/schema:/schema \
	-v $$(pwd)/test/configuration.yaml:/app/configuration.yaml:ro \
	runner make _test

_test: _test-configuration
	@yq -o json test/configuration.yaml | \
		jq --arg prefix build/ephemeral 'map(if .kind == "git" and (.repository | test("^/repository/")) then .repository = "\($$prefix)\(.repository)" else . end)' | \
		yq -p json -o yaml > build/configuration.yaml
	go mod tidy -diff
	@test -z "$$(gofmt -l .)"
ifneq ($(TOOL_VERSION),$(MOD_VERSION))
	@echo 'Mismatched go versions'
	@exit 1
endif
	@exit 0
	@printf $(GREEN) "OK"
	$(MAKE) _unit-tests

_test-configuration:
	@yq 'sort_by(.name)' configuration.yaml > build/configuration.sorted
	@diff configuration.yaml build/configuration.sorted
	@grep -q 'file://' configuration.yaml && exit 1 || true

_unit-tests:
	go test ./... -timeout 10s -shuffle on -p 1 -coverprofile=build/coverage.out -tags containers_image_openpgp
	go tool cover -html=build/coverage.out -o build/coverage.html
	go vet ./...

test-editorcheck:
	$(COMPOSE_RUN) editorconfig ec -exclude '^schema/|^\.git/'
	@printf $(GREEN) "OK"

test-docker:
	$(COMPOSE_RUN) hadolint --ignore DL3018 Dockerfile
	@printf $(GREEN) "OK"
	docker compose config -q
	@printf $(GREEN) "OK"

test-makefile:
	$(COMPOSE_RUN) checkmake
	@printf $(GREEN) "OK"
