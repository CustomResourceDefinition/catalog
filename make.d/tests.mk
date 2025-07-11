GREEN='\e[1;32m%-6s\e[m\n'
TOOL_VERSION = $(shell grep '^golang ' .tool-versions | sed 's/golang //')
MOD_VERSION = $(shell grep '^go ' go.mod | sed 's/go //')

test: build test-docker test-makefile test-editorcheck
	$(COMPOSE_RUN) \
	-v $$(pwd)/build/ephemeral/schema:/schema \
	-v $$(pwd)/test/configuration.yaml:/app/configuration.yaml:ro \
	-v $$(pwd)/test:/app/test \
	runner make _test

_test: _test-schemas
	go mod tidy -diff
	@test -z "$$(gofmt -l .)"
ifneq ($(TOOL_VERSION),$(MOD_VERSION))
	@echo 'Mismatched go versions'
	@exit 1
endif
	@printf $(GREEN) "OK"
	$(MAKE) _unit-tests
	$(MAKE) _smoke-test

_unit-tests:
	go test ./... -timeout 10s -shuffle on -p 1 -coverprofile=build/coverage.out -tags $(GO_TAGS)
	go tool cover -html=build/coverage.out -o build/coverage.html
	go vet ./...

_test-schemas:
	build/bin/catalog verify --schema internal/configuration/schema.json --file configuration.yaml
	build/bin/catalog verify --schema internal/configuration/schema.json --file test/configuration.yaml
	build/bin/catalog verify --schema /opt/schemastore/dependabot-2.0.json --file .github/dependabot.yml
	build/bin/catalog verify --schema /opt/schemastore/github-workflow.json --file .github/workflows/tests.yaml
	build/bin/catalog verify --schema /opt/schemastore/github-workflow.json --file .github/workflows/scheduled-jobs.yaml

_smoke-test:
	cat test/prepare-*.sh > build/bin/test-prepare
	cat test/verify-*.sh > build/bin/test-verify

	sh build/bin/test-prepare all-versions
	HELM_OCI_PLAIN_HTTP=true build/bin/catalog update --configuration test/configuration.yaml --output build/ephemeral/schema
	sh build/bin/test-verify "Happy path works"
	@printf $(GREEN) "OK"

	sh build/bin/test-prepare only-latest
	HELM_OCI_PLAIN_HTTP=true build/bin/catalog update --configuration test/configuration.yaml --output build/ephemeral/schema
	sh build/bin/test-verify "Works using only latest version"
	@printf $(GREEN) "OK"

test-editorcheck:
	$(COMPOSE_RUN) editorconfig ec -exclude '^schema/|^\.git/|.DS_Store'
	@printf $(GREEN) "OK"

test-docker:
	$(COMPOSE_RUN) hadolint --ignore DL3018 Dockerfile
	@printf $(GREEN) "OK"
	docker compose config -q
	@printf $(GREEN) "OK"

test-makefile:
	$(COMPOSE_RUN) checkmake
	@printf $(GREEN) "OK"
