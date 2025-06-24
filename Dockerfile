FROM alpine:3

RUN apk add --no-cache -q python3 py3-yaml yq-go helm jq wget git kustomize make go check-jsonschema && \
    GOPATH=/usr/local/ go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
