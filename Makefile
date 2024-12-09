.PHONY: all clean test

COMPOSE_RUN = docker compose run --rm --quiet-pull

GREEN='\e[1;32m%-6s\e[m\n'

export DOCKER_CLI_HINTS=false

clean: _clean
	docker compose down --remove-orphans --volumes --rmi local

update:
	$(COMPOSE_RUN) \
	-v $$(pwd)/schema:/schema \
	-v $$(pwd)/configuration.yaml:/app/configuration.yaml:ro \
	runner make _update

test: test-docker test-makefile test-editorcheck test-shellcheck
	$(COMPOSE_RUN) \
	-v $$(pwd)/build/ephemeral/schema:/schema \
	-v $$(pwd)/test/configuration.yaml:/app/configuration.yaml:ro \
	-v $$(pwd)/test/.env:/app/.env:ro \
	runner make _test

shell:
	$(COMPOSE_RUN) \
	-v $$(pwd)/build/ephemeral/schema:/schema \
	-v $$(pwd)/test/configuration.yaml:/app/configuration.yaml:ro \
	-v $$(pwd)/test/.env:/app/.env:ro \
	runner /bin/sh

_clean:
	rm -r build/ephemeral/schema/* &>/dev/null || true
	rm -r build/ephemeral/templates/* &>/dev/null || true
	rm -r build/bin &>/dev/null || true

_build: _clean
	@mkdir build/bin
	@mkdir build/ephemeral &>/dev/null || true
	cp -r src/helpers build/bin/
	cat src/*.sh > build/bin/main
	@chmod +x build/bin/main

_build-test: _build
	cat test/prepare-*.sh > build/bin/test-prepare
	cat test/verify-*.sh > build/bin/test-verify
	cat src/0?-*.sh test/unit-test-*.sh > build/bin/unit-tests
	@chmod +x build/bin/test-prepare build/bin/test-verify build/bin/unit-tests
	@yq -o json test/configuration.yaml | \
	jq --arg prefix build/ephemeral 'map(if .kind == "git" and (.repository | test("^/repository/")) then .repository = "\($$prefix)\(.repository)" else . end)' | \
	yq -p json -o yaml > build/configuration.yaml

_update: _build
	build/bin/main

_test: _unit-tests _test-configuration _test-all-versions _test-only-latest _test-schemas

_test-all-versions: _build-test
	build/bin/test-prepare all-versions
	build/bin/main
	build/bin/test-verify "Happy path works"
	@printf $(GREEN) "OK"

_test-only-latest: _build-test
	build/bin/test-prepare only-latest
	build/bin/main
	build/bin/test-verify "Only lastest version used"
	@printf $(GREEN) "OK"

_test-configuration: _clean
	@yq 'sort_by(.name)' configuration.yaml > build/configuration.sorted
	@diff configuration.yaml build/configuration.sorted
	check-jsonschema --schemafile .schema.json configuration.yaml
	check-jsonschema --schemafile .schema.json test/configuration.yaml
	@grep -q 'file://' configuration.yaml && exit 1 || true

_unit-tests: _build-test
	build/bin/unit-tests
	@printf $(GREEN) "OK"

test-shellcheck:
	$(COMPOSE_RUN) shellcheck shellcheck src/*.sh test/*.sh
	@printf $(GREEN) "OK"

_test-schemas:
	check-jsonschema --schemafile .schema.json configuration.yaml
	check-jsonschema --schemafile .schema.json test/configuration.yaml
	check-jsonschema --builtin-schema dependabot .github/dependabot.yml

test-editorcheck:
	$(COMPOSE_RUN) eclint ec -exclude '^schema/|^\.git/'
	@printf $(GREEN) "OK"

test-docker:
	$(COMPOSE_RUN) dockerlint --ignore DL3018 Dockerfile
	@printf $(GREEN) "OK"
	docker compose config -q
	@printf $(GREEN) "OK"

test-makefile:
	$(COMPOSE_RUN) makelint
	@printf $(GREEN) "OK"
