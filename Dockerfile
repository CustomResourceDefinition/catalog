FROM golang:1.25-alpine

RUN apk add --no-cache -q make yq-go helm git jq
RUN mkdir /opt/schemastore && \
        wget -qP /opt/schemastore/ https://json.schemastore.org/dependabot-2.0.json && \
        wget -qP /opt/schemastore/ https://json.schemastore.org/github-workflow.json && \
        wget -qO /opt/schemastore/jsonschema.json https://www.schemastore.org/metaschema-draft-07-unofficial-strict.json
