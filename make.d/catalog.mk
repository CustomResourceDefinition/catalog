generate-status-datreeio: build
	$(COMPOSE_RUN) \
	-v $$(pwd)/schema:/schema \
	-v $$(pwd)/configuration.yaml:/app/configuration.yaml:ro \
	runner make _generate-status-datreeio

_generate-status-datreeio:
	build/bin/catalog generate-status --current /schema --remote /build/remote/datreeio --out MIGRATION.md
