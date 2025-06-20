export RUNNER=ghcr.io/customresourcedefinition/catalog-runner:$(shell docker run -v $$(pwd)/Dockerfile:/Dockerfile --rm alpine /bin/sh -c 'md5sum < /Dockerfile | cut -f1 -d" "' 2>/dev/null)

build:
ifeq ($(strip $(CI)),)
	test -n "$$(docker images -q $(RUNNER))" || \
		docker build --tag $(RUNNER) .
else ifeq ($(strip $(GITHUB_REF)),main)
	test -n "$$(docker images -q $(RUNNER))" || \
		docker pull $(RUNNER) || \
		(echo "$$GITHUB_TOKEN" | docker login ghcr.io -u $$GITHUB_ACTOR --password-stdin) || \
		docker build --tag $(RUNNER) --push .
else
	test -n "$$(docker images -q $(RUNNER))" || \
		docker pull $(RUNNER) || \
		(echo "$$GITHUB_TOKEN" | docker login ghcr.io -u $$GITHUB_ACTOR --password-stdin) || \
		docker build --tag $(RUNNER)$(PUSH) .
endif
	$(COMPOSE_RUN) \
	runner make _build

_build: _clean
	mkdir -p build/bin build/ephemeral build/remote/datreeio
	cat src/*.sh > build/bin/main
	chmod +x build/bin/main
	cd tools && \
		go build -o ../build/bin/catalog -buildvcs=false
