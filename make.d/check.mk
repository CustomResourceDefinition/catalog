_check:
	-find build/ephemeral -not -name ".gitignore" -and -not -name ".gitkeep" -type f -delete
	-find build/ephemeral -type d -empty -delete
	touch build/configuration.yaml
	build/bin/catalog update --configuration build/configuration.yaml --output build/ephemeral/schema --definitions build/ephemeral/schema
