comparison: build
	$(COMPOSE_RUN) \
	-v $$(pwd)/schema:/schema:ro \
	-v $$(pwd)/configuration-ignore.yaml:/ignores.yaml:ro \
	-v $$(pwd)/build/tmp:/tmp/ephemeral \
	-e GOTMPDIR=/tmp/ephemeral \
	runner make _comparison

_comparison:
	build/bin/catalog compare --current /schema --ignore /ignores.yaml --datreeio build/remote/datreeio --out docs/COMPARISON.md
