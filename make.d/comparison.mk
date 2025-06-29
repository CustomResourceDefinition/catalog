comparison: build
	$(COMPOSE_RUN) \
	-v $$(pwd)/schema:/schema:ro \
	-v $$(pwd)/configuration-ignore.yaml:/ignores.yaml:ro \
	runner make _comparison

_comparison:
	build/bin/catalog compare --current /schema --ignore /ignores.yaml --datreeio build/remote/datreeio --out COMPARISON.md
