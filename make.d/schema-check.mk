_schema-check:
	find schema -type f -iname "*.json" -print0 | xargs -0 -n1 sh -ec 'build/bin/catalog verify --schema /opt/schemastore/jsonschema.json --file "$0"'
