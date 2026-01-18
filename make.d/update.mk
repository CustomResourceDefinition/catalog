CONFIGURATION ?= configuration.yaml

update: build/bin/catalog
	df -h
	build/bin/catalog update --configuration $(CONFIGURATION) --output schema --definitions definitions
