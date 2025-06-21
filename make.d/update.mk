update: build
	$(COMPOSE_RUN) \
	-v $$(pwd)/schema:/schema \
	-v $$(pwd)/configuration.yaml:/app/configuration.yaml:ro \
	runner make _update

_update:
	build/bin/main
