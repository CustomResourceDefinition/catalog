FROM golang:1.24-alpine

RUN apk add --no-cache -q make yq-go helm git
RUN mkdir /opt/schemastore && \
        wget -qP /opt/schemastore/ https://json.schemastore.org/dependabot-2.0.json && \
        wget -qP /opt/schemastore/ https://json.schemastore.org/github-workflow.json
