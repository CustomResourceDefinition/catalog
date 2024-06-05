.DEFAULT_GOAL := run
.PHONY: test ci-test run

export DOCKER_CLI_HINTS=false

clean:
	@(docker inspect --type=image crd-runner:latest &>/dev/null && (docker rm -f $$(docker container ls -aqf "ancestor=crd-runner:latest") &>/dev/null && docker rmi -f crd-runner:latest &>/dev/null)) || true
	@rm -r test/ephemeral &>/dev/null || true
	@rm src/helpers/convert.py &>/dev/null || true

fetch-converter:
	@wget -q -O src/helpers/convert.py https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py

build-image: fetch-converter
	@docker build -qt crd-runner . >/dev/null

build: build-image
	@docker container create --name crd-runner \
	-v ./test/ephemeral/helm/cache:/root/.cache/helm \
	-v ./test/ephemeral/helm/config:/root/.config/helm \
	-v ./test/ephemeral/helm/local:/root/.local/share/helm \
	-v ./test/ephemeral/templates:/templates \
	-v ./schema/:/schema \
	-v ./configuration.yaml:/app/configuration.yaml:ro \
	crd-runner >/dev/null
	@docker start crd-runner >/dev/null

build-test: build-image
	@docker container create --name crd-runner \
	-v ./test/ephemeral/helm/cache:/root/.cache/helm \
	-v ./test/ephemeral/helm/config:/root/.config/helm \
	-v ./test/ephemeral/helm/local:/root/.local/share/helm \
	-v ./test/ephemeral/templates:/templates \
	-v ./test/ephemeral/schema/:/schema \
	-v ./test/configuration.yaml:/app/configuration.yaml:ro \
	crd-runner >/dev/null
	@docker start crd-runner >/dev/null

shell: clean build
	@docker exec -it crd-runner /bin/sh

run: clean ci-run

ci-run: build
	@docker exec crd-runner /bin/sh /app/main.sh

test: clean ci-test

ci-test: test-all-versions test-only-latest

test-all-versions: build-test
	@docker exec crd-runner /bin/sh /app/test/prepare.sh all-versions
	@docker exec crd-runner /bin/sh /app/main.sh
	@docker exec crd-runner /bin/sh /app/test/verify.sh "Happy path works"

test-only-latest: build-test
	@docker exec crd-runner /bin/sh /app/test/prepare.sh only-latest
	@docker exec crd-runner /bin/sh /app/main.sh
	@docker exec crd-runner /bin/sh /app/test/verify.sh "Only lastest version used"
