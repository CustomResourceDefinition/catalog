update: build/bin/catalog
	df -h
	build/bin/catalog update --configuration configuration.yaml --output schema --definitions definitions
