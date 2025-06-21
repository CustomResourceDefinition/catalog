shell: build
	$(COMPOSE_RUN) \
	-v $$(pwd)/build/ephemeral/schema:/schema \
	-v $$(pwd)/test/configuration.yaml:/app/configuration.yaml:ro \
	-v $$(pwd)/test/.env:/app/.env:ro \
	runner /bin/sh
