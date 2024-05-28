.DEFAULT_GOAL := run
.PHONY: list
list:
	@grep '^[a-zA-Z0-9]' Makefile | cut -d\: -f1

clean:
	@(docker inspect --type=image runner:latest >/dev/null 2>/dev/null && docker rmi -f runner:latest >/dev/null) || true
	@find helm/templates test/schema -type d -mindepth 1 -maxdepth 1 -print0 | xargs -0 -I{} sh -c 'rm -r "{}"' || true

build: clean
	@docker build -qt runner --build-arg CONFIGURATION=helm-charts.yaml . 2>/dev/null

build-test: clean
	@docker build -qt runner --build-arg CONFIGURATION=test/helm-charts.yaml . 2>/dev/null

--update:
	@docker run \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	runner \
	/bin/sh app/10-helm-update.sh

--template:
	@docker run \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	-v ./helm/templates:/templates \
	runner \
	/bin/sh app/20-helm-template.sh

--deduplicate:
	@docker run \
	-v ./helm/templates:/templates \
	runner \
	/bin/sh app/25-manifests-deduplicate.sh

--split:
	@docker run \
	-v ./helm/templates:/templates \
	runner \
	/bin/sh app/30-manifests-split.sh

--arrange:
	@docker run \
	-v ./helm/templates:/templates \
	-v ./schema/:/schema \
	runner \
	/bin/sh app/40-manifest-arrange.sh

--arrange-test:
	@docker run \
	-v ./helm/templates:/templates \
	-v ./test/schema/:/schema \
	runner \
	/bin/sh app/40-manifest-arrange.sh

--convert:
	@docker run \
	-v ./schema/:/schema \
	runner \
	/bin/sh app/50-manifest-convert.sh

--convert-test:
	@docker run \
	-v ./test/schema/:/schema \
	runner \
	/bin/sh app/50-manifest-convert.sh

run: build --update --template --deduplicate --split --arrange --convert

test: build-test --update --template --deduplicate --split --arrange-test --convert-test
	@docker run \
	-v ./test/schema/:/schema \
	-v ./test/verified-schema/:/verified-schema \
	runner \
	/bin/sh /test/verify-test.sh

shell: build
	@docker run -it \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	runner \
	/bin/sh
