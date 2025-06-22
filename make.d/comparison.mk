generate-comparison: build
	$(COMPOSE_RUN) \
	-v $$(pwd)/schema:/schema:ro \
	-v $$(pwd)/configuration-ignore.yaml:/ignores.yaml:ro \
	runner make _generate-comparison

_generate-comparison:
	build/bin/catalog generate-status --current /schema --ignore /ignores.yaml --datreeio build/remote/datreeio --out COMPARISON.md
