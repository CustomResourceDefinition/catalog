update: build/bin/catalog
	build/bin/catalog update \
		--registry registry.yaml \
		--configuration configuration.yaml \
		--output schema \
		--definitions definitions \
		--performance-log build/ephemeral/performance.log
