export RUNNER=ghcr.io/customresourcedefinition/catalog-runner:$(shell docker run -v $$(pwd)/Dockerfile:/Dockerfile --rm alpine /bin/sh -c 'md5sum < /Dockerfile | cut -f1 -d" "' 2>/dev/null)
PLATFORM := $(shell docker version --format '{{.Server.Os}}/{{.Server.Arch}}')

# tags are also required in .vscode/settings.json
GO_TAGS = containers_image_openpgp

build:
ifneq ($(strip $(CI)),)
	@echo "$$GITHUB_TOKEN" | docker login ghcr.io -u $$GITHUB_ACTOR --password-stdin
endif

ifeq ($(strip $(GITHUB_REF)),refs/heads/main)
	test -n "$$(docker images -q $(RUNNER))" || \
		docker pull $(RUNNER) || \
		docker build --tag $(RUNNER) --push .
else ifneq ($(strip $(CI)),)
	test -n "$$(docker images -q $(RUNNER))" || \
		docker pull $(RUNNER) || \
		docker build --tag $(RUNNER) .
else
	test -n "$$(docker images -q $(RUNNER))" || \
		docker build --platform=$(PLATFORM) --tag $(RUNNER) .
endif

	test -n "$$(docker images -q $(RUNNER))"
	$(COMPOSE_RUN) runner make _build

_build: _clean
	@mkdir -p build/bin
	go build -o build/bin/catalog -buildvcs=false -tags $(GO_TAGS)
