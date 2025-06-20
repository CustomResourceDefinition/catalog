generate-comparison: build
	$(COMPOSE_RUN) \
	-v $$(pwd)/schema:/schema \
	-v $$(pwd)/configuration.yaml:/app/configuration.yaml:ro \
	runner make _generate-comparison

_generate-comparison:
	build/bin/catalog generate-status --current schema --remote build/remote/datreeio --out COMPARISON.md
