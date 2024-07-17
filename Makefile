.DEFAULT_GOAL := clean
.PHONY: test

export DOCKER_CLI_HINTS=false

CONFIGURATION=
TEMPLATES=./build/ephemeral/templates
SCHEMA=
CI_COMMAND=
ANCESTOR := $(shell docker container ls -aqf "ancestor=crd-runner:latest" 2>/dev/null)

clean:
	@rm -r build/ephemeral/schema/* &>/dev/null || true
	@rm -r build/ephemeral/templates/* &>/dev/null || true
	@rm -r build/bin &>/dev/null || true
	@rm src/helpers/convert.py &>/dev/null || true

configure:
	@$(eval CONFIGURATION=./configuration.yaml)
	@$(eval SCHEMA=./schema/)
	@$(eval CI_COMMAND=run)

configure-test:
	@$(eval CONFIGURATION=./test/configuration.yaml)
	@$(eval SCHEMA=./build/ephemeral/schema/)
	@$(eval CI_COMMAND=test)

build: clean configure
	@wget -qO src/helpers/convert.py https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py
	@mkdir build/bin
	@mkdir build/ephemeral &>/dev/null || true
	@cp -r src/helpers build/bin/
	@cat src/*.sh > build/bin/main
	@chmod +x build/bin/main

build-test: configure-test build
	@cat test/prepare-*.sh > build/bin/test-prepare
	@cat test/verify-*.sh > build/bin/test-verify
	@cat src/0?-*.sh test/unit-test-*.sh > build/bin/unit-tests
	@chmod +x build/bin/test-prepare build/bin/test-verify build/bin/unit-tests
	@yq -o json test/configuration.yaml | \
	jq --arg prefix build/ephemeral 'map(if .kind == "git" and (.repository | test("^/repository/")) then .repository = "\($$prefix)\(.repository)" else . end)' | \
	yq -p json -o yaml > build/configuration.yaml

run: build
	@build/bin/main

test: unit-tests test-configuration test-all-versions test-only-latest test-shellcheck

test-all-versions: build-test
	@build/bin/test-prepare all-versions
	@build/bin/main
	@build/bin/test-verify "Happy path works"

test-only-latest: build-test
	@build/bin/test-prepare only-latest
	@build/bin/main
	@build/bin/test-verify "Only lastest version used"

test-configuration: clean
	@yq 'sort_by(.name)' configuration.yaml > build/configuration.sorted
	@diff configuration.yaml build/configuration.sorted
	@check-jsonschema --schemafile .schema.json configuration.yaml
	@check-jsonschema --schemafile .schema.json test/configuration.yaml
	@grep -q 'file://' configuration.yaml && exit 1 || true

test-shellcheck:
	@find src test -type f -name "*.sh" -print0 | sort -z | xargs -0 -I {} shellcheck {}

unit-tests: build-test
	@build/bin/unit-tests

ci-test: configure-test shell-command

ci-run: configure shell-command

build-shell:
ifneq ($(ANCESTOR),)
	@docker rm -f $(ANCESTOR) &>/dev/null
	@docker rmi -f crd-runner:latest &>/dev/null
endif
	@docker build -qt crd-runner . >/dev/null

shell: configure-test build-shell
	@docker run -it \
	-v $(TEMPLATES):/templates \
	-v $(SCHEMA):/schema \
	-v $(CONFIGURATION):/app/configuration.yaml:ro \
	-v $$(pwd)/src/helpers:/app/helpers:ro \
	-v $$(pwd)/test/fixtures:/app/test/fixtures:ro \
	-v $$(pwd):/workspace \
	-w /workspace \
	crd-runner /bin/sh

shell-command: build-shell
	@docker run \
	-v $(TEMPLATES):/templates \
	-v $(SCHEMA):/schema \
	-v $(CONFIGURATION):/app/configuration.yaml:ro \
	-v $$(pwd)/src/helpers:/app/helpers:ro \
	-v $$(pwd)/test/fixtures:/app/test/fixtures:ro \
	-v $$(pwd):/workspace \
	-w /workspace \
	crd-runner /bin/sh -c 'make $(CI_COMMAND)'