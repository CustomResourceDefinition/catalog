.DEFAULT_GOAL := run
.PHONY: test

export DOCKER_CLI_HINTS=false

clean:
	@(docker inspect --type=image crd-runner:latest &>/dev/null && (docker rm -f $$(docker container ls -aqf "ancestor=crd-runner:latest") &>/dev/null && docker rmi -f crd-runner:latest &>/dev/null)) || true
	@cd mounts/; find helm/ -type d -mindepth 2 -maxdepth 2 -print0 | xargs -0 -I{} sh -c 'rm -r "{}"' || true
	@cd mounts/; find schema -type d -mindepth 1 -maxdepth 1 -print0 | xargs -0 -I{} sh -c 'rm -r "{}"' || true
	@find mounts/helm/config mounts/repository -type f ! -name .gitkeep ! -name .gitignore -delete

build-image: clean
	@docker build -qt crd-runner . >/dev/null

build: build-image
	@docker container create --name crd-runner \
	-v ./mounts/helm/cache:/root/.cache/helm \
	-v ./mounts/helm/config:/root/.config/helm \
	-v ./mounts/helm/local:/root/.local/share/helm \
	-v ./mounts/helm/templates:/templates \
	-v ./helm-charts.yaml:/app/configuration.yaml \
	crd-runner >/dev/null
	@docker start crd-runner >/dev/null

build-test: build-image
	@docker container create --name crd-runner \
	-v ./mounts/helm/cache:/root/.cache/helm \
	-v ./mounts/helm/config:/root/.config/helm \
	-v ./mounts/helm/local:/root/.local/share/helm \
	-v ./mounts/helm/templates:/templates \
	-v ./mounts/repository/:/repository \
	-v ./mounts/schema/:/schema \
	-v ./mounts/verified-schemas/:/verified-schemas \
	-v ./mounts/chart/:/chart \
	-v ./test/helm-charts.yaml:/app/configuration.yaml \
	crd-runner >/dev/null
	@docker start crd-runner >/dev/null

shell: clean build
	@docker exec -it crd-runner /bin/sh

run: build
	@docker exec crd-runner /bin/sh /app/main.sh

test: test-happy-path test-only-latest

test-happy-path: build-test
	@docker exec crd-runner /bin/sh /app/test/prepare-chart.sh
	@docker exec crd-runner /bin/sh /app/test/prepare-schema.sh happy-path
	@docker exec crd-runner /bin/sh /app/main.sh
	@docker exec crd-runner /bin/sh /app/test/verify-test.sh "Happy path works"

test-only-latest: build-test
	@docker exec crd-runner /bin/sh /app/test/prepare-chart.sh only-latest
	@docker exec crd-runner /bin/sh /app/test/prepare-schema.sh
	@mkdir -p mounts/schema/chart.local/
	@touch mounts/schema/chart.local/ingressroutetcp_v1alpha1.json
	@docker exec crd-runner /bin/sh /app/main.sh
	@docker exec crd-runner /bin/sh /app/test/verify-test.sh "Only lastest version used"
