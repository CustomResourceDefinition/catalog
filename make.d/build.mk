# tags are also required in .vscode/settings.json
GO_TAGS = containers_image_openpgp

build:
	@mkdir -p build/bin
	go build -o build/bin/catalog -buildvcs=false -tags $(GO_TAGS)
