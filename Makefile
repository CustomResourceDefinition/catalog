.DEFAULT_GOAL := run
.PHONY: test

export DOCKER_CLI_HINTS=false

clean:
	@(docker inspect --type=image crd-runner:latest &>/dev/null && (docker rm -f $$(docker container ls -aqf "ancestor=crd-runner:latest") &>/dev/null && docker rmi -f crd-runner:latest &>/dev/null)) || true
	@rm -rf mounts/ephemeral &>/dev/null || true

build-image: clean
	@docker build -qt crd-runner . >/dev/null

build: build-image
	@docker container create --name crd-runner \
	-v ./mounts/ephemeral/helm/cache:/root/.cache/helm \
	-v ./mounts/ephemeral/helm/config:/root/.config/helm \
	-v ./mounts/ephemeral/helm/local:/root/.local/share/helm \
	-v ./mounts/ephemeral/templates:/templates \
	-v ./schema/:/schema \
	-v ./helm-charts.yaml:/app/configuration.yaml:ro \
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
	-v ./test/helm-charts.yaml:/app/configuration.yaml:ro \
	crd-runner >/dev/null
	@docker start crd-runner >/dev/null

shell: build
	@docker exec -it crd-runner /bin/sh

run: build
	@docker exec crd-runner /bin/sh /app/main.sh

test: test-happy-path test-only-latest

test-happy-path: build-test
	@docker exec crd-runner /bin/sh /app/test/prepare-helm.sh
	@docker exec crd-runner /bin/sh /app/test/prepare-charts.sh
	@docker exec crd-runner /bin/sh /app/test/prepare-schema.sh happy-path
	@docker exec crd-runner /bin/sh /app/main.sh
	@docker exec crd-runner /bin/sh /app/test/verify-test.sh "Happy path works"

test-only-latest: build-test
	@docker exec crd-runner /bin/sh /app/test/prepare-helm.sh
	@docker exec crd-runner /bin/sh /app/test/prepare-charts.sh
	@docker exec crd-runner /bin/sh /app/test/prepare-schema.sh only-latest
	@mkdir -p mounts/ephemeral/schema/chart.local/
	@mkdir -p mounts/ephemeral/schema/chart.new/
	@docker exec crd-runner /bin/sh /app/main.sh
	@docker exec crd-runner /bin/sh /app/test/verify-test.sh "Only lastest version used"
