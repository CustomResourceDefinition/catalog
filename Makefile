.DEFAULT_GOAL := run
.PHONY: test ci-test run

export DOCKER_CLI_HINTS=false

clean:
	@(docker inspect --type=image crd-runner:latest &>/dev/null && (docker rm -f $$(docker container ls -aqf "ancestor=crd-runner:latest") &>/dev/null && docker rmi -f crd-runner:latest &>/dev/null)) || true
	@rm -rf mounts/ephemeral &>/dev/null || true

build-image:
	@docker build -qt crd-runner . >/dev/null

build: build-image
	@docker container create --name crd-runner \
	-v ./mounts/ephemeral/helm/cache:/root/.cache/helm \
	-v ./mounts/ephemeral/helm/config:/root/.config/helm \
	-v ./mounts/ephemeral/helm/local:/root/.local/share/helm \
	-v ./mounts/ephemeral/templates:/templates \
	-v ./schema/:/schema \
	-v ./manifest-uris.yaml:/app/manifest-uris.yaml:ro \
	-v ./helm-charts.yaml:/app/helm-charts.yaml:ro \
	-v ./oci-charts.yaml:/app/oci-charts.yaml:ro \
	-v ./git-charts.yaml:/app/git-charts.yaml:ro \
	crd-runner >/dev/null
	@docker start crd-runner >/dev/null

build-test: build-image
	@docker container create --name crd-runner \
	-v ./mounts/ephemeral/helm/cache:/root/.cache/helm \
	-v ./mounts/ephemeral/helm/config:/root/.config/helm \
	-v ./mounts/ephemeral/helm/local:/root/.local/share/helm \
	-v ./mounts/ephemeral/helm/templates:/templates \
	-v ./mounts/ephemeral/repository/:/repository \
	-v ./mounts/ephemeral/schema/:/schema \
	-v ./mounts/verified-schemas/:/verified-schemas:ro \
	-v ./mounts/chart/:/chart:ro \
	-v ./mounts/chart-value-based/:/chart-value-based:ro \
	-v ./test/manifest-uris.yaml:/app/manifest-uris.yaml:ro \
	-v ./test/helm-charts.yaml:/app/helm-charts.yaml:ro \
	-v ./test/oci-charts.yaml:/app/oci-charts.yaml:ro \
	-v ./test/git-charts.yaml:/app/git-charts.yaml:ro \
	crd-runner >/dev/null
	@docker start crd-runner >/dev/null

shell: clean build
	@docker exec -it crd-runner /bin/sh

run: clean ci-run

ci-run: build
	@docker exec crd-runner /bin/sh /app/main.sh

test: clean ci-test

ci-test: test-happy-path test-only-latest

test-happy-path: build-test
	@docker exec crd-runner /bin/sh /app/test/prepare.sh happy-path
	@docker exec crd-runner /bin/sh /app/main.sh
	@docker exec crd-runner /bin/sh /app/test/verify.sh "Happy path works"

test-only-latest: build-test
	@docker exec crd-runner /bin/sh /app/test/prepare.sh only-latest
	@docker exec crd-runner /bin/sh /app/main.sh
	@docker exec crd-runner /bin/sh /app/test/verify.sh "Only lastest version used"
