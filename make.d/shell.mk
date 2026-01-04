shell: build
	$(COMPOSE_RUN) \
	-v $$(pwd)/build/ephemeral/schema:/schema \
	-v $$(pwd)/test/configuration.yaml:/app/configuration.yaml:ro \
	-v $$(pwd)/test:/app/test \
	-v $$(pwd)/build/tmp:/tmp/ephemeral \
	-e GOTMPDIR=/tmp/ephemeral \
	runner /bin/sh
