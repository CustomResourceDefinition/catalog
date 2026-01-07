schema-check: build/bin/catalog build/schemastore/jsonschema.json
	find schema -type f -iname "*.json" -print0 | xargs -0 -n1 sh -ec 'build/bin/catalog verify --schema build/schemastore/jsonschema.json --file "$0"'

build/schemastore/jsonschema.json:
	@mkdir -p build/schemastore
	$(DOWNLOADER) $@ https://www.schemastore.org/metaschema-draft-07-unofficial-strict.json
