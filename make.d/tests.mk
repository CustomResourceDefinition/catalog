GREEN='\e[1;32m%-6s\e[m\n'
TOOL_VERSION = $(shell grep '^golang ' .tool-versions | sed 's/golang //')
MOD_VERSION = $(shell grep '^go ' go.mod | sed 's/go //')

test: build test-go test-docker test-makefile test-editorcheck test-schemas unit-tests smoke-tests

test-go:
	@echo 'Checking for mod file changes ...'
	go mod tidy -diff

	@echo 'Checking for incorrectly formatted files ...'
	@test -z "$$(gofmt -l .)"

ifneq ($(TOOL_VERSION),$(MOD_VERSION))
	@echo 'Mismatched go versions'
	@exit 1
endif
	@printf $(GREEN) "OK"

unit-tests:
	@echo 'Running go unit tests ...'
	go test ./... -timeout 10s -shuffle on -p 1 -coverprofile=build/coverage.out -tags $(GO_TAGS) -skip 'TestCheckLocal'
	go tool cover -html=build/coverage.out -o build/coverage.html
	@echo 'Running static analysis ...'
	go vet ./...

test-schemas: build/schemastore/dependabot-2.0.json build/schemastore/github-workflow.json
	@echo 'Checking validity of certain files ...'
	build/bin/catalog verify --schema internal/configuration/schema.json --file configuration.yaml
	build/bin/catalog verify --schema internal/configuration/schema.json --file test/configuration.yaml
	build/bin/catalog verify --schema build/schemastore/dependabot-2.0.json --file .github/dependabot.yml
	build/bin/catalog verify --schema build/schemastore/github-workflow.json --file .github/workflows/tests.yaml
	build/bin/catalog verify --schema build/schemastore/github-workflow.json --file .github/workflows/scheduled-jobs.yaml

smoke-tests:
	@echo 'Prepare smoke test files ...'
	cat test/prepare-*.sh > build/bin/test-prepare
	cat test/verify-*.sh > build/bin/test-verify
	chmod +x build/bin/test-*

ifneq ($(strip $(CI)),)
	git config --global user.email "test@runner.local"
	git config --global user.name "Test Runner"
endif

	-find build/ephemeral/schema build/ephemeral/verified build/ephemeral/repository -not -name ".gitignore" -and -not -name ".gitkeep" -type f -delete
	-find build/ephemeral/schema build/ephemeral/verified build/ephemeral/repository -type d -empty -delete
	mkdir -p build/ephemeral/schema build/ephemeral/verified build/ephemeral/repository/http

	docker compose up --quiet-pull --wait -d registry nginx

	@echo 'Run first smoke test ...'
	build/bin/test-prepare all-versions build/ephemeral/schema build/ephemeral/verified build/ephemeral/repository
	HELM_OCI_PLAIN_HTTP=true build/bin/catalog update --configuration test/configuration.yaml --output build/ephemeral/schema --definitions build/ephemeral/schema

# FIXME: remove debugging
	find build/ephemeral/repository
	docker compose logs nginx
	docker compose ps

	build/bin/test-verify "Happy path works" build/ephemeral/schema build/ephemeral/verified
	@printf $(GREEN) "OK"

	-find build/ephemeral/schema build/ephemeral/verified build/ephemeral/repository -not -name ".gitignore" -and -not -name ".gitkeep" -type f -delete

	@echo 'Run second smoke test ...'
	build/bin/test-prepare only-latest build/ephemeral/schema build/ephemeral/verified build/ephemeral/repository
	HELM_OCI_PLAIN_HTTP=true build/bin/catalog update --configuration test/configuration.yaml --output build/ephemeral/schema --definitions build/ephemeral/schema
	build/bin/test-verify "Works using only latest version" build/ephemeral/schema build/ephemeral/verified
	@printf $(GREEN) "OK"

	docker compose down

test-editorcheck:
	@echo 'Checking general formatting of all files ...'
	$(COMPOSE_RUN) editorconfig ec -exclude '^schema/|^\.git/|.DS_Store|^build/remote/|^definitions/'
	@printf $(GREEN) "OK"

test-docker:
	@echo 'Checking formatting of docker-compose.yml ...'
	docker compose config -q
	@printf $(GREEN) "OK"

test-makefile:
	@echo 'Checking formatting of Makefile and *.mk files ...'
	$(COMPOSE_RUN) checkmake
	@printf $(GREEN) "OK"

build/schemastore/dependabot-2.0.json:
	@mkdir -p build/schemastore
	$(DOWNLOADER) $@ https://json.schemastore.org/dependabot-2.0.json

build/schemastore/github-workflow.json:
	@mkdir -p build/schemastore
	$(DOWNLOADER) $@ https://json.schemastore.org/github-workflow.json
