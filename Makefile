.DEFAULT_GOAL := run
.PHONY: list
list:
	@grep '^[a-zA-Z0-9]' Makefile | cut -d\: -f1

clean:
	@docker rmi -f runner:latest

build: clean
	@docker build -qt runner --build-arg CONFIGURATION=helm-charts.yaml . 2>/dev/null

build-test: clean
	@docker build -qt runner --build-arg CONFIGURATION=helm-charts-test.yaml . 2>/dev/null

--update:
	@docker run \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	runner \
	/bin/sh app/update-helm.sh

--template:
	@docker run \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	-v ./helm/templates:/templates \
	runner \
	/bin/sh app/helm-template.sh

--split:
	@docker run \
	-v ./helm/templates:/templates \
	runner \
	/bin/sh app/split.sh

--convert:
	@docker run \
	-v ./helm/templates:/templates \
	runner \
	/bin/sh app/convert.sh

--arrange:
	@docker run \
	-v ./helm/templates:/templates \
	-v ./schema/:/schema \
	runner \
	/bin/sh app/arrange.sh

run: clean build --update --template --split --convert --arrange
	@docker run \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	runner \
	/bin/sh app/test.sh

test: clean build-test --update --template --split --convert --arrange
	@docker run \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	runner \
	/bin/sh app/test2.sh

shell: clean build
	@docker run -it \
	-v ./helm/cache:/root/.cache/helm \
	-v ./helm/config:/root/.config/helm \
	-v ./helm/local:/root/.local/share/helm \
	runner \
	/bin/sh
