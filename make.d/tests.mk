GREEN='\e[1;32m%-6s\e[m\n'
TOOL_VERSION = $(shell grep '^golang ' .tool-versions | sed 's/golang //')
MOD_VERSION = $(shell grep '^go ' tools/go.mod | sed 's/go //')

test: build test-docker test-makefile test-editorcheck test-shellcheck
	$(COMPOSE_RUN) \
	-v $$(pwd)/build/ephemeral/schema:/schema \
	-v $$(pwd)/test/configuration.yaml:/app/configuration.yaml:ro \
	-v $$(pwd)/test/.env:/app/.env:ro \
	runner make _test

_build-test:
	@yq -o json test/configuration.yaml | \
		jq --arg prefix build/ephemeral 'map(if .kind == "git" and (.repository | test("^/repository/")) then .repository = "\($$prefix)\(.repository)" else . end)' | \
		yq -p json -o yaml > build/configuration.yaml
	cat test/prepare-*.sh > build/bin/test-prepare
	cat test/verify-*.sh > build/bin/test-verify
	cat src/0?-*.sh test/unit-test-*.sh > build/bin/unit-tests

_test: _unit-tests _test-configuration _test-all-versions _test-only-latest _test-schemas _test-tools

_test-all-versions: _build-test
	sh build/bin/test-prepare all-versions
	build/bin/main
	sh build/bin/test-verify "Happy path works"
	@printf $(GREEN) "OK"

_test-only-latest: _build-test
	sh build/bin/test-prepare only-latest
	build/bin/main
	sh build/bin/test-verify "Only latest version used"
	@printf $(GREEN) "OK"

_test-configuration:
	@yq 'sort_by(.name)' configuration.yaml > build/configuration.sorted
	@diff configuration.yaml build/configuration.sorted
	check-jsonschema --schemafile .schema.json configuration.yaml
	check-jsonschema --schemafile .schema.json test/configuration.yaml
	@grep -q 'file://' configuration.yaml && exit 1 || true

_test-tools:
	@printf 'Running tool tests:\n'
	cd tools && test -z "$$(gofmt -l .)"
	cd tools && go mod tidy -diff
	cd tools && \
		go test -timeout 10s -shuffle on -p 1 -coverprofile=../build/coverage.out && \
		go tool cover -html=../build/coverage.out -o ../build/coverage.html
	cd tools && go vet ./...
ifneq ($(TOOL_VERSION),$(MOD_VERSION))
	@echo 'Mismatched go versions'
	@exit 1
endif
	@exit 0
	@printf $(GREEN) "OK"

_unit-tests: _build-test
	sh build/bin/unit-tests
	@printf $(GREEN) "OK"

test-shellcheck:
	$(COMPOSE_RUN) shellcheck shellcheck src/*.sh test/*.sh
	@printf $(GREEN) "OK"

_test-schemas:
	check-jsonschema --schemafile .schema.json configuration.yaml
	check-jsonschema --schemafile .schema.json test/configuration.yaml
	check-jsonschema --builtin-schema dependabot .github/dependabot.yml

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
