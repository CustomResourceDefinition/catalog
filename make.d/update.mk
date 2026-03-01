update: build/bin/catalog
	df -h
	build/bin/catalog update --registry registry.yaml --configuration configuration.yaml --output schema --definitions definitions
