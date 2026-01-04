update: build
	$(COMPOSE_RUN) \
	-v $$(pwd)/schema:/schema \
	-v $$(pwd)/definitions:/definitions \
	-v $$(pwd)/configuration.yaml:/app/configuration.yaml:ro \
	-v $$(pwd)/build/tmp:/tmp/ephemeral \
	-e GOTMPDIR=/tmp/ephemeral \
	runner make _update

_update:
	build/bin/catalog update --configuration configuration.yaml --output /schema --definitions /definitions
