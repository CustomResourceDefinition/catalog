check: build/bin/catalog
	-find build/ephemeral -not -name ".gitignore" -and -not -name ".gitkeep" -type f -delete
	-find build/ephemeral -type d -empty -delete
	touch build/configuration.yaml
	build/bin/catalog update \
		--configuration build/configuration.yaml \
		--output build/ephemeral/schema \
		--definitions build/ephemeral/schema \
		--performance-log build/ephemeral/performance.log

check-sort:
	go test ./... -run '^TestConfigurationIsSortedAndValid$$' -timeout 10s -v
