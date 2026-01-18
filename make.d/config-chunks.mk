CHUNK_SIZE ?= 100

config-chunks:
	cd build && \
		yq -o=json '.' ../configuration.yaml \
		| jq -c ". as \$$arr \
		| [range(0; \$$arr | length; $(CHUNK_SIZE)) | \$$arr[. : .+$(CHUNK_SIZE)]]" \
		| yq -P \
		| yq '.[]' -s '"configuration_" + $$index'
