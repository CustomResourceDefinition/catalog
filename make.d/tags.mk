tags:
	test -d build/remote/kubernetes || exit 1

	(cd build/remote/kubernetes && git for-each-ref --sort=creatordate --format='%(refname:short)|%(creatordate:unix)' refs/tags) | \
		grep -E '^v[0-9]+\.[0-9]+\.[0-9]+\|' | \
		grep -vE '^v0\.[0-9]+\.[0-9]+\|' \
		> build/ephemeral/tags-in

	while IFS='|' read -r TAG TS; do \
		test -z "$$TAG" && continue; \
		\
		test "$$TS" -lt "$(shell git log --reverse --format='%at' | head -n 1)" && continue; \
		test "$$TS" -gt "$(shell date +%s)" && continue; \
		\
		COMMIT="$$(git rev-list -1 --date=unix --date=format:unix --before="$$TS" HEAD || true)"; \
		test -z "$$COMMIT" && continue; \
		\
		echo "$$TAG|$$COMMIT"; \
	done < build/ephemeral/tags-in > build/ephemeral/tags-all

	@cat build/ephemeral/tags-all

	while IFS='|' read -r TAG COMMIT; do \
		test -z "$$TAG" && continue; \
		\
		git rev-parse "$$TAG" >/dev/null 2>&1 && continue; \
		\
		git tag $$TAG $$COMMIT; \
	done < build/ephemeral/tags-all
