.DEFAULT_GOAL := run
.PHONY: list
list:
	@grep '^[a-zA-Z0-9]' Makefile | cut -d\: -f1

clean:
	@docker rmi -f runner:latest

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

--split:
	@docker run \
	-v ./helm/templates:/templates \
	runner \
	/bin/sh app/30-manifests-split.sh

--convert:
	@docker run \
	-v ./helm/templates:/templates \
	runner \
	/bin/sh app/40-manifest-convert.sh

--arrange:
	@docker run \
	-v ./helm/templates:/templates \
	-v ./schema/:/schema \
	runner \
	/bin/sh app/50-schema-arrange.sh

--arrange-test:
	@docker run \
	-v ./helm/templates:/templates \
	-v ./test/schema/:/schema \
	runner \
	/bin/sh app/50-schema-arrange.sh

run: clean build --update --template --split --convert --arrange

test: clean build-test --update --template --split --convert --arrange-test
	@docker run \
	-v ./test/schema/:/schema \
	-v ./test/verified-schema/:/verified-schema \
	runner \
	/bin/sh /test/verify-test.sh

shell: clean build
	@docker run -it \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	runner \
	/bin/sh
