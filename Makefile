.DEFAULT_GOAL := run
.PHONY: test run

export DOCKER_CLI_HINTS=false

clean:
	@rm src/helpers/convert.py &>/dev/null || true
	@rm -r build/* &>/dev/null || true

fetch-converter:
	@wget -q -O src/helpers/convert.py https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py

build: fetch-converter
	@mkdir build/bin
	@mkdir build/ephemeral
	@cp -r src/helpers build/bin/
	@cat src/*.sh > build/bin/main
	@chmod +x build/bin/main

build-test: build
	@cat test/prepare-*.sh > build/bin/test-prepare
	@cat test/verify-*.sh > build/bin/test-verify
	@chmod +x build/bin/test-prepare build/bin/test-verify
	yq '[.[] | select(.kind == "git" and (.repository | test("^/repository/")))]' test/configuration.yaml > build/configuration.yaml

run: clean build
	@build/bin/main configuration.yaml schema build/ephemeral

test: clean build-test test-all-versions test-only-latest

test-all-versions: build-test
	@build/bin/test-prepare all-versions build/ephemeral/schema build/ephemeral
	@build/bin/main build/configuration.yaml build/ephemeral/schema build/ephemeral
	@build/bin/test-verify all-versions build/ephemeral/schema build/ephemeral

test-only-latest: build-test
	@build/bin/test-prepare only-latest build/ephemeral/schema build/ephemeral
	@build/bin/main build/configuration.yaml build/ephemeral/schema build/ephemeral
	@build/bin/test-verify only-latest build/ephemeral/schema build/ephemeral

shell: clean-shell
	@docker build -qt crd-runner . >/dev/null
	@docker run -it -v $$(pwd):/workspace -w /workspace crd-runner /bin/sh

clean-shell: clean
	@(docker inspect --type=image crd-runner:latest &>/dev/null && (docker rm -f $$(docker container ls -aqf "ancestor=crd-runner:latest") &>/dev/null && docker rmi -f crd-runner:latest &>/dev/null)) || true
