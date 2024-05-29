.DEFAULT_GOAL := run
.PHONY: test

export DOCKER_CLI_HINTS=false

clean:
	@(docker inspect --type=image crd-runner:latest &>/dev/null && (docker rm -f $$(docker container ls -aqf "ancestor=crd-runner:latest") &>/dev/null && docker rmi -f crd-runner:latest &>/dev/null)) || true
	@find helm/ -type d -mindepth 2 -maxdepth 2 -print0 | xargs -0 -I{} sh -c 'rm -r "{}"' || true
	@find test/schema -type d -mindepth 1 -maxdepth 1 -print0 | xargs -0 -I{} sh -c 'rm -r "{}"' || true
	@find helm/config test/repository -type f ! -name .gitkeep ! -name .gitignore -delete

build-image: clean
	@docker build -qt crd-runner . >/dev/null

build: build-image
	@docker container create --name crd-runner \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	-v ./helm/templates:/templates \
	-v ./helm-charts.yaml:/app/configuration.yaml \
	crd-runner >/dev/null
	@docker start crd-runner >/dev/null

build-test: build-image
	@docker container create --name crd-runner \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	-v ./helm/templates:/templates \
	-v ./test/repository/:/repository \
	-v ./test/schema/:/schema \
	-v ./test/verified-schemas/happy-path/:/verified-schema \
	-v ./test/chart/:/chart \
	-v ./test/helm-charts.yaml:/app/configuration.yaml \
	crd-runner >/dev/null
	@docker start crd-runner >/dev/null

shell: clean build
	@docker exec -it crd-runner /bin/sh

run: build
	@docker exec -it crd-runner /bin/sh app/main.sh

test:
	@$(MAKE) test-happy-path
	@$(MAKE) test-only-latest

test-happy-path: build-test
	@docker exec -it crd-runner /bin/sh /test/prepare-chart.sh
	@docker exec -it crd-runner /bin/sh app/main.sh
	@docker exec -it crd-runner /bin/sh /test/verify-test.sh "Happy path works"

test-only-latest: build-test
	@docker exec -it crd-runner /bin/sh /test/prepare-chart.sh
	@mkdir -p test/schema/chart.local/
	@touch test/schema/chart.local/ingressroutetcp_v1alpha1.json
	@docker exec -it crd-runner /bin/sh app/main.sh
	@docker exec -it crd-runner /bin/sh /test/verify-test.sh "Only lastest version used"
